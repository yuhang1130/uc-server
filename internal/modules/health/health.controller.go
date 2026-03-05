package health

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/pkg/response"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (h *HealthController) Check(c *gin.Context) {
	response.WriteSuccess(c, gin.H{
		"status":  "ok",
		"message": "success",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}
