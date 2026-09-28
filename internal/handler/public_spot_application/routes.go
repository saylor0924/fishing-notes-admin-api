package publicspotapplication

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
		{PermissionCode: "public.spot.application.read", Route: rest.Route{Method: http.MethodGet, Path: "/public-spot-applications", Handler: listHandler(svcCtx)}},
		{PermissionCode: "public.spot.application.read", Route: rest.Route{Method: http.MethodGet, Path: "/public-spot-applications/:applicationId", Handler: detailHandler(svcCtx)}},
		{PermissionCode: "public.spot.application.approve", Route: rest.Route{Method: http.MethodPost, Path: "/public-spot-applications/:applicationId/approve", Handler: reviewHandler(svcCtx, "approved")}},
		{PermissionCode: "public.spot.application.reject", Route: rest.Route{Method: http.MethodPost, Path: "/public-spot-applications/:applicationId/reject", Handler: reviewHandler(svcCtx, "rejected")}},
	})
}

func listHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListPublicSpotApplicationsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid public spot application query"))
			return
		}
		resp, err := svcCtx.BusinessService.ListPublicSpotApplications(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func detailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublicSpotApplicationPathRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid application id"))
			return
		}
		resp, err := svcCtx.BusinessService.GetPublicSpotApplication(r.Context(), req.ApplicationID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func reviewHandler(svcCtx *svc.ServiceContext, status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.PublicSpotApplicationPathRequest
		if err := httpx.Parse(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid application id"))
			return
		}
		var req types.ReviewPublicSpotApplicationRequest
		if err := handlerx.ParseOptionalJSONBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid review payload"))
			return
		}
		if err := svcCtx.BusinessService.ReviewPublicSpotApplication(r.Context(), pathReq.ApplicationID, status, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}
