package spot

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
			PermissionCode: "spot.review.read",
			Route: rest.Route{
				Method:  http.MethodGet,
				Path:    "/spots/reviews",
				Handler: listSpotReviewsHandler(svcCtx),
			},
		},
		{
			PermissionCode: "spot.review.read",
			Route: rest.Route{
				Method:  http.MethodGet,
				Path:    "/spots/reviews/:spotId",
				Handler: spotReviewDetailHandler(svcCtx),
			},
		},
		{
			PermissionCode: "spot.review.approve",
			Route: rest.Route{
				Method:  http.MethodPost,
				Path:    "/spots/reviews/:spotId/approve",
				Handler: approveSpotReviewHandler(svcCtx),
			},
		},
		{
			PermissionCode: "spot.review.reject",
			Route: rest.Route{
				Method:  http.MethodPost,
				Path:    "/spots/reviews/:spotId/reject",
				Handler: rejectSpotReviewHandler(svcCtx),
			},
		},
	})
}

func listSpotReviewsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListFishingSpotReviewsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid fishing spot review query"))
			return
		}

		resp, err := svcCtx.BusinessService.ListFishingSpotReviews(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func spotReviewDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SpotPathRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid spot id"))
			return
		}

		resp, err := svcCtx.BusinessService.FishingSpotReviewDetail(r.Context(), req.SpotID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func approveSpotReviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.SpotPathRequest
		if err := httpx.Parse(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid spot id"))
			return
		}

		var req types.ApproveFishingSpotRequest
		if err := handlerx.ParseOptionalJSONBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid approve payload"))
			return
		}

		if err := svcCtx.BusinessService.ApproveFishingSpot(r.Context(), pathReq.SpotID, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}

func rejectSpotReviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.SpotPathRequest
		if err := httpx.Parse(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid spot id"))
			return
		}

		var req types.RejectFishingSpotRequest
		if err := handlerx.ParseOptionalJSONBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid reject payload"))
			return
		}

		if err := svcCtx.BusinessService.RejectFishingSpot(r.Context(), pathReq.SpotID, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}
