package service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"fishing-notes-admin-api/internal/assembler"
	"fishing-notes-admin-api/internal/audit"
	authpkg "fishing-notes-admin-api/internal/auth"
	"fishing-notes-admin-api/internal/config"
	"fishing-notes-admin-api/internal/errorsx"
	"fishing-notes-admin-api/internal/repository"
	"fishing-notes-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminService struct {
	repo   *repository.AdminRepository
	audit  *repository.AuditRepository
	config config.AuthConf
}

func NewAdminService(repo *repository.AdminRepository, cfg config.AuthConf, auditRepos ...*repository.AuditRepository) *AdminService {
	var auditRepo *repository.AuditRepository
	if len(auditRepos) > 0 {
		auditRepo = auditRepos[0]
	}
	return &AdminService{
		repo:   repo,
		audit:  auditRepo,
		config: cfg,
	}
}

func (s *AdminService) Login(ctx context.Context, req *types.LoginRequest) (*types.LoginResponse, error) {
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		return nil, errorsx.BadRequest("username and password are required")
	}

	user, err := s.repo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.Unauthorized("username or password is incorrect")
		}
		return nil, errorsx.Internal("failed to query user")
	}

	if user.Status != 1 {
		return nil, errorsx.Forbidden("user is disabled")
	}

	if err := authpkg.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, errorsx.Unauthorized("username or password is incorrect")
	}

	token, expiresAt, err := authpkg.GenerateToken(
		s.config.AccessSecret,
		s.config.AccessExpire,
		user.ID,
		user.Username,
	)
	if err != nil {
		return nil, errorsx.Internal("failed to generate access token")
	}

	currentUser, err := s.buildCurrentUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateLastLoginAt(ctx, user.ID, time.Now()); err != nil {
		// Login should still succeed if the audit field update fails.
	}

	return &types.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        *currentUser,
	}, nil
}

func (s *AdminService) CurrentUser(ctx context.Context) (*types.CurrentUser, error) {
	userID, err := s.mustUserID(ctx)
	if err != nil {
		return nil, err
	}

	return s.buildCurrentUser(ctx, userID)
}

func (s *AdminService) PermissionCodes(ctx context.Context) (*types.PermissionCodesResponse, error) {
	userID, err := s.mustUserID(ctx)
	if err != nil {
		return nil, err
	}

	codes, err := s.repo.ListPermissionCodesByUserID(ctx, userID)
	if err != nil {
		return nil, errorsx.Internal("failed to query permission codes")
	}

	return &types.PermissionCodesResponse{Permissions: codes}, nil
}

func (s *AdminService) Menus(ctx context.Context) (*types.MenusResponse, error) {
	userID, err := s.mustUserID(ctx)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.ListPermissionsByUserID(ctx, userID)
	if err != nil {
		return nil, errorsx.Internal("failed to query menus")
	}

	return &types.MenusResponse{
		Menus: assembler.Menus(permissions),
	}, nil
}

func (s *AdminService) ListRoles(ctx context.Context) (*types.ListRolesResponse, error) {
	roles, err := s.repo.ListAllRoles(ctx)
	if err != nil {
		return nil, errorsx.Internal("failed to query roles")
	}

	pairs, err := s.repo.ListRolePermissionPairs(ctx)
	if err != nil {
		return nil, errorsx.Internal("failed to query role permissions")
	}

	return &types.ListRolesResponse{
		Roles: assembler.Roles(roles, assembler.GroupPermissionCodesByRole(pairs)),
	}, nil
}

func (s *AdminService) ListPermissions(ctx context.Context) (*types.ListPermissionsResponse, error) {
	permissions, err := s.repo.ListAllPermissions(ctx)
	if err != nil {
		return nil, errorsx.Internal("failed to query permissions")
	}

	return &types.ListPermissionsResponse{
		Permissions: assembler.Permissions(permissions),
	}, nil
}

