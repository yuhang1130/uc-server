package main

import (
	"log"

	"github.com/yuhang1130/gin-server/config"
	"github.com/yuhang1130/gin-server/internal/app"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app.Run(cfg)
}
