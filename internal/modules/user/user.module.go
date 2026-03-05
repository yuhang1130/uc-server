package user

import (
	api_router "github.com/yuhang1130/gin-server/internal/router/api"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"user",

	fx.Provide(
		// user cache
		NewUserCache,
		// user repository
		NewUserRepository,
		NewTenantUserRepository,
		// user service
		NewUserService,
		// user controller
		NewUserController,
	),

	// register router
	fx.Invoke(func(
		router api_router.APIRouterParams,
		userController *UserController,
	) {
		// 所有user路由都需要认证，所以放在Private API下
		userGroup := router.Private_API_V1.Group("/user")
		{
			userGroup.GET("/me", userController.GetCurrentUser)
			userGroup.GET("/list", userController.ListUsers)
			userGroup.POST("/create", userController.CreateUser)
			userGroup.GET("/info", userController.GetUserByID)
			userGroup.POST("/update", userController.UpdateUser)
			userGroup.POST("/delete", userController.DeleteUser)
		}
	}),
)
