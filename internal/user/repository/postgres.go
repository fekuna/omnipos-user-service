package repository

import (
	"context"
	"database/sql"
	"time"

	userv1 "github.com/fekuna/omnipos-proto/proto/user/v1"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *userv1.User, passwordHash string) (string, error)
	GetUser(ctx context.Context, id string) (*userv1.User, error)
	GetUserByUsername(ctx context.Context, merchantID, username string) (*userv1.User, string, error) // Returns User + PasswordHash
	ListUsers(ctx context.Context, merchantID string, page, pageSize int32) ([]*userv1.User, int32, error)
	UpdateUser(ctx context.Context, user *userv1.User) error
	UpdateUserPassword(ctx context.Context, id, passwordHash string) error
	DeleteUser(ctx context.Context, id string) error
}

type postgresUserRepository struct {
	db *sqlx.DB
}

func NewPostgresUserRepository(db *sqlx.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

type userModel struct {
	ID           string         `db:"id"`
	MerchantID   string         `db:"merchant_id"`
	Username     string         `db:"username"`
	Email        sql.NullString `db:"email"`
	Phone        sql.NullString `db:"phone"`
	FullName     string         `db:"full_name"`
	PasswordHash string         `db:"password_hash"`
	RoleID       sql.NullString `db:"role_id"`
	Status       string         `db:"status"`
	LastLoginAt  sql.NullTime   `db:"last_login_at"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	Timezone     sql.NullString `db:"timezone"`

	// Joined fields
	RoleName sql.NullString `db:"role_name"`
}

func (r *postgresUserRepository) toProto(m *userModel) *userv1.User {
	var lastLogin *timestamppb.Timestamp
	if m.LastLoginAt.Valid {
		lastLogin = timestamppb.New(m.LastLoginAt.Time)
	}

	u := &userv1.User{
		Id:          m.ID,
		MerchantId:  m.MerchantID,
		Username:    m.Username,
		Email:       m.Email.String,
		FullName:    m.FullName,
		RoleId:      m.RoleID.String,
		Status:      m.Status,
		LastLoginAt: lastLogin,
		CreatedAt:   timestamppb.New(m.CreatedAt),
		UpdatedAt:   timestamppb.New(m.UpdatedAt),
	}

	if m.RoleID.Valid && m.RoleName.Valid {
		u.Role = &userv1.Role{
			Id:   m.RoleID.String,
			Name: m.RoleName.String,
		}
	}

	return u
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *userv1.User, passwordHash string) (string, error) {
	query := `
		INSERT INTO users (merchant_id, username, email, full_name, password_hash, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`

	var roleID interface{}
	if user.RoleId != "" {
		roleID = user.RoleId
	}

	var id string
	err := r.db.QueryRowContext(ctx, query,
		user.MerchantId,
		user.Username,
		user.Email,
		user.FullName,
		passwordHash,
		roleID,
		"active", // Default status
	).Scan(&id)

	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *postgresUserRepository) GetUser(ctx context.Context, id string) (*userv1.User, error) {
	var m userModel
	// Join with roles to get role name
	query := `
        SELECT u.*, r.name as role_name
        FROM users u
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE u.id = $1
    `
	err := r.db.GetContext(ctx, &m, query, id)
	if err != nil {
		return nil, err
	}
	user := r.toProto(&m)

	if user.Role != nil {
		perms, err := r.getPermissionsForRole(ctx, user.Role.Id)
		if err != nil {
			return nil, err
		}
		user.Role.Permissions = perms
	}
	return user, nil
}

func (r *postgresUserRepository) GetUserByUsername(ctx context.Context, merchantID, username string) (*userv1.User, string, error) {
	var m userModel
	query := `
        SELECT u.*, r.name as role_name
        FROM users u
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE u.merchant_id = $1 AND (u.username = $2 OR u.email = $2)
    `
	err := r.db.GetContext(ctx, &m, query, merchantID, username)
	if err != nil {
		return nil, "", err
	}
	user := r.toProto(&m)

	if user.Role != nil {
		perms, err := r.getPermissionsForRole(ctx, user.Role.Id)
		if err != nil {
			return nil, "", err
		}
		user.Role.Permissions = perms
	}

	return user, m.PasswordHash, nil
}

func (r *postgresUserRepository) ListUsers(ctx context.Context, merchantID string, page, pageSize int32) ([]*userv1.User, int32, error) {
	offset := (page - 1) * pageSize
	var total int32

	countQuery := `SELECT count(*) FROM users WHERE merchant_id = $1`
	if err := r.db.GetContext(ctx, &total, countQuery, merchantID); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*userv1.User{}, 0, nil
	}

	var models []userModel
	query := `
		SELECT u.*, r.name as role_name
		FROM users u
        LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.merchant_id = $1
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &models, query, merchantID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var users []*userv1.User
	for _, m := range models {
		users = append(users, r.toProto(&m))
	}

	return users, total, nil
}

func (r *postgresUserRepository) UpdateUser(ctx context.Context, user *userv1.User) error {
	query := `
		UPDATE users 
		SET full_name = $1, role_id = $2, status = $3, updated_at = NOW()
		WHERE id = $4
	`

	var roleID interface{}
	if user.RoleId != "" {
		roleID = user.RoleId
	}

	_, err := r.db.ExecContext(ctx, query,
		user.FullName,
		roleID,
		user.Status,
		user.Id,
	)
	return err
}

func (r *postgresUserRepository) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, passwordHash, id)
	return err
}

func (r *postgresUserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

type permissionModel struct {
	ID          string `db:"id"`
	Code        string `db:"code"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Module      string `db:"module"`
}

func (r *postgresUserRepository) getPermissionsForRole(ctx context.Context, roleID string) ([]*userv1.Permission, error) {
	query := `
		SELECT p.id, p.code, p.name, p.description, p.module
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`
	var models []permissionModel
	err := r.db.SelectContext(ctx, &models, query, roleID)
	if err != nil {
		return nil, err
	}

	var perms []*userv1.Permission
	for _, m := range models {
		perms = append(perms, &userv1.Permission{
			Id:          m.ID,
			Code:        m.Code,
			Name:        m.Name,
			Description: m.Description,
			Module:      m.Module,
		})
	}
	return perms, nil
}
