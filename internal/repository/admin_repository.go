package repository

import (
	"context"
	"errors"
	"time"

	"fishing-notes-admin-api/internal/model"
	"fishing-notes-admin-api/internal/permission"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminRepository struct {
	db *gorm.DB
}

type AdminUserListFilter struct {
	Username string
	Nickname string
	Status   *int64
	Offset   int64
	Limit    int64
}

type adminUserRoleRepositoryModel struct {
	UserID int64 `gorm:"column:user_id;primaryKey"`
	RoleID int64 `gorm:"column:role_id;primaryKey"`
}

func (adminUserRoleRepositoryModel) TableName() string { return "admin_user_roles" }

type adminRolePermissionRepositoryModel struct {
	RoleID       int64 `gorm:"column:role_id;primaryKey"`
	PermissionID int64 `gorm:"column:permission_id;primaryKey"`
}

// AutoMigrate 创建管理端自有表，复用业务查询 model 作为 GORM schema 定义。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.AdminUser{}, &model.AdminRole{}, &model.AdminPermission{},
		&adminUserRoleRepositoryModel{}, &adminRolePermissionRepositoryModel{},
		&adminAuditLogModel{},
	)
}

func (adminRolePermissionRepositoryModel) TableName() string { return "admin_role_permissions" }

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) FindUserByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var user model.AdminUser
	err := r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("username = ?", username).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AdminRepository) FindUserByID(ctx context.Context, userID int64) (*model.AdminUser, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var user model.AdminUser
	err := r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", userID).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AdminRepository) UpdateLastLoginAt(ctx context.Context, userID int64, loginAt time.Time) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", userID).Updates(map[string]any{
		"last_login_at": loginAt, "updated_at": loginAt,
	}).Error
}

func (r *AdminRepository) CreateAdminUser(ctx context.Context, username, passwordHash, nickname string, roleIDs []int64) (*model.AdminUser, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	user := model.AdminUser{Username: username, PasswordHash: passwordHash, Nickname: nickname, Status: 1}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AdminUser{}).Create(&user).Error; err != nil {
			return err
		}
		links := make([]adminUserRoleRepositoryModel, len(roleIDs))
		for index, roleID := range roleIDs {
			links[index] = adminUserRoleRepositoryModel{UserID: user.ID, RoleID: roleID}
		}
		if len(links) == 0 {
			return nil
		}
		return tx.Create(&links).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindUserByID(ctx, user.ID)
}

func (r *AdminRepository) UpdateAdminUser(ctx context.Context, userID int64, nickname string, status int64) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", userID).Updates(map[string]any{
		"nickname": nickname, "status": status, "updated_at": time.Now(),
	}).Error
}

func (r *AdminRepository) UpdateAdminPassword(ctx context.Context, userID int64, passwordHash string) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", userID).Updates(map[string]any{
		"password_hash": passwordHash, "updated_at": time.Now(),
	}).Error
}

