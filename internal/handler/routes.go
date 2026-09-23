package handler

import (
	adminHandler "fishing-notes-admin-api/internal/handler/admin"
	articleHandler "fishing-notes-admin-api/internal/handler/article"
	authHandler "fishing-notes-admin-api/internal/handler/auth"
	correctionHandler "fishing-notes-admin-api/internal/handler/correction"
	docsHandler "fishing-notes-admin-api/internal/handler/docs"
	rbacHandler "fishing-notes-admin-api/internal/handler/rbac"
	speciesHandler "fishing-notes-admin-api/internal/handler/species"
	spotHandler "fishing-notes-admin-api/internal/handler/spot"
	userHandler "fishing-notes-admin-api/internal/handler/user"
	"fishing-notes-admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	docsHandler.RegisterRoutes(server, serverCtx)
	authHandler.RegisterRoutes(server, serverCtx)
	rbacHandler.RegisterRoutes(server, serverCtx)
	adminHandler.RegisterRoutes(server, serverCtx)
	spotHandler.RegisterRoutes(server, serverCtx)
	correctionHandler.RegisterRoutes(server, serverCtx)
	articleHandler.RegisterRoutes(server, serverCtx)
	speciesHandler.RegisterRoutes(server, serverCtx)
	userHandler.RegisterRoutes(server, serverCtx)
}
