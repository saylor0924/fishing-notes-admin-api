// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"
	"fishing-notes-admin-api/internal/config"
	"fishing-notes-admin-api/internal/permission"
	"fishing-notes-admin-api/internal/repository"
	"fishing-notes-admin-api/internal/service"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	_ "github.com/go-sql-driver/mysql"
)

type ServiceContext struct {
	Config          config.Config
	AdminService    *service.AdminService
	BusinessService *service.BusinessService
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn(c.Database.Driver, c.Database.DSN)

	if rawDB, err := conn.RawDB(); err == nil {
		if c.Database.MaxOpenConn > 0 {
			rawDB.SetMaxOpenConns(c.Database.MaxOpenConn)
		}
		if c.Database.MaxIdleConn > 0 {
			rawDB.SetMaxIdleConns(c.Database.MaxIdleConn)
		}
	}

	adminRepo := repository.NewAdminRepository(conn)
	businessRepo := repository.NewBusinessRepository(conn)
	if c.Permission.SyncOnStartup {
		if err := adminRepo.SyncBuiltinPermissions(context.Background(), permission.BuiltinDefinitions()); err != nil {
			panic(fmt.Errorf("failed to sync builtin admin permissions: %w", err))
		}
	}

	return &ServiceContext{
		Config:          c,
		AdminService:    service.NewAdminService(adminRepo, c.Auth),
		BusinessService: service.NewBusinessService(businessRepo),
	}
}
