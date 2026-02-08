package usecase

import (
	"context"
	"fmt"

	userv1 "github.com/fekuna/omnipos-proto/gen/go/omnipos/user/v1"
	"github.com/fekuna/omnipos-user-service/internal/role/repository"
)

type Usecase interface {
	CreateRole(ctx context.Context, merchantID string, req *userv1.CreateRoleRequest) (*userv1.Role, error)
	GetRole(ctx context.Context, id string) (*userv1.Role, error)
	ListRoles(ctx context.Context, merchantID string, req *userv1.ListRolesRequest) (*userv1.ListRolesResponse, error)
	ListPermissions(ctx context.Context) (*userv1.ListPermissionsResponse, error)
}

type roleUsecase struct {
	repo repository.Repository
}

func NewRoleUsecase(repo repository.Repository) Usecase {
	return &roleUsecase{repo: repo}
}

func (uc *roleUsecase) CreateRole(ctx context.Context, merchantID string, req *userv1.CreateRoleRequest) (*userv1.Role, error) {
	if merchantID == "" {
		return nil, fmt.Errorf("merchantID is required")
	}

	role := &userv1.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	id, err := uc.repo.CreateRole(ctx, merchantID, role, req.PermissionIds)
	if err != nil {
		return nil, err
	}

	// Return the created role
	// Optimization: We could return the struct constructed with ID, instead of fetching again.
	// But fetching ensures we return exactly what's in DB (including default values if any).
	return uc.GetRole(ctx, id)
}

func (uc *roleUsecase) GetRole(ctx context.Context, id string) (*userv1.Role, error) {
	return uc.repo.GetRole(ctx, id)
}

func (uc *roleUsecase) ListRoles(ctx context.Context, merchantID string, req *userv1.ListRolesRequest) (*userv1.ListRolesResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	roles, total, err := uc.repo.ListRoles(ctx, merchantID, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &userv1.ListRolesResponse{
		Roles: roles,
		Total: total,
	}, nil
}

func (uc *roleUsecase) ListPermissions(ctx context.Context) (*userv1.ListPermissionsResponse, error) {
	perms, err := uc.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return &userv1.ListPermissionsResponse{Permissions: perms}, nil
}
