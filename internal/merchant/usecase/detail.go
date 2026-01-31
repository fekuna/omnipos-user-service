package usecase

import (
	"context"

	"github.com/fekuna/omnipos-user-service/internal/model"
	"go.uber.org/zap"
)

// GetMerchantDetail retrieves merchant details by ID
func (u *merchantUsecase) GetMerchantDetail(ctx context.Context, merchantID string) (*model.Merchant, error) {
	u.logger.Info("getting merchant detail", zap.String("merchant_id", merchantID))

	// Fetch merchant from repository
	merchant, err := u.merchantRepo.FindByID(ctx, merchantID)
	if err != nil {
		u.logger.Error("failed to get merchant", zap.Error(err))
		return nil, err
	}

	if merchant == nil {
		u.logger.Warn("merchant not found", zap.String("merchant_id", merchantID))
		return nil, ErrMerchantNotFound
	}

	u.logger.Info("merchant detail retrieved successfully", zap.String("merchant_id", merchantID))
	return merchant, nil
}
