package article

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
		{PermissionCode: "article.read", Route: rest.Route{Method: http.MethodGet, Path: "/articles", Handler: listArticlesHandler(svcCtx)}},
		{PermissionCode: "article.read", Route: rest.Route{Method: http.MethodGet, Path: "/articles/:articleId", Handler: articleDetailHandler(svcCtx)}},
		{PermissionCode: "article.create", Route: rest.Route{Method: http.MethodPost, Path: "/articles", Handler: createArticleHandler(svcCtx)}},
		{PermissionCode: "article.update", Route: rest.Route{Method: http.MethodPut, Path: "/articles/:articleId", Handler: updateArticleHandler(svcCtx)}},
		{PermissionCode: "article.publish", Route: rest.Route{Method: http.MethodPost, Path: "/articles/:articleId/publish", Handler: publishArticleHandler(svcCtx)}},
		{PermissionCode: "article.offline", Route: rest.Route{Method: http.MethodPost, Path: "/articles/:articleId/offline", Handler: offlineArticleHandler(svcCtx)}},
	})
}

func listArticlesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListArticlesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid article query"))
			return
		}

		resp, err := svcCtx.BusinessService.ListArticles(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func articleDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ArticlePathRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid article id"))
			return
		}

		resp, err := svcCtx.BusinessService.ArticleDetail(r.Context(), req.ArticleID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func createArticleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ArticleUpsertRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid article payload"))
			return
		}

		resp, err := svcCtx.BusinessService.CreateArticle(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func updateArticleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.ArticlePathRequest
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid article id"))
			return
		}

		var req types.ArticleUpsertRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid article payload"))
			return
		}

		resp, err := svcCtx.BusinessService.UpdateArticle(r.Context(), pathReq.ArticleID, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func publishArticleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.ArticlePathRequest
		if err := httpx.Parse(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid article id"))
			return
		}

		var req types.ArticlePublishRequest
		if err := handlerx.ParseOptionalJSONBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid publish payload"))
			return
		}

		if err := svcCtx.BusinessService.PublishArticle(r.Context(), pathReq.ArticleID, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}

func offlineArticleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ArticlePathRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid article id"))
			return
		}

		if err := svcCtx.BusinessService.OfflineArticle(r.Context(), req.ArticleID); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}
