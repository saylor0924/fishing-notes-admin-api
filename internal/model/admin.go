package model

import "database/sql"

type AdminUser struct {
	ID           int64        `db:"id"`
	Username     string       `db:"username"`
	PasswordHash string       `db:"password_hash"`
	Nickname     string       `db:"nickname"`
	Status       int64        `db:"status"`
	IsSuper      int64        `db:"is_super"`
	LastLoginAt  sql.NullTime `db:"last_login_at"`
}

type AdminRole struct {
	ID          int64  `db:"id"`
	Code        string `db:"code"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Status      int64  `db:"status"`
	Sort        int64  `db:"sort"`
}

type AdminPermission struct {
	ID          int64  `db:"id"`
	ParentID    int64  `db:"parent_id"`
	Code        string `db:"code"`
	Name        string `db:"name"`
	Type        string `db:"type"`
	Path        string `db:"path"`
	Method      string `db:"method"`
	Component   string `db:"component"`
	Icon        string `db:"icon"`
	Description string `db:"description"`
	Status      int64  `db:"status"`
	Sort        int64  `db:"sort"`
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
