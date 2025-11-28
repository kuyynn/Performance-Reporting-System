package model

import "time"

type AchievementStatus string

const (
	AchievementStatusDraft     AchievementStatus = "draft"
	AchievementStatusSubmitted AchievementStatus = "submitted"
	AchievementStatusVerified  AchievementStatus = "verified"
	AchievementStatusRejected  AchievementStatus = "rejected"
)

type AchievementReference struct {
	ID                 string            `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	StudentID          string            `gorm:"type:uuid;not null" json:"student_id"`
	Student            Student           `gorm:"foreignKey:StudentID" json:"student"`
	MongoAchievementID string            `gorm:"size:24;not null" json:"mongo_achievement_id"`
	Status             AchievementStatus `gorm:"type:achievement_status;not null" json:"status"`
	SubmittedAt        *time.Time        `json:"submitted_at,omitempty"`
	VerifiedAt         *time.Time        `json:"verified_at,omitempty"`
	VerifiedBy         *string           `gorm:"type:uuid" json:"verified_by,omitempty"`
	RejectionNote      *string           `json:"rejection_note,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}