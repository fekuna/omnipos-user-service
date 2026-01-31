package auth

import "context"

// UserContext represents authenticated user information extracted from request metadata
type UserContext struct {
	MerchantID string
	UserID     string // For future use when implementing user-level authentication
	Email      string // For future use
	Role       string // For future use (admin, manager, cashier, etc.)
}

// Context key type for type safety
type contextKey string

const userContextKey contextKey = "user_context"

// WithUserContext adds user context to the context
func WithUserContext(ctx context.Context, userCtx *UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, userCtx)
}

// GetUserContext extracts user context from context
// Returns nil if not found
func GetUserContext(ctx context.Context) *UserContext {
	userCtx, ok := ctx.Value(userContextKey).(*UserContext)
	if !ok {
		return nil
	}
	return userCtx
}

// MustGetUserContext extracts user context or panics
// Use this when you KNOW the context must exist (after auth middleware)
func MustGetUserContext(ctx context.Context) *UserContext {
	userCtx := GetUserContext(ctx)
	if userCtx == nil {
		panic("user context not found - authentication middleware not applied")
	}
	return userCtx
}

// GetMerchantID is a convenience method to get merchant ID from context
// Returns empty string if context is not found
func GetMerchantID(ctx context.Context) string {
	userCtx := GetUserContext(ctx)
	if userCtx == nil {
		return ""
	}
	return userCtx.MerchantID
}
