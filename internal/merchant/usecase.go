package merchant

import "context"

// MerchantUsecase defines the business logic interface for merchant operations
type MerchantUsecase interface {
	Login(ctx context.Context, phone, pin string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAllDevices(ctx context.Context, merchantID string) error
	RefreshAccessToken(ctx context.Context, refreshToken string) (accessToken string, err error)
}
