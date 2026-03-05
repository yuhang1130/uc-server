package modules

import (
	"github.com/yuhang1130/gin-server/internal/modules/auth"
	"github.com/yuhang1130/gin-server/internal/modules/health"
	"github.com/yuhang1130/gin-server/internal/modules/user"
	"go.uber.org/fx"
)

var APIModule = fx.Module(
	"api",
	health.Module,
	auth.Module,
	user.Module,
)
