// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import "github.com/zeromicro/go-zero/rest"

type DatabaseConf struct {
	Driver      string
	DSN         string
	MaxOpenConn int
	MaxIdleConn int
}

type AuthConf struct {
	AccessSecret string
	AccessExpire int64
}

// PermissionConf 控制权限定义的初始化行为。
type PermissionConf struct {
	SyncOnStartup bool // SyncOnStartup 仅在显式开启时于启动阶段同步内置权限。
}

type Config struct {
	rest.RestConf
	Database   DatabaseConf
	Auth       AuthConf
	Permission PermissionConf
}
