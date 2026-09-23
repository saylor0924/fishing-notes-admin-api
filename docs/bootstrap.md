# 测试环境初始化

管理端登录依赖共享数据库中的 `admin_*` 表。首次部署或测试账号无法登录时，使用测试数据库执行：

```sh
mysql -h <host> -P <port> -u <user> -p <database> < scripts/bootstrap_admin.sql
```

该脚本会创建管理端权限表和 `spot_corrections` 测试表，并将 `admin` 重置为密码 `123456`、状态设为启用、设为超级管理员。脚本只能用于本地或测试环境；生产环境不得保留默认密码。正式环境的 `spot_corrections` 应通过小程序 API 的 `000042` migration 创建。
