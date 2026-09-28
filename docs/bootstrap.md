# 测试环境初始化

当前接口范围和后续建设计划见[管理后台 API 当前状态](project-status.md)。

管理端登录依赖共享数据库中的 `admin_*` 表。首次部署或测试账号无法登录时，使用测试数据库执行：

```sh
mysql -h <host> -P <port> -u <user> -p <database> < scripts/bootstrap_admin.sql
```

该脚本会创建管理端权限表、`admin_audit_logs` 审计表和 `spot_corrections` 测试表，补充当前内置 RBAC 权限，并将 `admin` 重置为密码 `123456`、状态设为启用、设为超级管理员。脚本只能用于本地或测试环境；生产环境不得保留默认密码。正式环境的 `spot_corrections` 应通过小程序 API 的 `000042` migration 创建，公开鱼情审核字段应先执行 `000043`，图片审核字段应执行 `000049`。

已有环境不应重复执行完整 bootstrap，也不应为补权限长期打开 `Permission.SyncOnStartup`。本次新增的 `dashboard.read` 和 `review.task.export` 权限可在维护窗口执行 `scripts/sync_permissions.sql`，该脚本不会重置管理员密码；其他角色再通过后台角色页面按需授权。
