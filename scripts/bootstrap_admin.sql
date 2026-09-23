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

INSERT IGNORE INTO admin_user_roles (user_id, role_id)
SELECT u.id, r.id FROM admin_users u CROSS JOIN admin_roles r
WHERE u.username = 'admin' AND r.code = 'super_admin';
