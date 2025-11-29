package repository

import (
	"uas/app/model"
	"gorm.io/gorm"
)

type RoleRepository interface {
	FindAll() ([]model.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindAll() ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.Order("created_at ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}