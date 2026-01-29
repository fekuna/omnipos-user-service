package handler

import (
	"context"
	"errors"

	"github.com/fekuna/omnipos-pkg/logger"
	"github.com/fekuna/omnipos-user-service/internal/merchant/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MerchantHandler implements the UserService gRPC handler
type MerchantHandler struct {
	// Note: We'll uncomment this when proto is generated
	// userv1.UnimplementedUserServiceServer
	merchantUsecase usecase.MerchantUsecase
	logger          logger.ZapLogger
}

// NewMerchantHandler creates a new merchant gRPC handler
func NewMerchantHandler(merchantUsecase usecase.MerchantUsecase, log logger.ZapLogger) *MerchantHandler {
	return &MerchantHandler{
		merchantUsecase: merchantUsecase,
		logger:          log,
	}
}

// LoginMerchant handles the merchant login gRPC request
// Note: This signature will be updated once proto is generated
// For now, we're using simplified types
func (h *MerchantHandler) LoginMerchant(ctx context.Context, req interface{}) (interface{}, error) {
	// Type assertion will be added when proto is generated
	// For now, we define the expected structure
	type LoginRequest struct {
		Phone string
		Pin   string
	}

	type LoginResponse struct {
		AccessToken  string
		RefreshToken string
	}

	// This is a placeholder - actual implementation will use generated proto types
	loginReq, ok := req.(*LoginRequest)
	if !ok {
		h.logger.Error("invalid request type")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	h.logger.Info("processing login request", zap.String("phone", loginReq.Phone))

	// Call use case
	accessToken, refreshToken, err := h.merchantUsecase.Login(loginReq.Phone, loginReq.Pin)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))

		// Map errors to gRPC status codes
		if errors.Is(err, usecase.ErrMerchantNotFound) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid phone or PIN")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	h.logger.Info("login successful", zap.String("phone", loginReq.Phone))

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
