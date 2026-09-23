package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"fishing-notes-admin-api/internal/assembler"
	authpkg "fishing-notes-admin-api/internal/auth"
	"fishing-notes-admin-api/internal/config"
	"fishing-notes-admin-api/internal/errorsx"
	"fishing-notes-admin-api/internal/repository"
	"fishing-notes-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminService struct {
	repo   *repository.AdminRepository
	config config.AuthConf
}

func NewAdminService(repo *repository.AdminRepository, cfg config.AuthConf) *AdminService {
	return &AdminService{
		repo:   repo,
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
