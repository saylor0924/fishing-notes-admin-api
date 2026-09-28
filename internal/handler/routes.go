package handler

import (
	adminHandler "fishing-notes-admin-api/internal/handler/admin"
	articleHandler "fishing-notes-admin-api/internal/handler/article"
	authHandler "fishing-notes-admin-api/internal/handler/auth"
	bannerHandler "fishing-notes-admin-api/internal/handler/banner"
	correctionHandler "fishing-notes-admin-api/internal/handler/correction"
	dashboardHandler "fishing-notes-admin-api/internal/handler/dashboard"
	docsHandler "fishing-notes-admin-api/internal/handler/docs"
	mediaHandler "fishing-notes-admin-api/internal/handler/media"
	publicSpotApplicationHandler "fishing-notes-admin-api/internal/handler/public_spot_application"
	rbacHandler "fishing-notes-admin-api/internal/handler/rbac"
	reviewHandler "fishing-notes-admin-api/internal/handler/review"
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
	reviewHandler.RegisterRoutes(server, serverCtx)
	dashboardHandler.RegisterRoutes(server, serverCtx)
	adminHandler.RegisterRoutes(server, serverCtx)
	spotHandler.RegisterRoutes(server, serverCtx)
	correctionHandler.RegisterRoutes(server, serverCtx)
	mediaHandler.RegisterRoutes(server, serverCtx)
	publicSpotApplicationHandler.RegisterRoutes(server, serverCtx)
	articleHandler.RegisterRoutes(server, serverCtx)
	bannerHandler.RegisterRoutes(server, serverCtx)
	speciesHandler.RegisterRoutes(server, serverCtx)
	userHandler.RegisterRoutes(server, serverCtx)
}
