package api_router

import (
	"github.com/yuhang1130/gin-server/internal/middleware"
	"github.com/yuhang1130/gin-server/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type APIRouter struct {
	fx.Out
	// fx.Out表示这个结构体是用来提供（输出）依赖的
	// FX 会将这个结构体的字段注册到依赖容器中
	// 其他组件可以通过 fx.In 来消费这些依赖
	Public_API_V1  *gin.RouterGroup `name:"public:api:v1"`
	Private_API_V1 *gin.RouterGroup `name:"private:api:v1"`
}

type APIRouterParams struct {
	fx.In
	Public_API_V1  *gin.RouterGroup `name:"public:api:v1"`
	Private_API_V1 *gin.RouterGroup `name:"private:api:v1"`
}

func NewAPIRouter(
	jwt jwt.JWT,
	blacklistChecker middleware.TokenBlacklistChecker,
	sessionProvider middleware.SessionDataProvider,
	app *gin.Engine,
) APIRouter {
	publicApiV1 := app.Group("/api/v1")
	privateApiV1 := publicApiV1.Group("/")

	privateApiV1.Use(middleware.APIAuthHandler(jwt, blacklistChecker, sessionProvider))

	return APIRouter{
		Public_API_V1:  publicApiV1,
		Private_API_V1: privateApiV1,
	}
}
