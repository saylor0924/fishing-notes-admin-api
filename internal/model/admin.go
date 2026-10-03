package model

import (
	"database/sql"
	"time"
)

type AdminUser struct {
	ID           int64        `db:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Username     string       `db:"username" gorm:"column:username;size:64;not null;uniqueIndex:uk_admin_users_username"`
	PasswordHash string       `db:"password_hash" gorm:"column:password_hash;size:255;not null"`
	Nickname     string       `db:"nickname" gorm:"column:nickname;size:64;not null;default:''"`
	Status       int64        `db:"status" gorm:"column:status;not null;default:1"`
	IsSuper      int64        `db:"is_super" gorm:"column:is_super;not null;default:0"`
	LastLoginAt  sql.NullTime `db:"last_login_at" gorm:"column:last_login_at"`
	CreatedAt    time.Time    `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt    time.Time    `gorm:"column:updated_at;not null;autoUpdateTime"`
}

type AdminRole struct {
	ID          int64  `db:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Code        string `db:"code" gorm:"column:code;size:64;not null;uniqueIndex:uk_admin_roles_code"`
	Name        string `db:"name" gorm:"column:name;size:64;not null"`
	Description string `db:"description" gorm:"column:description;size:255;not null;default:''"`
	Status      int64  `db:"status" gorm:"column:status;not null;default:1"`
	Sort        int64  `db:"sort" gorm:"column:sort;not null;default:0"`
}

type AdminPermission struct {
	ID          int64  `db:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ParentID    int64  `db:"parent_id" gorm:"column:parent_id;not null;default:0"`
	Code        string `db:"code" gorm:"column:code;size:128;not null;uniqueIndex:uk_admin_permissions_code"`
	Name        string `db:"name" gorm:"column:name;size:128;not null"`
	Type        string `db:"type" gorm:"column:type;size:32;not null"`
	Path        string `db:"path" gorm:"column:path;size:255;not null;default:''"`
	Method      string `db:"method" gorm:"column:method;size:16;not null;default:''"`
	Component   string `db:"component" gorm:"column:component;size:255;not null;default:''"`
	Icon        string `db:"icon" gorm:"column:icon;size:64;not null;default:''"`
	Description string `db:"description" gorm:"column:description;size:255;not null;default:''"`
	Status      int64  `db:"status" gorm:"column:status;not null;default:1"`
	Sort        int64  `db:"sort" gorm:"column:sort;not null;default:0"`
}

type RolePermissionPair struct {
	RoleID         int64  `db:"role_id"`
	PermissionCode string `db:"permission_code"`
}

type AdminUserRole struct {
	UserID      int64  `db:"user_id"`
	RoleID      int64  `db:"role_id"`
	Code        string `db:"code"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Status      int64  `db:"status"`
	Sort        int64  `db:"sort"`
}
