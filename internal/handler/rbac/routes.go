package rbac

import (
	"net/http"

	"fishing-notes-admin-api/internal/errorsx"
	"fishing-notes-admin-api/internal/middleware"
	"fishing-notes-admin-api/internal/svc"
	"fishing-notes-admin-api/internal/types"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterRoutes(server *rest.Server, svcCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares([]rest.Middleware{
			middleware.RequirePermission(svcCtx, "rbac.role.read"),
		}, rest.Route{
			Method:  http.MethodGet,
			Path:    "/rbac/roles",
			Handler: listRolesHandler(svcCtx),
		}),
		rest.WithPrefix("/api/v1"),
		rest.WithJwt(svcCtx.Config.Auth.AccessSecret),
	)

	server.AddRoutes(
		rest.WithMiddlewares([]rest.Middleware{
			middleware.RequirePermission(svcCtx, "rbac.role.create"),
		}, rest.Route{
			Method:  http.MethodPost,
			Path:    "/rbac/roles",
			Handler: createRoleHandler(svcCtx),
		}),
		rest.WithPrefix("/api/v1"),
		rest.WithJwt(svcCtx.Config.Auth.AccessSecret),
	)

	server.AddRoutes(
		rest.WithMiddlewares([]rest.Middleware{
			middleware.RequirePermission(svcCtx, "rbac.role.update"),
		}, rest.Route{
			Method:  http.MethodPut,
			Path:    "/rbac/roles/:roleId",
			Handler: updateRoleHandler(svcCtx),
		}),
		rest.WithPrefix("/api/v1"),
		rest.WithJwt(svcCtx.Config.Auth.AccessSecret),
	)

	server.AddRoutes(
		rest.WithMiddlewares([]rest.Middleware{
			middleware.RequirePermission(svcCtx, "rbac.role.delete"),
		}, rest.Route{
			Method:  http.MethodDelete,
			Path:    "/rbac/roles/:roleId",
			Handler: deleteRoleHandler(svcCtx),
		}),
		rest.WithPrefix("/api/v1"),
		rest.WithJwt(svcCtx.Config.Auth.AccessSecret),
	)

	server.AddRoutes(
		rest.WithMiddlewares([]rest.Middleware{
			middleware.RequirePermission(svcCtx, "rbac.permission.read"),
		}, rest.Route{
			Method:  http.MethodGet,
			Path:    "/rbac/permissions",
			Handler: listPermissionsHandler(svcCtx),
		}),
		rest.WithPrefix("/api/v1"),
		rest.WithJwt(svcCtx.Config.Auth.AccessSecret),
	)

	server.AddRoutes(
		rest.WithMiddlewares([]rest.Middleware{
			middleware.RequirePermission(svcCtx, "rbac.user.read"),
		}, rest.Route{
			Method:  http.MethodGet,
			Path:    "/rbac/users/:userId/roles",
			Handler: userRolesHandler(svcCtx),
		}),
		rest.WithPrefix("/api/v1"),
		rest.WithJwt(svcCtx.Config.Auth.AccessSecret),
	)

	server.AddRoutes(
		rest.WithMiddlewares([]rest.Middleware{
			middleware.RequirePermission(svcCtx, "rbac.user.update"),
		}, rest.Route{
			Method:  http.MethodPut,
			Path:    "/rbac/users/:userId/roles",
			Handler: updateUserRolesHandler(svcCtx),
		}),
		rest.WithPrefix("/api/v1"),
		rest.WithJwt(svcCtx.Config.Auth.AccessSecret),
	)

	server.AddRoutes(
		rest.WithMiddlewares([]rest.Middleware{
			middleware.RequirePermission(svcCtx, "rbac.role.permissions.update"),
		}, rest.Route{
			Method:  http.MethodPut,
			Path:    "/rbac/roles/:roleId/permissions",
			Handler: updateRolePermissionsHandler(svcCtx),
		}),
		rest.WithPrefix("/api/v1"),
		rest.WithJwt(svcCtx.Config.Auth.AccessSecret),
	)
}

func listRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := svcCtx.AdminService.ListRoles(r.Context())
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func listPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := svcCtx.AdminService.ListPermissions(r.Context())
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func createRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateRoleRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid role payload"))
			return
		}

		resp, err := svcCtx.AdminService.CreateRole(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func updateRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateRoleRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid role payload"))
			return
		}

		resp, err := svcCtx.AdminService.UpdateRole(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func deleteRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RolePathRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid role id"))
			return
		}

		if err := svcCtx.AdminService.DeleteRole(r.Context(), &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]bool{"success": true})
	}
}

func userRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserRolesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid user id"))
			return
		}

		resp, err := svcCtx.AdminService.UserRoles(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func updateUserRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserRolesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid user roles payload"))
			return
		}

		resp, err := svcCtx.AdminService.UpdateUserRoles(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func updateRolePermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateRolePermissionsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid role permissions payload"))
			return
		}

		resp, err := svcCtx.AdminService.UpdateRolePermissions(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
