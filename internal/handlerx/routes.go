package handlerx

import (
	"errors"
	"io"
	"net/http"

	"fishing-notes-admin-api/internal/middleware"
	"fishing-notes-admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type PermissionRoute struct {
	PermissionCode string
	Route          rest.Route
}

func AddPermissionRoutes(server *rest.Server, svcCtx *svc.ServiceContext, routes []PermissionRoute) {
	for _, item := range routes {
		server.AddRoutes(
			rest.WithMiddlewares([]rest.Middleware{
				middleware.RequirePermission(svcCtx, item.PermissionCode),
			}, item.Route),
			rest.WithPrefix("/api/v1"),
			rest.WithJwt(svcCtx.Config.Auth.AccessSecret),
		)
	}
}

func ParseOptionalJSONBody(r *http.Request, v any) error {
	err := httpx.ParseJsonBody(r, v)
	if err == nil || errors.Is(err, io.EOF) {
		return nil
	}

	return err
}
