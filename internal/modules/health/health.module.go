package health

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"health",

	fx.Provide(
		NewHealthController,
	),

	fx.Invoke(func(
		engine *gin.Engine,
		healthController *HealthController,
	) {
		engine.GET("/health/ready", healthController.Check)
		engine.GET("/health/live", healthController.Check)
	}),
)
