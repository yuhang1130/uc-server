package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/yuhang1130/gin-server/internal/middleware"
	"github.com/yuhang1130/gin-server/internal/pkg/response"
)

type Application struct {
	Server *http.Server
	AppCtx *AppContext
}

func NewApplication() (*Application, error) {
	// 初始化应用上下文
	if err := InitApp(); err != nil {
		return nil, err
	}

	appCtx := GetAppContext()

	// 初始化Gin引擎
	engine := setupGinEngine(appCtx)

	// 注册路由
	registerRoutes(engine, appCtx)

	server := &http.Server{
		Addr:    ":" + appCtx.Config.Server.Port,
		Handler: engine,
	}

	return &Application{
		Server: server,
		AppCtx: appCtx,
	}, nil
}

func (a *Application) Run() error {
	// 启动服务器
	go func() {
		log.Printf("Server is running on port %s", a.AppCtx.Config.Server.Port)
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 设置超时时间，优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// 关闭资源（使用 AppContext 的 Close 方法）
	if err := Close(); err != nil {
		log.Println("Error closing resources:", err)
	}

	log.Println("Server exiting")
	return nil
}

func setupGinEngine(appCtx *AppContext) *gin.Engine {
	isProd := appCtx.Config.Server.Mode == "production"
	if isProd {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	/**
		注册pprof后，可以通过以下HTTP端点访问性能数据：
		- /debug/pprof/ - 性能分析首页
		- /debug/pprof/profile - CPU性能分析
		- /debug/pprof/heap - 堆内存分析
		- /debug/pprof/goroutine - Goroutine信息
		- /debug/pprof/block - 阻塞分析
		- /debug/pprof/threadcreate - 线程创建分析
	**/
	if !isProd {
		pprof.Register(engine)
	}

	// 全局中间件
	engine.Use(
		middleware.RequestIDMiddleware(),
		middleware.BodySizeLimitMiddleware(appCtx.Config.Server.MaxBodySize),
		middleware.LoggingMiddleware(appCtx.GetZapLogger()),
		middleware.RecoveryMiddleware(appCtx.GetZapLogger()),
		middleware.CORSMiddleware(appCtx.Config.Server.CorsAllowed),
	)

	// 获取验证器实例并注册自定义验证
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 添加自定义验证规则
		v.RegisterValidation("is_mobile", validateMobile)
		// 其他验证规则...
	}

	return engine
}

func registerRoutes(r *gin.Engine, appCtx *AppContext) {
	r.NoRoute(func(c *gin.Context) {
		// 对于其他未找到的路由，继续默认处理
		response.NotFoundFunc(c, "Not Found")
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.SuccessFunc(c, gin.H{
			"status":  "ok",
			"message": "success",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// API版本路由
	api := r.Group("/api/v1")
	{
		// 认证相关路由（公开） - 使用单例 Handler
		authHandler := appCtx.GetAuthHandler()
		authGroup := api.Group("/auth")
		{
			// authGroup.POST("/register", authHandler.Register) // 用户注册

			// 身份认证
			authGroup.POST("/login",
				middleware.IPRateLimiter(appCtx.userCache, middleware.DefaultIPRateLimit),
				authHandler.Login,
			)
		}

		// 需要认证的路由
		protected := api.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(appCtx))
		{
			// 认证相关（需要登录）
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", authHandler.Logout)                  // 登出
				authProtected.POST("/change-password", authHandler.ChangePassword) // 修改密码
			}

			// 用户相关路由 - 使用单例 Handler
			userHandler := appCtx.GetUserHandler()
			userGroup := protected.Group("/user")
			{
				userGroup.POST("/me", userHandler.GetCurrentUser) // 获取当前用户
				userGroup.POST("/list", userHandler.ListUsers)    // 获取用户列表
				userGroup.POST("/create", userHandler.CreateUser) // 创建用户
				userGroup.POST("/info", userHandler.GetUserByID)  // 获取指定用户
				userGroup.POST("/update", userHandler.UpdateUser) // 更新用户
				userGroup.POST("/delete", userHandler.DeleteUser) // 删除用户
			}
		}
	}
}

func validateMobile(fl validator.FieldLevel) bool {
	// 获取字段值
	mobile := fl.Field().String()

	// 简单的手机号格式验证（以中国大陆手机号为例）
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, mobile)
	return matched
}