func (s *AdminService) CreateRole(ctx context.Context, req *types.CreateRoleRequest) (*types.RoleItem, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if err := validateRoleCode(code); err != nil {
		return nil, errorsx.BadRequest("invalid role code")
	}
	if name == "" || len(name) > 64 || len(req.Description) > 255 || req.Status < 0 || req.Status > 1 || req.Sort < 0 {
		return nil, errorsx.BadRequest("invalid role")
	}
	if err := validateIDs(req.PermissionIDs); err != nil {
		return nil, errorsx.BadRequest("invalid permission ids")
	}
	if _, err := s.repo.FindRoleByCode(ctx, code); err == nil {
		return nil, errorsx.BadRequest("role code already exists")
	} else if !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorsx.Internal("failed to query role")
	}
	permissionCount, err := s.repo.CountActivePermissions(ctx, req.PermissionIDs)
	if err != nil {
		return nil, errorsx.Internal("failed to query permissions")
	}
	if permissionCount != int64(len(req.PermissionIDs)) {
		return nil, errorsx.BadRequest("one or more permissions are invalid")
	}
	role, err := s.repo.CreateRole(ctx, code, name, strings.TrimSpace(req.Description), req.Status, req.Sort, req.PermissionIDs)
	if err != nil {
		return nil, errorsx.Internal("failed to create role")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "rbac.role.create", ResourceType: "role", ResourceID: role.ID, Detail: map[string]any{"code": role.Code, "name": role.Name}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}
	codes, err := s.repo.ListPermissionCodesByRoleID(ctx, role.ID)
	if err != nil {
		return nil, errorsx.Internal("failed to query role permissions")
	}
	return &types.RoleItem{
		ID: role.ID, Code: role.Code, Name: role.Name, Description: role.Description,
		Status: role.Status, Sort: role.Sort, PermissionCodes: codes,
	}, nil
}

func (s *AdminService) UpdateRole(ctx context.Context, req *types.UpdateRoleRequest) (*types.RoleItem, error) {
	if req.RoleID <= 0 || strings.TrimSpace(req.Name) == "" || len(strings.TrimSpace(req.Name)) > 64 || len(req.Description) > 255 || req.Status < 0 || req.Status > 1 || req.Sort < 0 {
		return nil, errorsx.BadRequest("invalid role")
	}
	role, err := s.repo.FindRoleByID(ctx, req.RoleID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("role not found")
		}
		return nil, errorsx.Internal("failed to query role")
	}
	if err := s.repo.UpdateRole(ctx, req.RoleID, strings.TrimSpace(req.Name), strings.TrimSpace(req.Description), req.Status, req.Sort); err != nil {
		return nil, errorsx.Internal("failed to update role")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "rbac.role.update", ResourceType: "role", ResourceID: req.RoleID, Detail: map[string]any{"name": strings.TrimSpace(req.Name), "status": req.Status, "sort": req.Sort}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}
	role.Name = strings.TrimSpace(req.Name)
	role.Description = strings.TrimSpace(req.Description)
	role.Status = req.Status
	role.Sort = req.Sort
	codes, err := s.repo.ListPermissionCodesByRoleID(ctx, role.ID)
	if err != nil {
		return nil, errorsx.Internal("failed to query role permissions")
	}
	return &types.RoleItem{
		ID: role.ID, Code: role.Code, Name: role.Name, Description: role.Description,
		Status: role.Status, Sort: role.Sort, PermissionCodes: codes,
	}, nil
}

func (s *AdminService) DeleteRole(ctx context.Context, req *types.RolePathRequest) error {
	if req.RoleID <= 0 {
		return errorsx.BadRequest("invalid role id")
	}
	role, err := s.repo.FindRoleByID(ctx, req.RoleID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return errorsx.NotFound("role not found")
		}
		return errorsx.Internal("failed to query role")
	}
	if role.Code == "super_admin" {
		return errorsx.BadRequest("super admin role cannot be deleted")
	}
	users, err := s.repo.CountUsersByRoleID(ctx, req.RoleID)
	if err != nil {
		return errorsx.Internal("failed to query role users")
	}
	if users > 0 {
		return errorsx.BadRequest("role is still assigned to administrators")
	}
	if err := s.repo.DeleteRole(ctx, req.RoleID); err != nil {
		return errorsx.Internal("failed to delete role")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "rbac.role.delete", ResourceType: "role", ResourceID: req.RoleID, Detail: map[string]any{"code": role.Code}}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}
	return nil
}

func (s *AdminService) UserRoles(ctx context.Context, req *types.UserRolesRequest) (*types.UserRolesResponse, error) {
	_, err := s.repo.FindUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("user not found")
		}
		return nil, errorsx.Internal("failed to query user")
	}

	roles, err := s.repo.ListRolesByUserID(ctx, req.UserID)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorsx.Internal("failed to query user roles")
	}

	return &types.UserRolesResponse{
		UserID: req.UserID,
		Roles:  assembler.Roles(roles, nil),
	}, nil
}

func (s *AdminService) UpdateUserRoles(ctx context.Context, req *types.UpdateUserRolesRequest) (*types.UserRolesResponse, error) {
	if req.UserID <= 0 {
		return nil, errorsx.BadRequest("invalid user id")
	}

	if err := validateIDs(req.RoleIDs); err != nil {
		return nil, errorsx.BadRequest("invalid role ids")
	}

	if _, err := s.repo.FindUserByID(ctx, req.UserID); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("user not found")
		}
		return nil, errorsx.Internal("failed to query user")
	}

	roleCount, err := s.repo.CountActiveRoles(ctx, req.RoleIDs)
	if err != nil {
		return nil, errorsx.Internal("failed to query roles")
	}
	if roleCount != int64(len(req.RoleIDs)) {
		return nil, errorsx.BadRequest("one or more roles are invalid")
	}

	if err := s.repo.ReplaceUserRoles(ctx, req.UserID, req.RoleIDs); err != nil {
		return nil, errorsx.Internal("failed to update user roles")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "admin.user.roles.update", ResourceType: "admin_user", ResourceID: req.UserID, Detail: map[string]any{"roleIds": req.RoleIDs}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}

	return s.UserRoles(ctx, &types.UserRolesRequest{UserID: req.UserID})
}

