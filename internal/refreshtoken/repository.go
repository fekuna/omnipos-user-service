package refreshtoken

import "github.com/fekuna/omnipos-user-service/internal/model"

// Repository defines the interface for refresh token operations
type Repository interface {
	Create(token *model.RefreshToken) error
	FindByToken(token string) (*model.RefreshToken, error)
	RevokeToken(token string) error
	RevokeAllByMerchantID(merchantID string) error
	DeleteByMerchantID(merchantID string) error
	DeleteExpiredTokens() error
}