func (r *AdminRepository) ListRolesByUserID(ctx context.Context, userID int64) ([]model.AdminRole, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var links []adminUserRoleRepositoryModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&links).Error; err != nil {
		return nil, err
	}
	roleIDs := make([]int64, 0, len(links))
	for _, link := range links {
		roleIDs = append(roleIDs, link.RoleID)
	}
	roles := make([]model.AdminRole, 0)
	if len(roleIDs) == 0 {
		return roles, nil
	}
	if err := r.db.WithContext(ctx).Model(&model.AdminRole{}).Where("id IN ? AND status = ?", roleIDs, 1).
		Order("sort ASC").Order("id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *AdminRepository) FindRoleByID(ctx context.Context, roleID int64) (*model.AdminRole, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var role model.AdminRole
	err := r.db.WithContext(ctx).Model(&model.AdminRole{}).Where("id = ?", roleID).Take(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *AdminRepository) FindRoleByCode(ctx context.Context, code string) (*model.AdminRole, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var role model.AdminRole
	err := r.db.WithContext(ctx).Model(&model.AdminRole{}).Where("code = ?", code).Take(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *AdminRepository) CreateRole(ctx context.Context, code, name, description string, status, sort int64, permissionIDs []int64) (*model.AdminRole, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	role := model.AdminRole{Code: code, Name: name, Description: description, Status: status, Sort: sort}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AdminRole{}).Create(&role).Error; err != nil {
			return err
		}
		links := make([]adminRolePermissionRepositoryModel, len(permissionIDs))
		for index, permissionID := range permissionIDs {
			links[index] = adminRolePermissionRepositoryModel{RoleID: role.ID, PermissionID: permissionID}
		}
		if len(links) == 0 {
			return nil
		}
		return tx.Create(&links).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindRoleByID(ctx, role.ID)
}

func (r *AdminRepository) UpdateRole(ctx context.Context, roleID int64, name, description string, status, sort int64) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.AdminRole{}).Where("id = ?", roleID).Updates(map[string]any{
		"name": name, "description": description, "status": status, "sort": sort,
	}).Error
}

func (r *AdminRepository) CountUsersByRoleID(ctx context.Context, roleID int64) (int64, error) {
	if r.db == nil {
		return 0, errors.New("GORM database is not configured")
	}
	var total int64
	err := r.db.WithContext(ctx).Model(&adminUserRoleRepositoryModel{}).Where("role_id = ?", roleID).Count(&total).Error
	return total, err
}

func (r *AdminRepository) DeleteRole(ctx context.Context, roleID int64) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&adminRolePermissionRepositoryModel{}).Error; err != nil {
			return err
		}
		return tx.Model(&model.AdminRole{}).Where("id = ?", roleID).Delete(&model.AdminRole{}).Error
	})
}

func (r *AdminRepository) ReplaceUserRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&adminUserRoleRepositoryModel{}).Error; err != nil {
			return err
		}
		links := make([]adminUserRoleRepositoryModel, len(roleIDs))
		for index, roleID := range roleIDs {
			links[index] = adminUserRoleRepositoryModel{UserID: userID, RoleID: roleID}
		}
		if len(links) == 0 {
			return nil
		}
		return tx.Create(&links).Error
	})
}

func (r *AdminRepository) ReplaceRolePermissions(ctx context.Context, roleID int64, permissionIDs []int64) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&adminRolePermissionRepositoryModel{}).Error; err != nil {
			return err
		}
		links := make([]adminRolePermissionRepositoryModel, len(permissionIDs))
		for index, permissionID := range permissionIDs {
			links[index] = adminRolePermissionRepositoryModel{RoleID: roleID, PermissionID: permissionID}
		}
		if len(links) == 0 {
			return nil
		}
		return tx.Create(&links).Error
	})
}

func (r *AdminRepository) CountActiveRoles(ctx context.Context, roleIDs []int64) (int64, error) {
	if len(roleIDs) == 0 {
		return 0, nil
	}
	if r.db == nil {
		return 0, errors.New("GORM database is not configured")
	}
	var total int64
	err := r.db.WithContext(ctx).Model(&model.AdminRole{}).Where("status = ? AND id IN ?", 1, roleIDs).Count(&total).Error
	return total, err
}

func (r *AdminRepository) CountActivePermissions(ctx context.Context, permissionIDs []int64) (int64, error) {
	if len(permissionIDs) == 0 {
		return 0, nil
	}

	if r.db == nil {
		return 0, errors.New("GORM database is not configured")
	}
	var total int64
	err := r.db.WithContext(ctx).Model(&model.AdminPermission{}).Where("status = ? AND id IN ?", 1, permissionIDs).Count(&total).Error
	return total, err
}

