package response

// 通用成功状态码
const (
	SuccessCode = 0 // 成功
)

// 通用错误状态码 (1000-1999)
const (
	UnknownError         = 1000 // 未知错误
	BadRequest           = 1001 // 请求错误
	Unauthorized         = 1002 // 未授权
	Forbidden            = 1003 // 禁止访问
	NotFound             = 1004 // 未找到
	MethodNotAllowed     = 1005 // 方法不允许
	RequestTimeout       = 1006 // 请求超时
	TooManyRequests      = 1007 // 请求过于频繁
	PayloadTooLarge      = 1008 // 请求实体过大
	UnsupportedMediaType = 1009 // 不支持的媒体类型
	RateLimit            = 1010 // 频率限制
	InternalServerError  = 1011 // 内部服务器错误
	NotImplemented       = 1012 // 未实现
	ServiceUnavailable   = 1013 // 服务不可用
	GatewayTimeout       = 1014 // 网关超时
)

// 用户相关错误状态码 (2000-2999)
const (
	UserNotFound               = 2001 // 用户不存在
	UserAlreadyExists          = 2002 // 用户已存在
	UserPasswordIncorrect      = 2003 // 用户密码错误
	UserAccountLocked          = 2004 // 账户被锁定
	UserAccountDisabled        = 2005 // 账户被禁用
	UserAccountExpired         = 2006 // 账户已过期
	UserPasswordChangeRequired = 2007 // 需要更改密码
	UserInvalidToken           = 2008 // 无效的令牌
	UserTokenExpired           = 2009 // 令牌已过期
	UserInsufficientPrivilege  = 2010 // 权限不足
	UserTenantMismatch         = 2011 // 租户不匹配
)

// 认证相关错误状态码 (3000-3999)
const (
	AuthLoginFailed           = 3001 // 登录失败
	AuthRegisterFailed        = 3002 // 注册失败
	AuthLogoutFailed          = 3003 // 登出失败
	AuthInvalidCredentials    = 3004 // 无效的凭据
	AuthMissingToken          = 3005 // 缺少令牌
	AuthInvalidTokenFormat    = 3006 // 无效的令牌格式
	AuthTokenRefreshFailed    = 3007 // 令牌刷新失败
	AuthPasswordComplexity    = 3008 // 密码复杂度不符合要求
	AuthTokenInvalidSignature = 3009 // 令牌签名无效
	AuthTwoFactorRequired     = 3010 // 需要二次验证
)

// 租户相关错误状态码 (4000-4999)
const (
	TenantNotFound      = 4001 // 租户不存在
	TenantAlreadyExists = 4002 // 租户已存在
	TenantInactive      = 4003 // 租户未激活
	TenantSuspended     = 4004 // 租户被暂停
	TenantUserNotFound  = 4005 // 租户用户不存在
	TenantUserConflict  = 4006 // 租户用户冲突
)

// 数据库相关错误状态码 (5000-5999)
const (
	DatabaseOperationFailed = 5001 // 数据库操作失败
	DatabaseConnectionError = 5002 // 数据库连接错误
	DatabaseDuplicateEntry  = 5003 // 数据库重复条目
	DatabaseRecordNotFound  = 5004 // 数据库记录未找到
	DatabaseConstraintError = 5005 // 数据库约束错误
	DatabaseQueryTimeout    = 5006 // 数据库查询超时
)

// 缓存相关错误状态码 (6000-6999)
const (
	CacheOperationFailed = 6001 // 缓存操作失败
	CacheConnectionError = 6002 // 缓存连接错误
	CacheKeyNotFound     = 6003 // 缓存键未找到
	CacheKeyExpired      = 6004 // 缓存键已过期
	CacheSerialization   = 6005 // 缓存序列化错误
)

// 验证相关错误状态码 (7000-7999)
const (
	ValidationError           = 7001 // 验证错误
	ValidationFieldRequired   = 7002 // 必填字段缺失
	ValidationFieldTypeError  = 7003 // 字段类型错误
	ValidationFieldValueError = 7004 // 字段值错误
	ValidationFieldFormat     = 7005 // 字段格式错误
)

// 业务逻辑错误状态码 (8000-8999)
const (
	BusinessRuleViolation     = 8001 // 违反业务规则
	BusinessResourceExhausted = 8002 // 资源耗尽
	BusinessOperationInvalid  = 8003 // 操作无效
)

// 第三方服务错误状态码 (9000-9999)
const (
	ThirdPartyServiceError = 9001 // 第三方服务错误
	ThirdPartyAPITimeout   = 9002 // 第三方API超时
	ThirdPartyAPIError     = 9003 // 第三方API错误
)
