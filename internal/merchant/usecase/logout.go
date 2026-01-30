package usecase

import (
	"context"

	"github.com/fekuna/omnipos-user-service/internal/helper"
	"go.uber.org/zap"
)

// Logout revokes a specific refresh token (single device logout)
func (u *merchantUsecase) Logout(ctx context.Context, refreshToken string) error {
	u.logger.Info("merchant logout - revoking token")

	// Revoke the specific refresh token
	err := u.refreshTokenRepo.RevokeToken(ctx, refreshToken)
	if err != nil {
		u.logger.Error("failed to revoke refresh token", zap.Error(err))
		return err
	}

	u.logger.Info("refresh token revoked successfully")
	return nil
}

// LogoutAllDevices revokes all refresh tokens for a merchant (all devices logout)
func (u *merchantUsecase) LogoutAllDevices(ctx context.Context, merchantID string) error {
	u.logger.Info("merchant logout all devices", zap.String("merchant_id", merchantID))

	// Revoke all refresh tokens for this merchant
	err := u.refreshTokenRepo.RevokeAllByMerchantID(ctx, merchantID)
	if err != nil {
		u.logger.Error("failed to revoke all tokens", zap.Error(err))
		return err
	}

	u.logger.Info("all refresh tokens revoked successfully", zap.String("merchant_id", merchantID))
	return nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (u *merchantUsecase) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	u.logger.Info("attempting to refresh access token")

	// Find the refresh token in database
	token, err := u.refreshTokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		u.logger.Error("failed to find refresh token", zap.Error(err))
		return "", err
	}

	if token == nil {
		u.logger.Warn("refresh token not found, expired, or revoked")
		return "", ErrInvalidCredentials
	}

	// If we reach here, token is valid and not revoked (checked in repository)
	// Generate new access token
	jwtHelper := helper.NewJWTHelper(
		u.jwtSecretKey,
		u.accessTokenExpiry,
		u.refreshTokenExpiry,
	)

	accessToken, err := jwtHelper.GenerateAccessToken(token.MerchantID)
	if err != nil {
		u.logger.Error("failed to generate new access token", zap.Error(err))
		return "", err
	}

	u.logger.Info("access token refreshed successfully", zap.String("merchant_id", token.MerchantID))
	return accessToken, nil
}
