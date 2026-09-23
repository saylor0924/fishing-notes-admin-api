package middleware

import (
	"net/http"

	"fishing-notes-admin-api/internal/errorsx"
	"fishing-notes-admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RequirePermission(svcCtx *svc.ServiceContext, permissionCode string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			allowed, err := svcCtx.AdminService.HasPermission(r.Context(), permissionCode)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}

			if !allowed {
				httpx.ErrorCtx(r.Context(), w, errorsx.Forbidden("permission denied"))
				return
			}

			next(w, r)
		}
	}
}
