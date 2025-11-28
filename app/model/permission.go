package model

type Permission struct {
	ID          string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name        string `gorm:"size:100;unique;not null" json:"name"`
	Resource    string `gorm:"size:50;not null" json:"resource"`
	Action      string `gorm:"size:50;not null" json:"action"`
	Description string `json:"description"`
}
