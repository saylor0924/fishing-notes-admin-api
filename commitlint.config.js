const MAX_HEADER_LENGTH = 100

const TYPES = [
  { value: 'feat', name: 'feat:     ✨ 新增功能', emoji: '✨' },
  { value: 'fix', name: 'fix:      🐛 修复缺陷', emoji: '🐛' },
  { value: 'docs', name: 'docs:     📝 文档变更', emoji: '📝' },
  { value: 'style', name: 'style:    💄 代码格式，不影响逻辑', emoji: '💄' },
  { value: 'refactor', name: 'refactor: ♻️ 代码重构', emoji: '♻️' },
  { value: 'perf', name: 'perf:     ⚡ 性能优化', emoji: '⚡' },
  { value: 'test', name: 'test:     ✅ 新增或调整测试', emoji: '✅' },
  { value: 'build', name: 'build:    📦 构建系统或依赖变更', emoji: '📦' },
  { value: 'ci', name: 'ci:       🤖 CI 配置变更', emoji: '🤖' },
  { value: 'chore', name: 'chore:    🔧 工程杂项', emoji: '🔧' },
  { value: 'revert', name: 'revert:   ⏪ 回滚提交', emoji: '⏪' },
]

const SCOPES = [
  { value: 'admin', name: 'admin: 管理员账号与授权' },
  { value: 'auth', name: 'auth: 登录、会话和鉴权' },
  { value: 'config', name: 'config: 配置和启动参数' },
  { value: 'handler', name: 'handler: HTTP 路由与请求响应' },
  { value: 'middleware', name: 'middleware: 中间件' },
  { value: 'permission', name: 'permission: 权限目录和权限同步' },
  { value: 'repository', name: 'repository: 数据访问与查询' },
  { value: 'service', name: 'service: 业务服务' },
  { value: 'types', name: 'types: 请求响应类型' },
  { value: 'docs', name: 'docs: 文档和接口描述' },
  { value: 'build', name: 'build: 构建和依赖' },
  { value: 'ci', name: 'ci: 持续集成配置' },
  { value: 'root', name: 'root: 跨模块或项目级变更' },
]

const TYPE_VALUES = TYPES.map(({ value }) => value)
const SCOPE_VALUES = SCOPES.map(({ value }) => value)

/** @type {import('@commitlint/types').UserConfig & import('cz-git').UserConfig} */
module.exports = {
  rules: {
    'type-enum': [2, 'always', TYPE_VALUES],
    'type-empty': [2, 'never'],
    'scope-enum': [1, 'always', SCOPE_VALUES],
    'scope-case': [2, 'always', ['lower-case', 'kebab-case']],
    'scope-empty': [0],
    'subject-empty': [2, 'never'],
    'subject-full-stop': [2, 'never', ['.', '。']],
    'header-max-length': [2, 'always', MAX_HEADER_LENGTH],
    'body-leading-blank': [1, 'always'],
    'footer-leading-blank': [1, 'always'],
  },
  prompt: {
    useEmoji: true,
    emojiAlign: 'left',
    themeColorCode: '',
    messages: {
      type: '选择提交类型:',
      scope: '选择影响功能点（可多选，可跳过）:',
      customScope: '填写自定义功能点:',
      subject: '填写简短说明:',
      body: '填写详细说明，按回车跳过:',
      breaking: '列出破坏性变更，按回车跳过:',
      footerPrefixesSelect: '选择关联 Issue 类型，按回车跳过:',
      customFooterPrefix: '输入自定义 Issue 前缀:',
      footer: '填写关联 Issue，例如 #123，按回车跳过:',
      confirmCommit: '确认生成以上提交信息?',
    },
    types: TYPES,
    scopes: SCOPES,
    enableMultipleScopes: true,
    scopeEnumSeparator: ',',
    allowCustomScopes: true,
    allowEmptyScopes: true,
    customScopesAlign: 'bottom',
    emptyScopesAlias: 'none: 不填写范围',
    customScopesAlias: 'custom: 自定义功能点',
    allowBreakingChanges: ['feat', 'fix', 'refactor'],
    issuePrefixes: [
      { value: 'closed', name: 'closed: 关闭 Issue' },
      { value: 'refs', name: 'refs: 关联 Issue' },
    ],
    allowCustomIssuePrefix: true,
    allowEmptyIssuePrefix: true,
    emptyIssuePrefixAlias: 'skip: 跳过',
    customIssuePrefixAlias: 'custom: 自定义',
    maxHeaderLength: MAX_HEADER_LENGTH,
    breaklineNumber: MAX_HEADER_LENGTH,
    confirmColorize: true,
    subjectLimit: MAX_HEADER_LENGTH,
  },
}
