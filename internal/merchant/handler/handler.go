package handler

import (
	"context"
	"errors"

	"github.com/fekuna/omnipos-pkg/logger"
	userv1 "github.com/fekuna/omnipos-proto/proto/user/v1"
	"github.com/fekuna/omnipos-user-service/internal/merchant/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MerchantHandler struct {
	userv1.UnimplementedMerchantServiceServer

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

func (h *MerchantHandler) LoginMerchant(ctx context.Context, req *userv1.LoginMerchantRequest) (*userv1.LoginMerchantResponse, error) {
	type LoginRequest struct {
		Phone string
		Pin   string
	}

	type LoginResponse struct {
		AccessToken  string
		RefreshToken string
	}

	h.logger.Info("processing login request", zap.String("phone", req.Phone))

	// Call use case
	accessToken, refreshToken, err := h.merchantUsecase.Login(req.Phone, req.Pin)
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

	h.logger.Info("login successful", zap.String("phone", req.Phone))

	return &userv1.LoginMerchantResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// LogoutMerchant handles the merchant logout gRPC request (single device)
// Note: This signature will be updated once proto is generated
func (h *MerchantHandler) LogoutMerchant(ctx context.Context, req interface{}) (interface{}, error) {
	type LogoutRequest struct {
		RefreshToken string
	}

	type LogoutResponse struct {
		Success bool
		Message string
	}

	logoutReq, ok := req.(*LogoutRequest)
	if !ok {
		h.logger.Error("invalid request type")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	h.logger.Info("processing logout request")

	// Call use case
	err := h.merchantUsecase.Logout(logoutReq.RefreshToken)
	if err != nil {
		h.logger.Error("logout failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	h.logger.Info("logout successful")

	return &LogoutResponse{
		Success: true,
		Message: "logged out successfully",
	}, nil
}

// LogoutAllDevices handles logout from all devices for a merchant
// Note: This signature will be updated once proto is generated
func (h *MerchantHandler) LogoutAllDevices(ctx context.Context, req interface{}) (interface{}, error) {
	type LogoutAllRequest struct {
		MerchantID string
	}

	type LogoutAllResponse struct {
		Success bool
		Message string
	}

	logoutAllReq, ok := req.(*LogoutAllRequest)
	if !ok {
		h.logger.Error("invalid request type")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	h.logger.Info("processing logout all devices request", zap.String("merchant_id", logoutAllReq.MerchantID))

	// Call use case
	err := h.merchantUsecase.LogoutAllDevices(logoutAllReq.MerchantID)
	if err != nil {
		h.logger.Error("logout all devices failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to logout from all devices")
	}

	h.logger.Info("logout all devices successful", zap.String("merchant_id", logoutAllReq.MerchantID))

	return &LogoutAllResponse{
		Success: true,
		Message: "logged out from all devices successfully",
	}, nil
}

// RefreshToken handles refresh token rotation to get a new access token
// Note: This signature will be updated once proto is generated
func (h *MerchantHandler) RefreshToken(ctx context.Context, req interface{}) (interface{}, error) {
	type RefreshTokenRequest struct {
		RefreshToken string
	}

	type RefreshTokenResponse struct {
		AccessToken string
	}

	refreshReq, ok := req.(*RefreshTokenRequest)
	if !ok {
		h.logger.Error("invalid request type")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	h.logger.Info("processing refresh token request")

	// Call use case
	accessToken, err := h.merchantUsecase.RefreshAccessToken(refreshReq.RefreshToken)
	if err != nil {
		h.logger.Error("token refresh failed", zap.Error(err))

		// Map errors to gRPC status codes
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid or revoked refresh token")
		}

		return nil, status.Error(codes.Internal, "failed to refresh token")
	}

	h.logger.Info("token refresh successful")

	return &RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}
