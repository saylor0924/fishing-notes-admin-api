package admin

import (
	"net/http"

	"fishing-notes-admin-api/internal/errorsx"
	"fishing-notes-admin-api/internal/handlerx"
	"fishing-notes-admin-api/internal/svc"
	"fishing-notes-admin-api/internal/types"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterRoutes(server *rest.Server, svcCtx *svc.ServiceContext) {
	handlerx.AddPermissionRoutes(server, svcCtx, []handlerx.PermissionRoute{
		{
			PermissionCode: "rbac.user.read",
			Route: rest.Route{
				Method:  http.MethodGet,
				Path:    "/admin/users",
				Handler: listUsersHandler(svcCtx),
			},
		},
		{
			PermissionCode: "admin.user.create",
			Route: rest.Route{
				Method:  http.MethodPost,
				Path:    "/admin/users",
				Handler: createUserHandler(svcCtx),
			},
		},
		{
			PermissionCode: "admin.user.update",
			Route: rest.Route{
				Method:  http.MethodPut,
				Path:    "/admin/users/:userId",
				Handler: updateUserHandler(svcCtx),
			},
		},
		{
			PermissionCode: "admin.user.password.update",
			Route: rest.Route{
				Method:  http.MethodPost,
				Path:    "/admin/users/:userId/password",
				Handler: resetPasswordHandler(svcCtx),
			},
		},
	})
}

func listUsersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListAdminUsersRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid list users query"))
			return
		}

		resp, err := svcCtx.AdminService.ListUsers(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func createUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateAdminUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid admin user payload"))
			return
		}

		resp, err := svcCtx.AdminService.CreateAdminUser(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func updateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateAdminUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid admin user payload"))
			return
		}

		resp, err := svcCtx.AdminService.UpdateAdminUser(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func resetPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ResetAdminPasswordRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid password payload"))
			return
		}

		if err := svcCtx.AdminService.ResetAdminPassword(r.Context(), &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]bool{"success": true})
	}
}