func (r *AdminRepository) ListAllRoles(ctx context.Context) ([]model.AdminRole, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	roles := make([]model.AdminRole, 0)
	if err := r.db.WithContext(ctx).Model(&model.AdminRole{}).Order("sort ASC").Order("id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *AdminRepository) ListAllPermissions(ctx context.Context) ([]model.AdminPermission, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	permissions := make([]model.AdminPermission, 0)
	if err := r.db.WithContext(ctx).Model(&model.AdminPermission{}).Where("status = ?", 1).
		Order("sort ASC").Order("id ASC").Find(&permissions).Error; err != nil {
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

	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var userLinks []adminUserRoleRepositoryModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userLinks).Error; err != nil {
		return nil, err
	}
	roleIDs := make([]int64, 0, len(userLinks))
	for _, link := range userLinks {
		roleIDs = append(roleIDs, link.RoleID)
	}
	if len(roleIDs) == 0 {
		return []model.AdminPermission{}, nil
	}
	var roleLinks []adminRolePermissionRepositoryModel
	if err := r.db.WithContext(ctx).Where("role_id IN ?", roleIDs).Find(&roleLinks).Error; err != nil {
		return nil, err
	}
	permissionIDs := make([]int64, 0, len(roleLinks))
	for _, link := range roleLinks {
		permissionIDs = append(permissionIDs, link.PermissionID)
	}
	if len(permissionIDs) == 0 {
		return []model.AdminPermission{}, nil
	}
	permissions := make([]model.AdminPermission, 0)
	if err := r.db.WithContext(ctx).Model(&model.AdminPermission{}).Where("id IN ? AND status = ?", permissionIDs, 1).
		Order("sort ASC").Order("id ASC").Find(&permissions).Error; err != nil {
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
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var links []adminRolePermissionRepositoryModel
	if err := r.db.WithContext(ctx).Find(&links).Error; err != nil {
		return nil, err
	}
	permissionIDs := make([]int64, 0, len(links))
	for _, link := range links {
		permissionIDs = append(permissionIDs, link.PermissionID)
	}
	if len(permissionIDs) == 0 {
		return []model.RolePermissionPair{}, nil
	}
	permissions := make([]model.AdminPermission, 0)
	if err := r.db.WithContext(ctx).Model(&model.AdminPermission{}).Where("id IN ? AND status = ?", permissionIDs, 1).
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	permissionByID := make(map[int64]model.AdminPermission, len(permissions))
	for _, item := range permissions {
		permissionByID[item.ID] = item
	}
	pairs := make([]model.RolePermissionPair, 0, len(links))
	for _, link := range links {
		if permission, ok := permissionByID[link.PermissionID]; ok {
			pairs = append(pairs, model.RolePermissionPair{RoleID: link.RoleID, PermissionCode: permission.Code})
		}
	}
	return pairs, nil
}

func (r *AdminRepository) ListPermissionCodesByRoleID(ctx context.Context, roleID int64) ([]string, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var links []adminRolePermissionRepositoryModel
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&links).Error; err != nil {
		return nil, err
	}
	permissionIDs := make([]int64, 0, len(links))
	for _, link := range links {
		permissionIDs = append(permissionIDs, link.PermissionID)
	}
	if len(permissionIDs) == 0 {
		return []string{}, nil
	}
	permissions := make([]model.AdminPermission, 0)
	if err := r.db.WithContext(ctx).Model(&model.AdminPermission{}).Where("id IN ? AND status = ?", permissionIDs, 1).
		Order("sort ASC").Order("id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		codes = append(codes, permission.Code)
	}
	return codes, nil
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

	if r.db == nil {
		return false, errors.New("GORM database is not configured")
	}
	var userLinks []adminUserRoleRepositoryModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userLinks).Error; err != nil {
		return false, err
	}
	roleIDs := make([]int64, 0, len(userLinks))
	for _, link := range userLinks {
		roleIDs = append(roleIDs, link.RoleID)
	}
	if len(roleIDs) == 0 {
		return false, nil
	}
	var roleLinks []adminRolePermissionRepositoryModel
	if err := r.db.WithContext(ctx).Where("role_id IN ?", roleIDs).Find(&roleLinks).Error; err != nil {
		return false, err
	}
	permissionIDs := make([]int64, 0, len(roleLinks))
	for _, link := range roleLinks {
		permissionIDs = append(permissionIDs, link.PermissionID)
	}
	if len(permissionIDs) == 0 {
		return false, nil
	}
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.AdminPermission{}).Where("id IN ? AND code = ? AND status = ?", permissionIDs, code, 1).Count(&total).Error; err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *AdminRepository) ListUsers(ctx context.Context, filter AdminUserListFilter) ([]model.AdminUser, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	query := r.db.WithContext(ctx).Model(&model.AdminUser{})
	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+filter.Nickname+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.AdminUser{}, 0, nil
	}
	users := make([]model.AdminUser, 0)
	if err := query.Order("id DESC").Limit(int(filter.Limit)).Offset(int(filter.Offset)).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *AdminRepository) ListRolesByUserIDs(ctx context.Context, userIDs []int64) ([]model.AdminUserRole, error) {
	if len(userIDs) == 0 {
		return []model.AdminUserRole{}, nil
	}

	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var links []adminUserRoleRepositoryModel
	if err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []model.AdminUserRole{}, nil
	}
	roleIDs := make([]int64, 0, len(links))
	for _, link := range links {
		roleIDs = append(roleIDs, link.RoleID)
	}
	roles := make([]model.AdminRole, 0)
	if err := r.db.WithContext(ctx).Model(&model.AdminRole{}).Where("id IN ? AND status = ?", roleIDs, 1).Find(&roles).Error; err != nil {
		return nil, err
	}
	roleByID := make(map[int64]model.AdminRole, len(roles))
	for _, role := range roles {
		roleByID[role.ID] = role
	}
	rows := make([]model.AdminUserRole, 0, len(links))
	for _, link := range links {
		if role, ok := roleByID[link.RoleID]; ok {
			rows = append(rows, model.AdminUserRole{
				UserID: link.UserID, RoleID: role.ID, Code: role.Code, Name: role.Name,
				Description: role.Description, Status: role.Status, Sort: role.Sort,
			})
		}
	}
	return rows, nil
}

