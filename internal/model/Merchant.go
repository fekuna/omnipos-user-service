package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Merchant struct {
	BaseModel
	Name         string       `db:"name"`
	Phone        string       `db:"phone"`
	Timezone     string       `db:"timezone"`
	Pin          string       `db:"pin"`
	FeatureFlags FeatureFlags `db:"feature_flags"`
}

type FeatureFlags struct {
	UserManagement bool `json:"user_management"`
}

func (f *FeatureFlags) Scan(value interface{}) error {
	if value == nil {
		*f = FeatureFlags{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &f)
}

func (f FeatureFlags) Value() (driver.Value, error) {
	return json.Marshal(f)
}
