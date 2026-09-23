package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fishing-notes-admin-api/internal/model"
	"fishing-notes-admin-api/internal/permission"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminRepository struct {
	conn sqlx.SqlConn
}

type AdminUserListFilter struct {
	Username string
	Nickname string
	Status   *int64
	Offset   int64
	Limit    int64
}

func NewAdminRepository(conn sqlx.SqlConn) *AdminRepository {
	return &AdminRepository{conn: conn}
}

func (r *AdminRepository) FindUserByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	const query = `
SELECT id, username, password_hash, nickname, status, is_super, last_login_at
FROM admin_users
WHERE username = ?
LIMIT 1`

	var user model.AdminUser
	if err := r.conn.QueryRowCtx(ctx, &user, query, username); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AdminRepository) FindUserByID(ctx context.Context, userID int64) (*model.AdminUser, error) {
	const query = `
SELECT id, username, password_hash, nickname, status, is_super, last_login_at
FROM admin_users
WHERE id = ?
LIMIT 1`

	var user model.AdminUser
	if err := r.conn.QueryRowCtx(ctx, &user, query, userID); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AdminRepository) UpdateLastLoginAt(ctx context.Context, userID int64, loginAt time.Time) error {
	const query = `UPDATE admin_users SET last_login_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.conn.ExecCtx(ctx, query, loginAt, userID)
	return err
}

func (r *AdminRepository) ListRolesByUserID(ctx context.Context, userID int64) ([]model.AdminRole, error) {
	const query = `
SELECT DISTINCT r.id, r.code, r.name, r.description, r.status, r.sort
FROM admin_roles r
INNER JOIN admin_user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = ? AND r.status = 1
ORDER BY r.sort ASC, r.id ASC`

	var roles []model.AdminRole
	if err := r.conn.QueryRowsCtx(ctx, &roles, query, userID); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *AdminRepository) ListAllRoles(ctx context.Context) ([]model.AdminRole, error) {
	const query = `
SELECT id, code, name, description, status, sort
FROM admin_roles
WHERE status = 1
ORDER BY sort ASC, id ASC`

	var roles []model.AdminRole
	if err := r.conn.QueryRowsCtx(ctx, &roles, query); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *AdminRepository) ListAllPermissions(ctx context.Context) ([]model.AdminPermission, error) {
	const query = `
SELECT id, parent_id, code, name, type, path, method, component, icon, description, status, sort
FROM admin_permissions
WHERE status = 1
ORDER BY sort ASC, id ASC`

	var permissions []model.AdminPermission
	if err := r.conn.QueryRowsCtx(ctx, &permissions, query); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *AdminRepository) ListPermissionsByUserID(ctx context.Context, userID int64) ([]model.AdminPermission, error) {
	user, err := r.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.IsSuper == 1 {
		return r.ListAllPermissions(ctx)
	}

	const query = `
SELECT DISTINCT p.id, p.parent_id, p.code, p.name, p.type, p.path, p.method, p.component, p.icon, p.description, p.status, p.sort
FROM admin_permissions p
INNER JOIN admin_role_permissions rp ON rp.permission_id = p.id
INNER JOIN admin_user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = ? AND p.status = 1
ORDER BY p.sort ASC, p.id ASC`

	var permissions []model.AdminPermission
	if err := r.conn.QueryRowsCtx(ctx, &permissions, query, userID); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *AdminRepository) ListPermissionCodesByUserID(ctx context.Context, userID int64) ([]string, error) {
	permissions, err := r.ListPermissionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		codes = append(codes, permission.Code)
	}

	return codes, nil
}

func (r *AdminRepository) ListRolePermissionPairs(ctx context.Context) ([]model.RolePermissionPair, error) {
	const query = `
SELECT rp.role_id, p.code AS permission_code
FROM admin_role_permissions rp
INNER JOIN admin_permissions p ON p.id = rp.permission_id
WHERE p.status = 1
ORDER BY rp.role_id ASC, p.sort ASC, p.id ASC`

	var pairs []model.RolePermissionPair
	if err := r.conn.QueryRowsCtx(ctx, &pairs, query); err != nil {
		return nil, err
	}

	return pairs, nil
}

func (r *AdminRepository) HasPermission(ctx context.Context, userID int64, code string) (bool, error) {
	user, err := r.FindUserByID(ctx, userID)
	if err != nil {
		return false, err
	}

	if user.Status != 1 {
		return false, nil
	}

	if user.IsSuper == 1 {
		return true, nil
	}

	const query = `
SELECT COUNT(1) AS total
FROM admin_permissions p
INNER JOIN admin_role_permissions rp ON rp.permission_id = p.id
INNER JOIN admin_user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = ? AND p.code = ? AND p.status = 1`

	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total, query, userID, code); err != nil {
		return false, err
	}

	return total > 0, nil
}

func (r *AdminRepository) ListUsers(ctx context.Context, filter AdminUserListFilter) ([]model.AdminUser, int64, error) {
	whereClause, args := buildAdminUserListWhere(filter)

	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM admin_users WHERE %s", whereClause)
	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	queryArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	listQuery := fmt.Sprintf(`
SELECT id, username, password_hash, nickname, status, is_super, last_login_at
FROM admin_users
WHERE %s
ORDER BY id DESC
LIMIT ? OFFSET ?`, whereClause)

	var users []model.AdminUser
	if err := r.conn.QueryRowsCtx(ctx, &users, listQuery, queryArgs...); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *AdminRepository) ListRolesByUserIDs(ctx context.Context, userIDs []int64) ([]model.AdminUserRole, error) {
	if len(userIDs) == 0 {
		return []model.AdminUserRole{}, nil
	}

	placeholders := make([]string, 0, len(userIDs))
	args := make([]any, 0, len(userIDs))
	for _, userID := range userIDs {
		placeholders = append(placeholders, "?")
		args = append(args, userID)
	}

	query := fmt.Sprintf(`
SELECT ur.user_id, r.id AS role_id, r.code, r.name, r.description, r.status, r.sort
FROM admin_user_roles ur
INNER JOIN admin_roles r ON r.id = ur.role_id
WHERE ur.user_id IN (%s) AND r.status = 1
ORDER BY ur.user_id ASC, r.sort ASC, r.id ASC`, strings.Join(placeholders, ", "))

	var rows []model.AdminUserRole
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *AdminRepository) SyncBuiltinPermissions(ctx context.Context, defs []permission.Definition) error {
	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)

		for _, def := range defs {
			if def.ParentCode != "" {
				continue
			}

			if err := upsertAdminPermission(ctx, conn, 0, def); err != nil {
				return err
			}
		}

		idByCode, err := queryAdminPermissionIDs(ctx, conn)
		if err != nil {
			return err
		}

		pending := make([]permission.Definition, 0, len(defs))
		for _, def := range defs {
			if def.ParentCode != "" {
				pending = append(pending, def)
			}
		}

		for len(pending) > 0 {
			nextPending := make([]permission.Definition, 0, len(pending))
			progress := false

			for _, def := range pending {
				parentID, ok := idByCode[def.ParentCode]
				if !ok {
					nextPending = append(nextPending, def)
					continue
				}

				if err := upsertAdminPermission(ctx, conn, parentID, def); err != nil {
					return err
				}
				progress = true
			}

			if !progress {
				return fmt.Errorf("failed to resolve admin permission parents")
			}

			idByCode, err = queryAdminPermissionIDs(ctx, conn)
			if err != nil {
				return err
			}
			pending = nextPending
		}

		return nil
	})
}

func buildAdminUserListWhere(filter AdminUserListFilter) (string, []any) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 3)

	if filter.Username != "" {
		conditions = append(conditions, "username LIKE ?")
		args = append(args, "%"+filter.Username+"%")
	}

	if filter.Nickname != "" {
		conditions = append(conditions, "nickname LIKE ?")
		args = append(args, "%"+filter.Nickname+"%")
	}

	if filter.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *filter.Status)
	}

	return strings.Join(conditions, " AND "), args
}

func queryAdminPermissionIDs(ctx context.Context, conn sqlx.SqlConn) (map[string]int64, error) {
	const query = `SELECT id, code FROM admin_permissions`

	var items []model.AdminPermissionIDItem
	if err := conn.QueryRowsCtx(ctx, &items, query); err != nil {
		return nil, err
	}

	idByCode := make(map[string]int64, len(items))
	for _, item := range items {
		idByCode[item.Code] = item.ID
	}

	return idByCode, nil
}

func upsertAdminPermission(ctx context.Context, conn sqlx.SqlConn, parentID int64, def permission.Definition) error {
	const query = `
INSERT INTO admin_permissions (parent_id, code, name, type, path, method, component, icon, description, status, sort)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	parent_id = VALUES(parent_id),
	name = VALUES(name),
	type = VALUES(type),
	path = VALUES(path),
	method = VALUES(method),
	component = VALUES(component),
	icon = VALUES(icon),
	description = VALUES(description),
	status = VALUES(status),
	sort = VALUES(sort),
	updated_at = CURRENT_TIMESTAMP`

	_, err := conn.ExecCtx(
		ctx,
		query,
		parentID,
		def.Code,
		def.Name,
		def.Type,
		def.Path,
		def.Method,
		def.Component,
		def.Icon,
		def.Description,
		def.Status,
		def.Sort,
	)
	return err
}
