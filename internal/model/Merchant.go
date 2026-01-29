package model

type Merchant struct {
	BaseModel
	Name     string `db:"name"`
	Phone    string `db:"phone"`
	Timezone string `db:"timezone"`
	Pin      string `db:"pin"`
}
