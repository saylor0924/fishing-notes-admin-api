package banner

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
		{PermissionCode: "banner.read", Route: rest.Route{Method: http.MethodGet, Path: "/banners", Handler: listHandler(svcCtx)}},
		{PermissionCode: "banner.read", Route: rest.Route{Method: http.MethodGet, Path: "/banners/:bannerId", Handler: detailHandler(svcCtx)}},
		{PermissionCode: "banner.create", Route: rest.Route{Method: http.MethodPost, Path: "/banners", Handler: createHandler(svcCtx)}},
		{PermissionCode: "banner.update", Route: rest.Route{Method: http.MethodPut, Path: "/banners/:bannerId", Handler: updateHandler(svcCtx)}},
		{PermissionCode: "banner.visibility.update", Route: rest.Route{Method: http.MethodPost, Path: "/banners/:bannerId/visibility", Handler: visibilityHandler(svcCtx)}},
	})
}

func listHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListBannersRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid banner query"))
			return
		}
		resp, err := svcCtx.BusinessService.ListBanners(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func detailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BannerPathRequest
		if err := httpx.ParsePath(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid banner id"))
			return
		}
		resp, err := svcCtx.BusinessService.BannerDetail(r.Context(), req.BannerID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func createHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BannerUpsertRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid banner payload"))
			return
		}
		resp, err := svcCtx.BusinessService.CreateBanner(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func updateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.BannerPathRequest
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid banner id"))
			return
		}
		var req types.BannerUpsertRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid banner payload"))
			return
		}
		resp, err := svcCtx.BusinessService.UpdateBanner(r.Context(), pathReq.BannerID, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func visibilityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.BannerPathRequest
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid banner id"))
			return
		}
		var req types.UpdateBannerVisibilityRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid banner visibility payload"))
			return
		}
		if err := svcCtx.BusinessService.UpdateBannerVisibility(r.Context(), pathReq.BannerID, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}
