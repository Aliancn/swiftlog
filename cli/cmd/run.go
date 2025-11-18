package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/aliancn/swiftlog/cli/internal/client"
	"github.com/aliancn/swiftlog/cli/internal/config"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- <command> [args...]",
	Short: "Execute a command and stream its logs to SwiftLog",
	Long: `Execute a command and capture its stdout/stderr output, streaming
the logs in real-time to the SwiftLog platform for storage and analysis.

Example:
  swiftlog run --project myapp --group build -- ./build.sh
  swiftlog run --project data -- python train_model.py`,
	RunE: runRun,
}

var (
	projectName string
	groupName   string
)

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVar(&projectName, "project", "", "Project name (default: \"default\")")
	runCmd.Flags().StringVar(&groupName, "group", "", "Group name (default: \"default\")")
}

func runRun(cmd *cobra.Command, args []string) error {
	// Find the "--" separator
	dashDashIndex := -1
	for i, arg := range os.Args {
		if arg == "--" {
			dashDashIndex = i
			break
		}
	}

	if dashDashIndex == -1 || dashDashIndex == len(os.Args)-1 {
		return fmt.Errorf("usage: swiftlog run [flags] -- <command> [args...]")
	}

	// Extract the command to execute
	commandArgs := os.Args[dashDashIndex+1:]
	if len(commandArgs) == 0 {
		return fmt.Errorf("no command specified after --")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w (run 'swiftlog config set --token YOUR_TOKEN' first)", err)
	}

	// Override with flags if provided
	if token, _ := cmd.Flags().GetString("token"); token != "" {
		cfg.Token = token
	}
	if server, _ := cmd.Flags().GetString("server"); server != "" {
		cfg.ServerAddr = server
	}

	if cfg.Token == "" {
		return fmt.Errorf("no API token configured. Run 'swiftlog config set --token YOUR_TOKEN' first")
	}

	// Set defaults
	if projectName == "" {
		projectName = "default"
	}
	if groupName == "" {
		groupName = "default"
	}

	// Create gRPC client
	grpcClient, err := client.NewClient(&client.Config{
		ServerAddr: cfg.ServerAddr,
		Token:      cfg.Token,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer grpcClient.Close()

	// Start streaming session
	ctx := context.Background()
	session, err := grpcClient.StartStream(ctx, projectName, groupName)
	if err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}
	defer session.Close()

	fmt.Printf("📝 Streaming logs to SwiftLog (Run ID: %s)\n", session.GetRunID())
	fmt.Printf("Project: %s, Group: %s\n", projectName, groupName)
	fmt.Println(strings.Repeat("-", 60))

	// Execute the command
	exitCode, err := executeCommand(commandArgs, session)
	if err != nil {
		return err
	}

	// Send completion message
	if err := session.SendCompletion(int32(exitCode)); err != nil {
		return fmt.Errorf("failed to send completion: %w", err)
	}

	// Wait for server acknowledgment
	session.WaitForCompletion()

	// Print summary
	fmt.Println(strings.Repeat("-", 60))
	if exitCode == 0 {
		fmt.Printf("✅ Run completed (Exit Code: %d)\n", exitCode)
	} else {
		fmt.Printf("❌ Run failed (Exit Code: %d)\n", exitCode)
	}
	fmt.Printf("Logs saved to Project[%s], Group[%s]\n", projectName, groupName)
	fmt.Printf("Run ID: %s\n", session.GetRunID())

	// Exit with the same code as the command
	os.Exit(exitCode)
	return nil
}

func executeCommand(args []string, session *client.StreamSession) (int, error) {
	// Create command
	command := exec.Command(args[0], args[1:]...)

	// Create pipes for stdout and stderr
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return 1, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := command.StderrPipe()
	if err != nil {
		return 1, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := command.Start(); err != nil {
		return 1, fmt.Errorf("failed to start command: %w", err)
	}

	// Channel to signal completion
	done := make(chan struct{})

	// Stream stdout
	go streamOutput(stdoutPipe, false, session, done)

	// Stream stderr
	go streamOutput(stderrPipe, true, session, done)

	// Wait for command to complete
	err = command.Wait()

	// Wait for both streams to finish
	<-done
	<-done

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			} else {
				exitCode = 1
			}
		} else {
			return 1, fmt.Errorf("command execution error: %w", err)
		}
	}

	return exitCode, nil
}

func streamOutput(pipe io.ReadCloser, isStderr bool, session *client.StreamSession, done chan struct{}) {
	defer func() { done <- struct{}{} }()

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()

		// Print to terminal
		if isStderr {
			fmt.Fprintln(os.Stderr, line)
		} else {
			fmt.Println(line)
		}

		// Send to SwiftLog server
		if err := session.SendLogLine(isStderr, line); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to send log line: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		// Ignore "file already closed" errors which occur when the command finishes
		if err != io.ErrClosedPipe && !strings.Contains(err.Error(), "file already closed") {
			fmt.Fprintf(os.Stderr, "Error reading output: %v\n", err)
		}
	}
}
