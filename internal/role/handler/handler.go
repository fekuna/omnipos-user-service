package handler

import (
	"context"

	"github.com/fekuna/omnipos-pkg/logger"
	userv1 "github.com/fekuna/omnipos-proto/gen/go/omnipos/user/v1"
	"github.com/fekuna/omnipos-user-service/internal/auth"
	"github.com/fekuna/omnipos-user-service/internal/role/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RoleHandler struct {
	userv1.UnimplementedRoleServiceServer
	uc     usecase.Usecase
	logger logger.ZapLogger
}

func NewRoleHandler(uc usecase.Usecase, logger logger.ZapLogger) *RoleHandler {
	return &RoleHandler{
		uc:     uc,
		logger: logger,
	}
}

func (h *RoleHandler) CreateRole(ctx context.Context, req *userv1.CreateRoleRequest) (*userv1.CreateRoleResponse, error) {
	// Extract MerchantID from context (TODO: Implement context extractor)
	// For now, assuming it's available or implemented.
	// If Auth interceptor adds it to context, we need a helper to get it.
	// Let's assume a helper `GetMerchantID(ctx)` exists or we get it from metadata.
	// BUT for now, since I don't have the helper handy, I will rely on the usecase to validate.
	// Wait, the usecase requires it explicitly.
	// I MUST get it from context.

	// TEMPORARY: Hardcoding or assuming header manipulation by Interceptor makes it available.
	// I'll check `internal/middleware` later.
	// Let's assume context has key "merchant_id".

	merchantID := auth.GetMerchantID(ctx)
	if merchantID == "" {
		return nil, status.Error(codes.Unauthenticated, "merchant_id missing from context")
	}

	role, err := h.uc.CreateRole(ctx, merchantID, req)
	if err != nil {
		h.logger.Error("failed to create role", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create role")
	}
	return &userv1.CreateRoleResponse{Role: role}, nil
}

func (h *RoleHandler) GetRole(ctx context.Context, req *userv1.GetRoleRequest) (*userv1.GetRoleResponse, error) {
	role, err := h.uc.GetRole(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to get role", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get role")
	}
	if role == nil {
		return nil, status.Error(codes.NotFound, "role not found")
	}
	return &userv1.GetRoleResponse{Role: role}, nil
}

func (h *RoleHandler) ListRoles(ctx context.Context, req *userv1.ListRolesRequest) (*userv1.ListRolesResponse, error) {
	merchantID := auth.GetMerchantID(ctx)

	if merchantID == "" {
		h.logger.Error("merchant_id missing from context") // Log relevant info
		return nil, status.Error(codes.Unauthenticated, "merchant_id missing from context")
	}

	res, err := h.uc.ListRoles(ctx, merchantID, req)
	if err != nil {
		h.logger.Error("failed to list roles", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list roles")
	}
	return res, nil
}

func (h *RoleHandler) ListPermissions(ctx context.Context, req *emptypb.Empty) (*userv1.ListPermissionsResponse, error) {
	res, err := h.uc.ListPermissions(ctx)
	if err != nil {
		h.logger.Error("failed to list permissions", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list permissions")
	}
	return res, nil
}
