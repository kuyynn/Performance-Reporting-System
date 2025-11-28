package model

import "time"

type Lecturer struct {
	ID         string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID     string    `gorm:"type:uuid;not null" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"user"`
	LecturerID string    `gorm:"size:20;unique;not null" json:"lecturer_id"`
	Department string    `gorm:"size:100" json:"department"`
	CreatedAt  time.Time `json:"created_at"`
}
