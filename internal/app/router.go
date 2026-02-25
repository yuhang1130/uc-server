package app

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/internal/handler"
	"github.com/yuhang1130/gin-server/internal/middleware"
	"github.com/yuhang1130/gin-server/internal/pkg/response"
)

func registerRoutesFx(
	engine *gin.Engine,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	adapter *AppContextAdapter,
) {
	engine.NoRoute(func(c *gin.Context) {
		response.NotFoundFunc(c, "Not Found")
	})

	engine.GET("/health", func(c *gin.Context) {
		response.SuccessFunc(c, gin.H{
			"status":  "ok",
			"message": "success",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	api := engine.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login",
				middleware.IPRateLimiter(adapter.GetUserCache(), middleware.DefaultIPRateLimit),
				authHandler.Login,
			)
		}

		protected := api.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(adapter))
		{
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", authHandler.Logout)
				authProtected.POST("/change-password", authHandler.ChangePassword)
			}

			userGroup := protected.Group("/user")
			{
				userGroup.GET("/me", userHandler.GetCurrentUser)
				userGroup.GET("/list", userHandler.ListUsers)
				userGroup.POST("/create", userHandler.CreateUser)
				userGroup.GET("/info", userHandler.GetUserByID)
				userGroup.POST("/update", userHandler.UpdateUser)
				userGroup.POST("/delete", userHandler.DeleteUser)
			}
		}
	}
}
