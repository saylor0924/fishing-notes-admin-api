package correction

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
		{PermissionCode: "spot.correction.read", Route: rest.Route{Method: http.MethodGet, Path: "/spot-corrections", Handler: listHandler(svcCtx)}},
		{PermissionCode: "spot.correction.read", Route: rest.Route{Method: http.MethodGet, Path: "/spot-corrections/:correctionId", Handler: detailHandler(svcCtx)}},
		{PermissionCode: "spot.correction.review", Route: rest.Route{Method: http.MethodPost, Path: "/spot-corrections/:correctionId/approve", Handler: reviewHandler(svcCtx, "approved")}},
		{PermissionCode: "spot.correction.review", Route: rest.Route{Method: http.MethodPost, Path: "/spot-corrections/:correctionId/reject", Handler: reviewHandler(svcCtx, "rejected")}},
	})
}

func detailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.SpotCorrectionPathRequest
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid spot correction id"))
			return
		}
		resp, err := svcCtx.BusinessService.GetSpotCorrection(r.Context(), pathReq.CorrectionID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func listHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListSpotCorrectionsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid spot correction query"))
			return
		}
		resp, err := svcCtx.BusinessService.ListSpotCorrections(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func reviewHandler(svcCtx *svc.ServiceContext, status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.SpotCorrectionPathRequest
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid spot correction id"))
			return
		}
		var req types.ReviewSpotCorrectionRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid correction review payload"))
			return
		}
		if err := svcCtx.BusinessService.ReviewSpotCorrection(r.Context(), pathReq.CorrectionID, status, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}
