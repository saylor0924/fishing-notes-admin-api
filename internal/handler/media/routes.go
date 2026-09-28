package media

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
		{PermissionCode: "media.asset.read", Route: rest.Route{Method: http.MethodGet, Path: "/media-assets", Handler: listHandler(svcCtx)}},
		{PermissionCode: "media.asset.approve", Route: rest.Route{Method: http.MethodPost, Path: "/media-assets/:mediaId/approve", Handler: reviewHandler(svcCtx, "approved")}},
		{PermissionCode: "media.asset.reject", Route: rest.Route{Method: http.MethodPost, Path: "/media-assets/:mediaId/reject", Handler: reviewHandler(svcCtx, "rejected")}},
	})
}

func listHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListMediaAssetsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid media asset query"))
			return
		}
		resp, err := svcCtx.BusinessService.ListMediaAssets(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func reviewHandler(svcCtx *svc.ServiceContext, status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.MediaAssetPathRequest
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid media asset id"))
			return
		}
		var req types.ReviewMediaAssetRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid media review payload"))
			return
		}
		if err := svcCtx.BusinessService.ReviewMediaAsset(r.Context(), pathReq.MediaID, status, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}
