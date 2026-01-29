package repository

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/fekuna/omnipos-user-service/internal/merchant/dto"
	"github.com/fekuna/omnipos-user-service/internal/model"
	"github.com/jmoiron/sqlx"
)

type PGRepository struct {
	DB *sqlx.DB
}

func NewPGRepository(db *sqlx.DB) *PGRepository {
	return &PGRepository{DB: db}
}

func (r *PGRepository) FindOneByAttributes(input *dto.FindOneByAttribute) (*model.Merchant, error) {
	var merchant model.Merchant

	conditions := []string{}
	args := map[string]interface{}{}

	if input.ID != "" {
		conditions = append(conditions, "id = :id")
		args["id"] = input.ID
	}
	if input.Name != "" {
		conditions = append(conditions, "name = :name")
		args["name"] = input.Name
	}
	if input.Phone != "" {
		conditions = append(conditions, "phone = :phone")
		args["phone"] = input.Phone
	}
	if input.Timezone != "" {
		conditions = append(conditions, "timezone = :timezone")
		args["timezone"] = input.Timezone
	}

	query := `
		SELECT id, name, phone, pin, timezone, created_at, updated_at
		FROM merchants
	`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " LIMIT 1"

	namedQuery, namedArgs, err := sqlx.Named(query, args)
	if err != nil {
		return nil, err
	}

	namedQuery = r.DB.Rebind(namedQuery)

	err = r.DB.Get(&merchant, namedQuery, namedArgs...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	return &merchant, nil
}
