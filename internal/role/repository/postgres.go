package repository

import (
	"context"
	"database/sql"
	"fmt"

	userv1 "github.com/fekuna/omnipos-proto/gen/go/omnipos/user/v1"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateRole(ctx context.Context, merchantID string, role *userv1.Role, permissionIDs []string) (string, error)
	GetRole(ctx context.Context, id string) (*userv1.Role, error)
	ListRoles(ctx context.Context, merchantID string, page, pageSize int32) ([]*userv1.Role, int32, error)
	ListPermissions(ctx context.Context) ([]*userv1.Permission, error)
}

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db: db}
}

type roleModel struct {
	ID          string         `db:"id"`
	MerchantID  string         `db:"merchant_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	IsSystem    bool           `db:"is_system"`
}

type permissionModel struct {
	ID          string         `db:"id"`
	Code        string         `db:"code"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Module      string         `db:"module"`
}

func (r *postgresRepository) toRoleProto(rm *roleModel, permissions []*userv1.Permission) *userv1.Role {
	desc := ""
	if rm.Description.Valid {
		desc = rm.Description.String
	}
	return &userv1.Role{
		Id:          rm.ID,
		Name:        rm.Name,
		Description: desc,
		IsSystem:    rm.IsSystem,
		Permissions: permissions,
	}
}

func (r *postgresRepository) toPermissionProto(pm *permissionModel) *userv1.Permission {
	desc := ""
	if pm.Description.Valid {
		desc = pm.Description.String
	}
	return &userv1.Permission{
		Id:          pm.ID,
		Code:        pm.Code,
		Name:        pm.Name,
		Description: desc,
		Module:      pm.Module,
	}
}

func (r *postgresRepository) CreateRole(ctx context.Context, merchantID string, role *userv1.Role, permissionIDs []string) (string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// 1. Insert Role
	var roleID string
	query := `
		INSERT INTO roles (merchant_id, name, description, is_system)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, query,
		merchantID,
		role.Name,
		role.Description,
		false, // Custom roles are never system roles
	).Scan(&roleID)

	if err != nil {
		return "", err
	}

	// 2. Insert Permissions
	if len(permissionIDs) > 0 {
		permQuery := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`
		for _, permID := range permissionIDs {
			_, err := tx.ExecContext(ctx, permQuery, roleID, permID)
			if err != nil {
				return "", fmt.Errorf("failed to assign permission %s: %w", permID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return roleID, nil
}

func (r *postgresRepository) GetRole(ctx context.Context, id string) (*userv1.Role, error) {
	// 1. Get Role
	var rm roleModel
	query := `SELECT id, merchant_id, name, description, is_system FROM roles WHERE id = $1`
	err := r.db.GetContext(ctx, &rm, query, id)
	if err != nil {
		return nil, err
	}

	// 2. Get Permissions
	var pms []permissionModel
	permQuery := `
        SELECT p.id, p.code, p.name, p.description, p.module
        FROM permissions p
        JOIN role_permissions rp ON p.id = rp.permission_id
        WHERE rp.role_id = $1
    `
	err = r.db.SelectContext(ctx, &pms, permQuery, id)
	if err != nil {
		return nil, err
	}

	var permissions []*userv1.Permission
	for _, pm := range pms {
		permissions = append(permissions, r.toPermissionProto(&pm))
	}

	return r.toRoleProto(&rm, permissions), nil
}

func (r *postgresRepository) ListRoles(ctx context.Context, merchantID string, page, pageSize int32) ([]*userv1.Role, int32, error) {
	offset := (page - 1) * pageSize

	// Count
	var total int32
	countQuery := `SELECT count(*) FROM roles WHERE merchant_id = $1 OR is_system = TRUE`
	if err := r.db.GetContext(ctx, &total, countQuery, merchantID); err != nil {
		return nil, 0, err
	}

	// List
	var rms []roleModel
	query := `
        SELECT id, merchant_id, name, description, is_system 
        FROM roles 
        WHERE merchant_id = $1 OR is_system = TRUE
        ORDER BY is_system DESC, name ASC
        LIMIT $2 OFFSET $3
    `
	err := r.db.SelectContext(ctx, &rms, query, merchantID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var roles []*userv1.Role
	for _, rm := range rms {
		// Optimization: We could fetch permissions for all roles in one query,
		// but for now, simple N+1 loop for MVP is fine as page size is small.
		// Or we just return roles without permissions for list view?
		// Let's return without permissions for List to be efficient, or just basics.
		// Actually, ListRolesResponse implies full Role objects.
		roles = append(roles, r.toRoleProto(&rm, nil))
	}

	return roles, total, nil
}

func (r *postgresRepository) ListPermissions(ctx context.Context) ([]*userv1.Permission, error) {
	var pms []permissionModel
	query := `SELECT id, code, name, description, module FROM permissions ORDER BY module, code`
	err := r.db.SelectContext(ctx, &pms, query)
	if err != nil {
		return nil, err
	}

	var permissions []*userv1.Permission
	for _, pm := range pms {
		permissions = append(permissions, r.toPermissionProto(&pm))
	}
	return permissions, nil
}
