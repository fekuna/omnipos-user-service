package usecase

import (
	"context"
	"time"

	"github.com/fekuna/omnipos-user-service/internal/helper"
	"github.com/fekuna/omnipos-user-service/internal/model"
	"github.com/google/uuid"
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
func (u *merchantUsecase) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	u.logger.Info("attempting to refresh access token")

	// Find the refresh token in database
	token, err := u.refreshTokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		u.logger.Error("failed to find refresh token", zap.Error(err))
		return "", "", err
	}

	if token == nil {
		u.logger.Warn("refresh token not found, expired, or revoked")
		return "", "", ErrInvalidCredentials
	}

	// If we reach here, token is valid and not revoked (checked in repository)

	// 1. Revoke the OLD refresh token (Rotation)
	// We do this BEFORE generating the new one to ensure single-use
	if err := u.refreshTokenRepo.RevokeToken(ctx, refreshToken); err != nil {
		u.logger.Error("failed to revoke old refresh token during rotation", zap.Error(err))
		// Should we fail? Yes, security first.
		return "", "", err
	}

	// 2. Generate NEW Access Token
	jwtHelper := helper.NewJWTHelper(
		u.jwtSecretKey,
		u.accessTokenExpiry,
		u.refreshTokenExpiry,
	)

	newAccessToken, err := jwtHelper.GenerateAccessToken(token.MerchantID)
	if err != nil {
		u.logger.Error("failed to generate new access token", zap.Error(err))
		return "", "", err
	}

	// 3. Generate NEW Refresh Token (String)
	newRefreshTokenString, err := jwtHelper.GenerateRefreshToken(token.MerchantID)
	if err != nil {
		u.logger.Error("failed to generate new refresh token string", zap.Error(err))
		return "", "", err
	}

	// 4. Save NEW Refresh Token to Database
	now := time.Now()
	newRefreshToken := &model.RefreshToken{
		BaseModel: model.BaseModel{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		MerchantID: token.MerchantID,
		Token:      newRefreshTokenString,
		IsRevoked:  false,
		ExpiresAt:  now.Add(u.refreshTokenExpiry),
	}

	if err := u.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		u.logger.Error("failed to save new refresh token", zap.Error(err))
		return "", "", err
	}

	u.logger.Info("tokens rotated (refresh) successfully", zap.String("merchant_id", token.MerchantID))
	return newAccessToken, newRefreshTokenString, nil
}