func (r *AdminRepository) SyncBuiltinPermissions(ctx context.Context, defs []permission.Definition) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, def := range defs {
			if def.ParentCode != "" {
				continue
			}

			if err := upsertAdminPermission(tx, 0, def); err != nil {
				return err
			}
		}

		idByCode, err := queryAdminPermissionIDs(tx)
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

				if err := upsertAdminPermission(tx, parentID, def); err != nil {
					return err
				}
				progress = true
			}

			if !progress {
				return errors.New("failed to resolve admin permission parents")
			}

			idByCode, err = queryAdminPermissionIDs(tx)
			if err != nil {
				return err
			}
			pending = nextPending
		}

		return nil
	})
}

func queryAdminPermissionIDs(tx *gorm.DB) (map[string]int64, error) {
	var items []struct {
		ID   int64  `gorm:"column:id"`
		Code string `gorm:"column:code"`
	}
	if err := tx.Model(&model.AdminPermission{}).Select("id", "code").Find(&items).Error; err != nil {
		return nil, err
	}

	idByCode := make(map[string]int64, len(items))
	for _, item := range items {
		idByCode[item.Code] = item.ID
	}

	return idByCode, nil
}

func upsertAdminPermission(tx *gorm.DB, parentID int64, def permission.Definition) error {
	item := model.AdminPermission{
		ParentID: parentID, Code: def.Code, Name: def.Name, Type: def.Type, Path: def.Path,
		Method: def.Method, Component: def.Component, Icon: def.Icon, Description: def.Description,
		Status: def.Status, Sort: def.Sort,
	}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"parent_id", "name", "type", "path", "method", "component", "icon", "description", "status", "sort",
		}),
	}).Create(&item).Error
}
