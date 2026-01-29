package usecase

import (
	"time"

	"github.com/fekuna/omnipos-user-service/internal/helper"
	"github.com/fekuna/omnipos-user-service/internal/merchant/dto"
	"github.com/fekuna/omnipos-user-service/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Login authenticates a merchant and returns access and refresh tokens
func (u *merchantUsecase) Login(phone, pin string) (string, string, error) {
	u.logger.Info("attempting merchant login", zap.String("phone", phone))

	// Find merchant by phone
	merchant, err := u.merchantRepo.FindOneByAttributes(&dto.FindOneByAttribute{
		Phone: phone,
	})
	if err != nil {
		u.logger.Error("failed to find merchant", zap.Error(err))
		return "", "", err
	}

	if merchant == nil {
		u.logger.Warn("merchant not found", zap.String("phone", phone))
		return "", "", ErrMerchantNotFound
	}

	// Verify PIN using bcrypt
	if !helper.ComparePassword(merchant.Pin, pin) {
		u.logger.Warn("invalid PIN attempt", zap.String("merchant_id", merchant.ID))
		return "", "", ErrInvalidCredentials
	}

	// Get JWT helper from context/config
	// Note: We'll pass this via constructor in the actual implementation
	jwtHelper := helper.NewJWTHelper(
		u.jwtSecretKey,
		u.accessTokenExpiry,
		u.refreshTokenExpiry,
	)

	// Generate access token
	accessToken, err := jwtHelper.GenerateAccessToken(merchant.ID)
	if err != nil {
		u.logger.Error("failed to generate access token", zap.Error(err))
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := jwtHelper.GenerateRefreshToken(merchant.ID)
	if err != nil {
		u.logger.Error("failed to generate refresh token", zap.Error(err))
		return "", "", err
	}

	// Store refresh token in database
	refreshTokenModel := &model.RefreshToken{
		BaseModel: model.BaseModel{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
		},
		MerchantID: merchant.ID,
		Token:      refreshToken,
		ExpiresAt:  time.Now().Add(jwtHelper.GetRefreshTokenExpiry()),
	}

	err = u.refreshTokenRepo.Create(refreshTokenModel)
	if err != nil {
		u.logger.Error("failed to store refresh token", zap.Error(err))
		return "", "", err
	}

	u.logger.Info("merchant login successful",
		zap.String("merchant_id", merchant.ID),
		zap.String("merchant_name", merchant.Name),
	)

	return accessToken, refreshToken, nil
}
