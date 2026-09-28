package dashboard

import (
	"net/http"

	"fishing-notes-admin-api/internal/handlerx"
	"fishing-notes-admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterRoutes(server *rest.Server, svcCtx *svc.ServiceContext) {
	handlerx.AddPermissionRoutes(server, svcCtx, []handlerx.PermissionRoute{
		{PermissionCode: "dashboard.read", Route: rest.Route{Method: http.MethodGet, Path: "/dashboard/overview", Handler: overviewHandler(svcCtx)}},
	})
}

func overviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := svcCtx.BusinessService.DashboardOverview(r.Context())
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, result)
	}
}
