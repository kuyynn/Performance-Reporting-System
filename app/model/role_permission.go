package model

type RolePermission struct {
	RoleID       string     `gorm:"type:uuid;primaryKey" json:"role_id"`
	PermissionID string     `gorm:"type:uuid;primaryKey" json:"permission_id"`
	Role         Role       `gorm:"foreignKey:RoleID" json:"-"`
	Permission   Permission `gorm:"foreignKey:PermissionID" json:"-"`
}