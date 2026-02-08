package middleware

import (
	"context"
	"strings"

	"github.com/fekuna/omnipos-pkg/logger"
	"github.com/fekuna/omnipos-user-service/internal/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthContextInterceptor extracts auth metadata and puts it in context
type AuthContextInterceptor struct {
	logger logger.ZapLogger
}

// NewAuthContextInterceptor creates a new auth context interceptor
func NewAuthContextInterceptor(log logger.ZapLogger) *AuthContextInterceptor {
	return &AuthContextInterceptor{
		logger: log,
	}
}

// isPublicEndpoint checks if the endpoint requires authentication
func (i *AuthContextInterceptor) isPublicEndpoint(method string) bool {
	// Convention: endpoints containing these keywords don't require auth
	publicPatterns := []string{
		"Login",
		"Register",
		"ForgotPassword",
		"ResetPassword",
		"RefreshToken",
	}

	for _, pattern := range publicPatterns {
		if strings.Contains(method, pattern) {
			return true
		}
	}

	return false
}

// Unary returns a server interceptor that enriches context with user data
func (i *AuthContextInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for public endpoints
		if i.isPublicEndpoint(info.FullMethod) {
			i.logger.Debug("skipping auth for public endpoint", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		// Extract metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			i.logger.Error("no metadata in request")
			return nil, status.Error(codes.Unauthenticated, "missing authentication context")
		}

		// Extract merchant ID (required)
		merchantIDs := md.Get("x-merchant-id")
		if len(merchantIDs) == 0 {
			i.logger.Error("no merchant ID in metadata")
			return nil, status.Error(codes.Unauthenticated, "missing merchant context")
		}

		// Build user context
		userCtx := &auth.UserContext{
			MerchantID: merchantIDs[0],
		}

		// Optional: extract additional fields for future use
		if userIDs := md.Get("x-user-id"); len(userIDs) > 0 {
			userCtx.UserID = userIDs[0]
		}
		if emails := md.Get("x-user-email"); len(emails) > 0 {
			userCtx.Email = emails[0]
		}
		if roles := md.Get("x-user-role"); len(roles) > 0 {
			userCtx.Role = roles[0]
		}

		i.logger.Debug("user context extracted",
			zap.String("merchant_id", userCtx.MerchantID),
			zap.String("user_id", userCtx.UserID),
			zap.String("method", info.FullMethod))

		// Add to context
		ctx = auth.WithUserContext(ctx, userCtx)

		// Continue with enriched context
		return handler(ctx, req)
	}
}
