package main

import (
	"go.uber.org/fx"

	"github.com/yuhang1130/gin-server/internal/app"
)

func main() {
	fx.New(app.Module).Run()
}
