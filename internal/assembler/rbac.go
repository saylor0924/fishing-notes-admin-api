package assembler

import (
	"fishing-notes-admin-api/internal/model"
	"fishing-notes-admin-api/internal/types"
)

func Roles(items []model.AdminRole, permissionCodes map[int64][]string) []types.RoleItem {
	roles := make([]types.RoleItem, 0, len(items))
	for _, item := range items {
		roles = append(roles, types.RoleItem{
			ID:              item.ID,
			Code:            item.Code,
			Name:            item.Name,
			Description:     item.Description,
			Status:          item.Status,
			Sort:            item.Sort,
			PermissionCodes: permissionCodes[item.ID],
		})
	}

	return roles
}

func Permissions(items []model.AdminPermission) []types.PermissionItem {
	permissions := make([]types.PermissionItem, 0, len(items))
	for _, item := range items {
		permissions = append(permissions, types.PermissionItem{
			ID:          item.ID,
			ParentID:    item.ParentID,
			Code:        item.Code,
			Name:        item.Name,
			Type:        item.Type,
			Path:        item.Path,
			Method:      item.Method,
			Component:   item.Component,
			Icon:        item.Icon,
			Description: item.Description,
			Status:      item.Status,
			Sort:        item.Sort,
		})
	}

	return permissions
}

func Menus(items []model.AdminPermission) []types.MenuItem {
	childrenByParent := make(map[int64][]types.MenuItem)

	for _, item := range items {
		if item.Type != "menu" {
			continue
		}

		childrenByParent[item.ParentID] = append(childrenByParent[item.ParentID], types.MenuItem{
			ID:        item.ID,
			ParentID:  item.ParentID,
			Code:      item.Code,
			Name:      item.Name,
			Path:      item.Path,
			Component: item.Component,
			Icon:      item.Icon,
			Sort:      item.Sort,
		})
	}

	return buildMenuTree(0, childrenByParent)
}

func GroupPermissionCodesByRole(pairs []model.RolePermissionPair) map[int64][]string {
	grouped := make(map[int64][]string)
	for _, pair := range pairs {
		grouped[pair.RoleID] = append(grouped[pair.RoleID], pair.PermissionCode)
	}

	return grouped
}

func GroupRolesByUser(items []model.AdminUserRole) map[int64][]types.RoleItem {
	grouped := make(map[int64][]types.RoleItem)
	for _, item := range items {
		grouped[item.UserID] = append(grouped[item.UserID], types.RoleItem{
			ID:          item.RoleID,
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			Status:      item.Status,
			Sort:        item.Sort,
		})
	}

	return grouped
}

func AdminUsers(items []model.AdminUser, rolesByUser map[int64][]types.RoleItem) []types.AdminUserItem {
	users := make([]types.AdminUserItem, 0, len(items))
	for _, item := range items {
		roles := rolesByUser[item.ID]
		if roles == nil {
			roles = []types.RoleItem{}
		}

		users = append(users, types.AdminUserItem{
			ID:          item.ID,
			Username:    item.Username,
			Nickname:    item.Nickname,
			IsSuper:     item.IsSuper == 1,
			Status:      item.Status,
			LastLoginAt: NullTimePtr(item.LastLoginAt),
			Roles:       roles,
		})
	}

	return users
}

func AdminUser(item *model.AdminUser, roles []types.RoleItem) *types.AdminUserItem {
	if roles == nil {
		roles = []types.RoleItem{}
	}
	return &types.AdminUserItem{
		ID:          item.ID,
		Username:    item.Username,
		Nickname:    item.Nickname,
		IsSuper:     item.IsSuper == 1,
		Status:      item.Status,
		LastLoginAt: NullTimePtr(item.LastLoginAt),
		Roles:       roles,
	}
}

func buildMenuTree(parentID int64, childrenByParent map[int64][]types.MenuItem) []types.MenuItem {
	children := childrenByParent[parentID]
	for index := range children {
		children[index].Children = buildMenuTree(children[index].ID, childrenByParent)
	}

	return children
}
