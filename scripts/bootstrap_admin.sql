-- 仅用于本地和测试环境。执行后必须修改 admin 默认密码。
CREATE TABLE IF NOT EXISTS admin_users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(64) NOT NULL DEFAULT '',
    status TINYINT NOT NULL DEFAULT 1,
    is_super TINYINT NOT NULL DEFAULT 0,
    last_login_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_admin_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(64) NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    status TINYINT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_admin_roles_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_permissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    code VARCHAR(128) NOT NULL,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(32) NOT NULL,
    path VARCHAR(255) NOT NULL DEFAULT '',
    method VARCHAR(16) NOT NULL DEFAULT '',
    component VARCHAR(255) NOT NULL DEFAULT '',
    icon VARCHAR(64) NOT NULL DEFAULT '',
    description VARCHAR(255) NOT NULL DEFAULT '',
    status TINYINT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_admin_permissions_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_role_permissions (
    role_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    actor_user_id BIGINT UNSIGNED NULL DEFAULT NULL,
    action VARCHAR(64) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id BIGINT UNSIGNED NULL DEFAULT NULL,
    detail_json TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_admin_audit_actor_created (actor_user_id, created_at DESC, id DESC),
    KEY idx_admin_audit_resource_created (resource_type, resource_id, created_at DESC, id DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 与 fishing-notes-api 的 000042_create_spot_corrections_table 迁移保持一致。
-- 管理端审核列表依赖该表；正式环境仍应由小程序 API 的 migration 管理。
CREATE TABLE IF NOT EXISTS spot_corrections (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    correction_code VARCHAR(64) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    spot_id BIGINT UNSIGNED NOT NULL,
    spot_code VARCHAR(64) NOT NULL,
    spot_name VARCHAR(128) NOT NULL DEFAULT '',
    content VARCHAR(500) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    review_note VARCHAR(500) NOT NULL DEFAULT '',
    reviewer_user_id BIGINT UNSIGNED NULL DEFAULT NULL,
    reviewed_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_spot_corrections_code (correction_code),
    KEY idx_spot_corrections_user_created (user_id, created_at DESC, id DESC),
    KEY idx_spot_corrections_status_created (status, created_at ASC, id ASC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- BCrypt hash for plaintext `123456`. 仅测试环境允许重置既有 admin 密码。
INSERT INTO admin_users (username, password_hash, nickname, status, is_super)
VALUES ('admin', '$2a$10$H4hP/MzNn0gQmN2JXI9hMe8o7Saso1c/LkYnA9kMd3Hxbw7ioEWVu', '系统管理员', 1, 1)
ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), nickname = VALUES(nickname), status = 1, is_super = 1;

INSERT INTO admin_roles (code, name, description, status, sort)
VALUES ('super_admin', '超级管理员', '测试环境默认超级管理员角色', 1, 1)
ON DUPLICATE KEY UPDATE name = VALUES(name), status = 1;

INSERT INTO admin_permissions (parent_id, code, name, type, path, method, description, status, sort)
VALUES
    (0, 'admin.user.create', '创建管理员账号', 'api', '/api/v1/admin/users', 'POST', 'Create an administrator account', 1, 118),
    (0, 'admin.user.update', '编辑管理员账号', 'api', '/api/v1/admin/users/:userId', 'PUT', 'Update administrator profile and status', 1, 119),
    (0, 'admin.user.password.update', '重置管理员密码', 'api', '/api/v1/admin/users/:userId/password', 'POST', 'Reset administrator password', 1, 120),
    (0, 'rbac.user.update', '更新管理员角色', 'api', '/api/v1/rbac/users/:userId/roles', 'PUT', 'Update user role assignments', 1, 121),
    (0, 'rbac.role.permissions.update', '更新角色权限', 'api', '/api/v1/rbac/roles/:roleId/permissions', 'PUT', 'Update role permission assignments', 1, 127),
    (0, 'rbac.role.create', '创建角色', 'api', '/api/v1/rbac/roles', 'POST', 'Create a role', 1, 123),
    (0, 'rbac.role.update', '编辑角色', 'api', '/api/v1/rbac/roles/:roleId', 'PUT', 'Update a role', 1, 124),
    (0, 'rbac.role.delete', '删除角色', 'api', '/api/v1/rbac/roles/:roleId', 'DELETE', 'Delete a role', 1, 125),
    (0, 'dashboard.read', '查看运营概览', 'api', '/api/v1/dashboard/overview', 'GET', 'Read aggregated operational dashboard metrics', 1, 2),
    (0, 'review.task.read', '查看审核任务', 'api', '/api/v1/review-tasks', 'GET', 'Read unified review tasks', 1, 128),
    (0, 'review.task.export', '导出操作日志', 'api', '/api/v1/audit-logs/export', 'GET', 'Export filtered admin audit logs as CSV', 1, 129),
    (0, 'public.fishing_report.approve', '通过公开鱼情', 'api', '/api/v1/review-tasks/public-reports/:recordId/approve', 'POST', 'Approve public fishing report', 1, 130),
    (0, 'public.fishing_report.reject', '驳回公开鱼情', 'api', '/api/v1/review-tasks/public-reports/:recordId/reject', 'POST', 'Reject public fishing report', 1, 131),
    (0, 'media.asset.read', '查看图片审核队列', 'api', '/api/v1/media-assets', 'GET', 'Read public fishing record image review queue', 1, 132),
    (0, 'media.asset.approve', '通过图片审核', 'api', '/api/v1/media-assets/:mediaId/approve', 'POST', 'Approve a public fishing record image', 1, 133),
    (0, 'media.asset.reject', '驳回图片审核', 'api', '/api/v1/media-assets/:mediaId/reject', 'POST', 'Reject a public fishing record image', 1, 134),
    (0, 'public.spot.application.read', '查看公共钓点申请', 'api', '/api/v1/public-spot-applications', 'GET', 'Read public fishing spot applications', 1, 135),
    (0, 'public.spot.application.approve', '通过公共钓点申请', 'api', '/api/v1/public-spot-applications/:applicationId/approve', 'POST', 'Approve and publish or merge a public fishing spot application', 1, 136),
    (0, 'public.spot.application.reject', '驳回公共钓点申请', 'api', '/api/v1/public-spot-applications/:applicationId/reject', 'POST', 'Reject a public fishing spot application', 1, 137),
    (0, 'banner.read', '查看发现横幅', 'api', '/api/v1/banners', 'GET', 'Read discover event banners', 1, 138),
    (0, 'banner.create', '创建发现横幅', 'api', '/api/v1/banners', 'POST', 'Create a discover event banner', 1, 139),
    (0, 'banner.update', '编辑发现横幅', 'api', '/api/v1/banners/:bannerId', 'PUT', 'Update a discover event banner', 1, 140),
    (0, 'banner.visibility.update', '发布发现横幅', 'api', '/api/v1/banners/:bannerId/visibility', 'POST', 'Publish or hide a discover event banner', 1, 141)
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    path = VALUES(path),
    method = VALUES(method),
    description = VALUES(description),
    status = VALUES(status),
    sort = VALUES(sort);

INSERT IGNORE INTO admin_user_roles (user_id, role_id)
SELECT u.id, r.id FROM admin_users u CROSS JOIN admin_roles r
WHERE u.username = 'admin' AND r.code = 'super_admin';

INSERT IGNORE INTO admin_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM admin_roles r
INNER JOIN admin_permissions p ON p.code IN ('admin.user.create', 'admin.user.update', 'admin.user.password.update', 'rbac.user.update', 'rbac.role.permissions.update', 'rbac.role.create', 'rbac.role.update', 'rbac.role.delete', 'dashboard.read', 'review.task.read', 'review.task.export', 'public.fishing_report.approve', 'public.fishing_report.reject', 'media.asset.read', 'media.asset.approve', 'media.asset.reject', 'public.spot.application.read', 'public.spot.application.approve', 'public.spot.application.reject', 'banner.read', 'banner.create', 'banner.update', 'banner.visibility.update')
WHERE r.code = 'super_admin';
