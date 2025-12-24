package main

import (
	"log"

	"github.com/yuhang1130/gin-server/internal/app"
)

func main() {
	// 初始化应用
	application, err := app.NewApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// 运行应用
	if err := application.Run(); err != nil {
		log.Fatalf("Application exited with error: %v", err)
	}
}
