package dto

type FindOneByAttribute struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Timezone string `json:"timezone"`
}
