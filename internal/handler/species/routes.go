package species

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
		{PermissionCode: "species.read", Route: rest.Route{Method: http.MethodGet, Path: "/species", Handler: listSpeciesHandler(svcCtx)}},
		{PermissionCode: "species.read", Route: rest.Route{Method: http.MethodGet, Path: "/species/:speciesId", Handler: speciesDetailHandler(svcCtx)}},
		{PermissionCode: "species.create", Route: rest.Route{Method: http.MethodPost, Path: "/species", Handler: createSpeciesHandler(svcCtx)}},
		{PermissionCode: "species.update", Route: rest.Route{Method: http.MethodPut, Path: "/species/:speciesId", Handler: updateSpeciesHandler(svcCtx)}},
		{PermissionCode: "species.visibility.update", Route: rest.Route{Method: http.MethodPost, Path: "/species/:speciesId/visibility", Handler: updateSpeciesVisibilityHandler(svcCtx)}},
	})
}

func listSpeciesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListSpeciesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid species query"))
			return
		}

		resp, err := svcCtx.BusinessService.ListSpecies(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func speciesDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SpeciesPathRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid species id"))
			return
		}

		resp, err := svcCtx.BusinessService.SpeciesDetail(r.Context(), req.SpeciesID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func createSpeciesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SpeciesUpsertRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid species payload"))
			return
		}

		resp, err := svcCtx.BusinessService.CreateSpecies(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func updateSpeciesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.SpeciesPathRequest
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid species id"))
			return
		}

		var req types.SpeciesUpsertRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid species payload"))
			return
		}

		resp, err := svcCtx.BusinessService.UpdateSpecies(r.Context(), pathReq.SpeciesID, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func updateSpeciesVisibilityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.SpeciesPathRequest
		if err := httpx.Parse(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid species id"))
			return
		}

		var req types.UpdateSpeciesVisibilityRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid visibility payload"))
			return
		}

		if err := svcCtx.BusinessService.UpdateSpeciesVisibility(r.Context(), pathReq.SpeciesID, req.VisibleStatus); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}
