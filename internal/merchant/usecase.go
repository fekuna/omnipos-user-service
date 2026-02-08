package merchant

import (
	"context"

	"github.com/fekuna/omnipos-user-service/internal/model"
)

// MerchantUsecase defines the business logic interface for merchant operations
type MerchantUsecase interface {
	Login(ctx context.Context, phone, pin string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAllDevices(ctx context.Context, merchantID string) error
	RefreshAccessToken(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error)
	GetMerchantDetail(ctx context.Context, merchantID string) (*model.Merchant, error)
	GetMerchantByPhone(ctx context.Context, phone string) (*model.Merchant, error)
}
