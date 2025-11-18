package auth

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ContextKey is the type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for the user ID
	UserIDKey ContextKey = "user_id"
	// AuthorizationMetadataKey is the metadata key for the authorization header
	AuthorizationMetadataKey = "authorization"
)

// GRPCAuthInterceptor creates a gRPC unary server interceptor for authentication
func GRPCAuthInterceptor(tokenService *TokenService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userID, err := authenticateRequest(ctx, tokenService)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		// Add user ID to context
		ctx = context.WithValue(ctx, UserIDKey, userID)

		return handler(ctx, req)
	}
}

// GRPCAuthStreamInterceptor creates a gRPC stream server interceptor for authentication
func GRPCAuthStreamInterceptor(tokenService *TokenService) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		userID, err := authenticateRequest(ss.Context(), tokenService)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		// Wrap the stream to include the user ID in the context
		wrapped := &authenticatedStream{
			ServerStream: ss,
			ctx:          context.WithValue(ss.Context(), UserIDKey, userID),
		}

		return handler(srv, wrapped)
	}
}

// authenticatedStream wraps a grpc.ServerStream with an authenticated context
type authenticatedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the authenticated context
func (s *authenticatedStream) Context() context.Context {
	return s.ctx
}

// authenticateRequest extracts and validates the API token from the gRPC metadata
func authenticateRequest(ctx context.Context, tokenService *TokenService) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, fmt.Errorf("missing metadata")
	}

	tokens := md.Get(AuthorizationMetadataKey)
	if len(tokens) == 0 {
		return uuid.Nil, fmt.Errorf("missing authorization token")
	}

	token := tokens[0]
	if token == "" {
		return uuid.Nil, fmt.Errorf("empty authorization token")
	}

	// Extract token (support both "Bearer <token>" and plain token)
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	userID, err := tokenService.ValidateToken(ctx, token)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// GetUserIDFromContext extracts the user ID from the context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// HTTPAuthMiddleware creates an HTTP middleware for authentication (for REST API)
func HTTPAuthMiddleware(tokenService *TokenService) func(next interface{}) interface{} {
	return func(next interface{}) interface{} {
		// This is a placeholder - actual implementation depends on the HTTP framework used
		// For now, return a generic function
		return func(ctx context.Context) (context.Context, error) {
			// Extract token from HTTP headers (implementation depends on framework)
			// For demonstration purposes, we'll just return the context
			return ctx, nil
		}
	}
}

// NewTokenService creates a new token service (convenience function)
func NewAuthTokenService(db *sql.DB) *TokenService {
	return NewTokenService(db)
}
