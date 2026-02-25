package app

import (
	goredis "github.com/go-redis/redis/v8"
	"github.com/yuhang1130/gin-server/internal/middleware"
	"github.com/yuhang1130/gin-server/internal/pkg/cache"
	"github.com/yuhang1130/gin-server/internal/pkg/jwt"
	"github.com/yuhang1130/gin-server/internal/service"
)

// 确保编译期 AppContextAdapter 实现了 middleware.AppContextProvider 接口
var _ middleware.AppContextProvider = (*AppContextAdapter)(nil)

// AppContextAdapter 实现 middleware.AppContextProvider 接口
type AppContextAdapter struct {
	jwtUtil     *jwt.JWTUtil
	userService service.UserService
	redisClient *goredis.Client
	userCache   *cache.UserCache
	authCache   *cache.AuthCache
	authService service.AuthService
}

// provideAppContextAdapter 将 fx 容器中的各个独立依赖聚合为 AppContextAdapter，
// 使其作为 middleware.AppContextProvider 接口的实现注入到路由中间件。
// 这是 fx 的标准聚合模式：多个独立 provider → 一个实现了接口的结构体。
func provideAppContextAdapter(
	jwtUtil *jwt.JWTUtil,
	userService service.UserService,
	redisClient *goredis.Client,
	userCache *cache.UserCache,
	authCache *cache.AuthCache,
	authService service.AuthService,
) *AppContextAdapter {
	return &AppContextAdapter{
		jwtUtil:     jwtUtil,
		userService: userService,
		redisClient: redisClient,
		userCache:   userCache,
		authCache:   authCache,
		authService: authService,
	}
}

func (a *AppContextAdapter) GetJWTUtil() *jwt.JWTUtil            { return a.jwtUtil }
func (a *AppContextAdapter) GetUserService() service.UserService { return a.userService }
func (a *AppContextAdapter) GetRedisClient() *goredis.Client     { return a.redisClient }
func (a *AppContextAdapter) GetUserCache() *cache.UserCache      { return a.userCache }
func (a *AppContextAdapter) GetAuthCache() *cache.AuthCache      { return a.authCache }
func (a *AppContextAdapter) GetAuthService() service.AuthService { return a.authService }