func (s *AdminService) UpdateRolePermissions(ctx context.Context, req *types.UpdateRolePermissionsRequest) (*types.RolePermissionsResponse, error) {
	if req.RoleID <= 0 {
		return nil, errorsx.BadRequest("invalid role id")
	}

	if err := validateIDs(req.PermissionIDs); err != nil {
		return nil, errorsx.BadRequest("invalid permission ids")
	}

	role, err := s.repo.FindRoleByID(ctx, req.RoleID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("role not found")
		}
		return nil, errorsx.Internal("failed to query role")
	}
	if role.Status != 1 {
		return nil, errorsx.BadRequest("role is disabled")
	}

	permissionCount, err := s.repo.CountActivePermissions(ctx, req.PermissionIDs)
	if err != nil {
		return nil, errorsx.Internal("failed to query permissions")
	}
	if permissionCount != int64(len(req.PermissionIDs)) {
		return nil, errorsx.BadRequest("one or more permissions are invalid")
	}

	if err := s.repo.ReplaceRolePermissions(ctx, req.RoleID, req.PermissionIDs); err != nil {
		return nil, errorsx.Internal("failed to update role permissions")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "rbac.role.permissions.update", ResourceType: "role", ResourceID: req.RoleID, Detail: map[string]any{"permissionIds": req.PermissionIDs}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}

	return &types.RolePermissionsResponse{RoleID: req.RoleID, PermissionIDs: req.PermissionIDs}, nil
}

func (s *AdminService) ListUsers(ctx context.Context, req *types.ListAdminUsersRequest) (*types.ListAdminUsersResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := repository.AdminUserListFilter{
		Username: strings.TrimSpace(req.Username),
		Nickname: strings.TrimSpace(req.Nickname),
		Offset:   (page - 1) * pageSize,
		Limit:    pageSize,
	}

	if strings.TrimSpace(req.Status) != "" {
		status, err := strconv.ParseInt(strings.TrimSpace(req.Status), 10, 64)
		if err != nil {
			return nil, errorsx.BadRequest("invalid status")
		}
		filter.Status = &status
	}

	users, total, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		return nil, errorsx.Internal("failed to query users")
	}

	userIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	userRoles, err := s.repo.ListRolesByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, errorsx.Internal("failed to query user roles")
	}

	return &types.ListAdminUsersResponse{
		List:     assembler.AdminUsers(users, assembler.GroupRolesByUser(userRoles)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) CreateAdminUser(ctx context.Context, req *types.CreateAdminUserRequest) (*types.AdminUserItem, error) {
	username := strings.TrimSpace(req.Username)
	nickname := strings.TrimSpace(req.Nickname)
	if username == "" || len(username) > 64 {
		return nil, errorsx.BadRequest("invalid username")
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, errorsx.BadRequest("invalid password")
	}
	if nickname == "" || len(nickname) > 64 {
		return nil, errorsx.BadRequest("invalid nickname")
	}
	if err := validateIDs(req.RoleIDs); err != nil {
		return nil, errorsx.BadRequest("invalid role ids")
	}
	if _, err := s.repo.FindUserByUsername(ctx, username); err == nil {
		return nil, errorsx.BadRequest("username already exists")
	} else if !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorsx.Internal("failed to query user")
	}
	roleCount, err := s.repo.CountActiveRoles(ctx, req.RoleIDs)
	if err != nil {
		return nil, errorsx.Internal("failed to query roles")
	}
	if roleCount != int64(len(req.RoleIDs)) {
		return nil, errorsx.BadRequest("one or more roles are invalid")
	}
	passwordHash, err := authpkg.HashPassword(req.Password)
	if err != nil {
		return nil, errorsx.Internal("failed to hash password")
	}
	user, err := s.repo.CreateAdminUser(ctx, username, passwordHash, nickname, req.RoleIDs)
	if err != nil {
		return nil, errorsx.Internal("failed to create admin user")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "admin.user.create", ResourceType: "admin_user", ResourceID: user.ID, Detail: map[string]any{"username": user.Username, "roleIds": req.RoleIDs}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}
	roles, err := s.repo.ListRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, errorsx.Internal("failed to query user roles")
	}
	return assembler.AdminUser(user, assembler.Roles(roles, nil)), nil
}

