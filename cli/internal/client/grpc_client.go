package client

import (
	"context"
	"fmt"
	"io"
	"time"

	pb "github.com/aliancn/swiftlog/cli/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Client wraps a gRPC connection to the SwiftLog ingestor service
type Client struct {
	conn   *grpc.ClientConn
	client pb.LogStreamerClient
	token  string
}

// Config holds client configuration
type Config struct {
	ServerAddr string
	Token      string
}

// NewClient creates a new gRPC client
func NewClient(cfg *Config) (*Client, error) {
	// Create gRPC connection
	conn, err := grpc.NewClient(
		cfg.ServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewLogStreamerClient(conn),
		token:  cfg.Token,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// StreamSession represents an active log streaming session
type StreamSession struct {
	stream pb.LogStreamer_StreamLogClient
	runID  string
}

// StartStream initiates a new log streaming session
func (c *Client) StartStream(ctx context.Context, projectName, groupName string) (*StreamSession, error) {
	// Add authentication metadata
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + c.token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Create bidirectional stream
	stream, err := c.client.StreamLog(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	// Send metadata as first message
	err = stream.Send(&pb.StreamLogRequest{
		Event: &pb.StreamLogRequest_Metadata{
			Metadata: &pb.StreamMetadata{
				ProjectName: projectName,
				GroupName:   groupName,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send metadata: %w", err)
	}

	// Wait for StreamStarted response
	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive started response: %w", err)
	}

	started := resp.GetStarted()
	if started == nil {
		if errMsg := resp.GetError(); errMsg != "" {
			return nil, fmt.Errorf("server error: %s", errMsg)
		}
		return nil, fmt.Errorf("unexpected response from server")
	}

	return &StreamSession{
		stream: stream,
		runID:  started.RunId,
	}, nil
}

// GetRunID returns the run ID for this session
func (s *StreamSession) GetRunID() string {
	return s.runID
}

// SendLogLine sends a log line to the server
func (s *StreamSession) SendLogLine(isStderr bool, content string) error {
	level := pb.LogLine_STDOUT
	if isStderr {
		level = pb.LogLine_STDERR
	}

	return s.stream.Send(&pb.StreamLogRequest{
		Event: &pb.StreamLogRequest_Line{
			Line: &pb.LogLine{
				Timestamp: timestamppb.New(time.Now()),
				Level:     level,
				Content:   content,
			},
		},
	})
}

// SendCompletion sends the completion message with exit code
func (s *StreamSession) SendCompletion(exitCode int32) error {
	return s.stream.Send(&pb.StreamLogRequest{
		Event: &pb.StreamLogRequest_Completion{
			Completion: &pb.StreamCompletion{
				ExitCode: exitCode,
			},
		},
	})
}

// Close closes the stream
func (s *StreamSession) Close() error {
	return s.stream.CloseSend()
}

// WaitForCompletion waits for any final messages from the server
func (s *StreamSession) WaitForCompletion() error {
	for {
		_, err := s.stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
