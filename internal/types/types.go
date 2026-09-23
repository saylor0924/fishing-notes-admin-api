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

type UserRolesResponse struct {
	UserID int64      `json:"userId"`
	Roles  []RoleItem `json:"roles"`
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
