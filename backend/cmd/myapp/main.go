package main

import (
	"Heckweek/internal/app/handler"
	"Heckweek/internal/app/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	repository.DBconnect()

	r := gin.Default()
	handler.BasicUser(r)
	handler.Flag(r)
	handler.BasicPost(r)
	handler.ChatWebSocket(r)
	handler.Ranking(r)
	r.Run("0.0.0.0:8080")
}
