package refreshtoken

import (
	"context"

	"github.com/fekuna/omnipos-user-service/internal/model"
)

// Repository defines the interface for refresh token operations
type Repository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*model.RefreshToken, error)
	RevokeToken(ctx context.Context, token string) error
	RevokeAllByMerchantID(ctx context.Context, merchantID string) error
	DeleteByMerchantID(ctx context.Context, merchantID string) error
	DeleteExpiredTokens(ctx context.Context) error
}
