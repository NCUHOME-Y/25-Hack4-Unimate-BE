package main

import (
	"Heckweek/internal/app/handler"
	"Heckweek/internal/app/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	repository.DBconnect()

	r := gin.Default()
	handler.BasicFlag(r)
	r.Run(":8080") // 默认监听并在 0.0.0.0:8083 上启动服务
}
