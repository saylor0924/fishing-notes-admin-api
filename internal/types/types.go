package types

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string      `json:"accessToken"`
	TokenType   string      `json:"tokenType"`
	ExpiresAt   int64       `json:"expiresAt"`
	User        CurrentUser `json:"user"`
}

type CurrentUser struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	IsSuper     bool       `json:"isSuper"`
	Status      int64      `json:"status"`
	LastLoginAt *string    `json:"lastLoginAt,omitempty"`
	Roles       []RoleItem `json:"roles"`
	Permissions []string   `json:"permissions"`
}

type PermissionCodesResponse struct {
	Permissions []string `json:"permissions"`
}

type MenusResponse struct {
	Menus []MenuItem `json:"menus"`
}

type RoleItem struct {
	ID              int64    `json:"id"`
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Status          int64    `json:"status"`
	Sort            int64    `json:"sort"`
	PermissionCodes []string `json:"permissionCodes,omitempty"`
}

type PermissionItem struct {
	ID          int64  `json:"id"`
	ParentID    int64  `json:"parentId"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Path        string `json:"path,omitempty"`
	Method      string `json:"method,omitempty"`
	Component   string `json:"component,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	Status      int64  `json:"status"`
	Sort        int64  `json:"sort"`
}

type MenuItem struct {
	ID        int64      `json:"id"`
	ParentID  int64      `json:"parentId"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Path      string     `json:"path,omitempty"`
	Component string     `json:"component,omitempty"`
	Icon      string     `json:"icon,omitempty"`
	Sort      int64      `json:"sort"`
	Children  []MenuItem `json:"children,omitempty"`
}

type ListRolesResponse struct {
	Roles []RoleItem `json:"roles"`
}

type ListPermissionsResponse struct {
	Permissions []PermissionItem `json:"permissions"`
}

type UserRolesRequest struct {
	UserID int64 `path:"userId"`
}

type UpdateUserRolesRequest struct {
	UserID  int64   `path:"userId"`  // 待修改角色的管理员 ID
	RoleIDs []int64 `json:"roleIds"` // 绑定的启用角色 ID 列表
}

type UserRolesResponse struct {
	UserID int64      `json:"userId"`
	Roles  []RoleItem `json:"roles"`
}

type UpdateRolePermissionsRequest struct {
	RoleID        int64   `path:"roleId"`        // 待修改权限的角色 ID
	PermissionIDs []int64 `json:"permissionIds"` // 绑定的启用权限 ID 列表
}

type RolePermissionsResponse struct {
	RoleID        int64   `json:"roleId"`
	PermissionIDs []int64 `json:"permissionIds"`
}

type RolePathRequest struct {
	RoleID int64 `path:"roleId"` // 待操作的角色 ID
}

type CreateRoleRequest struct {
	Code          string  `json:"code"`          // 角色编码
	Name          string  `json:"name"`          // 角色名称
	Description   string  `json:"description"`   // 角色说明
	Status        int64   `json:"status"`        // 角色状态，1 启用，0 停用
	Sort          int64   `json:"sort"`          // 角色排序值
	PermissionIDs []int64 `json:"permissionIds"` // 创建后绑定的启用权限 ID 列表
}

type UpdateRoleRequest struct {
	RoleID      int64  `path:"roleId"`      // 待修改的角色 ID
	Name        string `json:"name"`        // 角色名称
	Description string `json:"description"` // 角色说明
	Status      int64  `json:"status"`      // 角色状态，1 启用，0 停用
	Sort        int64  `json:"sort"`        // 角色排序值
}

type ListAdminUsersRequest struct {
	Page     int64  `form:"page,optional"`
	PageSize int64  `form:"pageSize,optional"`
	Username string `form:"username,optional"`
	Nickname string `form:"nickname,optional"`
	Status   string `form:"status,optional"`
}

type AdminUserItem struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	IsSuper     bool       `json:"isSuper"`
	Status      int64      `json:"status"`
	LastLoginAt *string    `json:"lastLoginAt,omitempty"`
	Roles       []RoleItem `json:"roles"`
}

type ListAdminUsersResponse struct {
	List     []AdminUserItem `json:"list"`
	Total    int64           `json:"total"`
	Page     int64           `json:"page"`
	PageSize int64           `json:"pageSize"`
}

type CreateAdminUserRequest struct {
	Username string  `json:"username"` // 管理员登录用户名
	Password string  `json:"password"` // 管理员初始密码
	Nickname string  `json:"nickname"` // 管理员显示名称
	RoleIDs  []int64 `json:"roleIds"`  // 创建后绑定的启用角色 ID 列表
}

type AdminUserPathRequest struct {
	UserID int64 `path:"userId"` // 待操作的管理员 ID
}

type UpdateAdminUserRequest struct {
	UserID   int64  `path:"userId"`   // 待修改的管理员 ID
	Nickname string `json:"nickname"` // 管理员显示名称
	Status   int64  `json:"status"`   // 管理员状态，1 启用，0 停用
}

type ResetAdminPasswordRequest struct {
	UserID   int64  `path:"userId"`   // 待重置密码的管理员 ID
	Password string `json:"password"` // 新密码
}