func (s *AdminService) UpdateAdminUser(ctx context.Context, req *types.UpdateAdminUserRequest) (*types.AdminUserItem, error) {
	if req.UserID <= 0 || (req.Status != 0 && req.Status != 1) {
		return nil, errorsx.BadRequest("invalid admin user")
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" || len(nickname) > 64 {
		return nil, errorsx.BadRequest("invalid nickname")
	}
	user, err := s.repo.FindUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("admin user not found")
		}
		return nil, errorsx.Internal("failed to query admin user")
	}
	if err := s.repo.UpdateAdminUser(ctx, req.UserID, nickname, req.Status); err != nil {
		return nil, errorsx.Internal("failed to update admin user")
	}
	user.Nickname = nickname
	user.Status = req.Status
	if err := s.writeAudit(ctx, audit.Event{Action: "admin.user.update", ResourceType: "admin_user", ResourceID: req.UserID, Detail: map[string]any{"status": req.Status}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}
	return assembler.AdminUser(user, nil), nil
}

func (s *AdminService) ResetAdminPassword(ctx context.Context, req *types.ResetAdminPasswordRequest) error {
	if req.UserID <= 0 {
		return errorsx.BadRequest("invalid admin user id")
	}
	if err := validatePassword(req.Password); err != nil {
		return errorsx.BadRequest("invalid password")
	}
	if _, err := s.repo.FindUserByID(ctx, req.UserID); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return errorsx.NotFound("admin user not found")
		}
		return errorsx.Internal("failed to query admin user")
	}
	passwordHash, err := authpkg.HashPassword(req.Password)
	if err != nil {
		return errorsx.Internal("failed to hash password")
	}
	if err := s.repo.UpdateAdminPassword(ctx, req.UserID, passwordHash); err != nil {
		return errorsx.Internal("failed to reset admin password")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "admin.user.password.update", ResourceType: "admin_user", ResourceID: req.UserID}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}
	return nil
}

func (s *AdminService) HasPermission(ctx context.Context, permissionCode string) (bool, error) {
	userID, err := s.mustUserID(ctx)
	if err != nil {
		return false, err
	}

	allowed, repoErr := s.repo.HasPermission(ctx, userID, permissionCode)
	if repoErr != nil {
		return false, errorsx.Internal("failed to validate permission")
	}

	return allowed, nil
}

func (s *AdminService) buildCurrentUser(ctx context.Context, userID int64) (*types.CurrentUser, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("user not found")
		}
		return nil, errorsx.Internal("failed to query user")
	}

	roles, err := s.repo.ListRolesByUserID(ctx, userID)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorsx.Internal("failed to query user roles")
	}

	permissionCodes, err := s.repo.ListPermissionCodesByUserID(ctx, userID)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorsx.Internal("failed to query user permissions")
	}

	return &types.CurrentUser{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		IsSuper:     user.IsSuper == 1,
		Status:      user.Status,
		LastLoginAt: assembler.NullTimePtr(user.LastLoginAt),
		Roles:       assembler.Roles(roles, nil),
		Permissions: permissionCodes,
	}, nil
}

func (s *AdminService) mustUserID(ctx context.Context) (int64, error) {
	userID, ok := authpkg.UserIDFromContext(ctx)
	if !ok || userID <= 0 {
		return 0, errorsx.Unauthorized("missing auth user")
	}

	return userID, nil
}

func (s *AdminService) writeAudit(ctx context.Context, event audit.Event) error {
	if s.audit == nil {
		return nil
	}
	actorUserID := event.ActorUserID
	if actorUserID <= 0 {
		actorUserID, _ = authpkg.UserIDFromContext(ctx)
	}
	detail, err := audit.MarshalDetail(event.Detail)
	if err != nil {
		slog.WarnContext(ctx, "failed to serialize admin audit detail", "action", event.Action, "resource_type", event.ResourceType, "resource_id", event.ResourceID, "error", err)
		return nil
	}
	if err := s.audit.Append(ctx, actorUserID, event.Action, event.ResourceType, event.ResourceID, detail); err != nil {
		slog.WarnContext(ctx, "failed to append admin audit log", "action", event.Action, "resource_type", event.ResourceType, "resource_id", event.ResourceID, "error", err)
	}
	return nil
}

func validateIDs(ids []int64) error {
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return errors.New("id must be positive")
		}
		if _, ok := seen[id]; ok {
			return errors.New("id must be unique")
		}
		seen[id] = struct{}{}
	}

	return nil
}

func validateRoleCode(code string) error {
	if code == "" || len(code) > 64 || strings.ContainsAny(code, " \t\r\n") {
		return errors.New("invalid role code")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 6 || len(password) > 72 {
		return errors.New("password length must be between 6 and 72")
	}
	return nil
}
