-- 仅补充新增内置权限，不创建业务表、不重置管理员密码。
-- 适用于已有环境，执行后可在后台角色页面将权限分配给其他角色。
INSERT INTO admin_permissions (parent_id, code, name, type, path, method, description, status, sort)
VALUES
    (0, 'dashboard.read', '查看运营概览', 'api', '/api/v1/dashboard/overview', 'GET', 'Read aggregated operational dashboard metrics', 1, 2),
    (0, 'review.task.export', '导出操作日志', 'api', '/api/v1/audit-logs/export', 'GET', 'Export filtered admin audit logs as CSV', 1, 129),
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

-- 保证超级管理员可使用新增能力；其他角色请通过后台角色页面按需授权。
INSERT IGNORE INTO admin_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM admin_roles r
INNER JOIN admin_permissions p ON p.code IN ('dashboard.read', 'review.task.export', 'public.spot.application.read', 'public.spot.application.approve', 'public.spot.application.reject', 'banner.read', 'banner.create', 'banner.update', 'banner.visibility.update')
WHERE r.code = 'super_admin';
