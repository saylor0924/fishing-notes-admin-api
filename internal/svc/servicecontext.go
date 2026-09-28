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
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	_ "github.com/go-sql-driver/mysql"
)

type ServiceContext struct {
	Config          config.Config
	AdminService    *service.AdminService
	BusinessService *service.BusinessService
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn(c.Database.Driver, c.Database.DSN)

	rawDB, err := conn.RawDB()
	if err != nil {
		panic(fmt.Errorf("failed to initialize database connection: %w", err))
	}
	if c.Database.MaxOpenConn > 0 {
		rawDB.SetMaxOpenConns(c.Database.MaxOpenConn)
	}
	if c.Database.MaxIdleConn > 0 {
		rawDB.SetMaxIdleConns(c.Database.MaxIdleConn)
	}
	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: rawDB}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		panic(fmt.Errorf("failed to initialize GORM repositories: %w", err))
	}

	adminRepo := repository.NewAdminRepository(gormDB)
	businessRepo := repository.NewBusinessRepository(gormDB)
	auditRepo := repository.NewAuditRepository(gormDB)
	if c.Permission.SyncOnStartup {
		if err := adminRepo.SyncBuiltinPermissions(context.Background(), permission.BuiltinDefinitions()); err != nil {
			panic(fmt.Errorf("failed to sync builtin admin permissions: %w", err))
		}
	}

	return &ServiceContext{
		Config:          c,
		AdminService:    service.NewAdminService(adminRepo, c.Auth, auditRepo),
		BusinessService: service.NewBusinessService(businessRepo, auditRepo),
	}
}
