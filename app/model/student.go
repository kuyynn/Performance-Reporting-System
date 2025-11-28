package model

import "time"

type Student struct {
	ID           string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID       string    `gorm:"type:uuid;not null" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"user"`
	StudentID    string    `gorm:"size:20;unique;not null" json:"student_id"`
	ProgramStudy string    `gorm:"size:100" json:"program_study"`
	AcademicYear string    `gorm:"size:10" json:"academic_year"`
	AdvisorID    string    `gorm:"type:uuid" json:"advisor_id"`
	CreatedAt    time.Time `json:"created_at"`
}