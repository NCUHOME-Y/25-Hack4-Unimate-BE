package handler

import (
	"Heckweek/internal/app/service"

	"github.com/gin-gonic/gin"
)

func BasicFlag(r *gin.Engine) {
	r.POST("/register", service.RegisterUser())
	r.POST("/login", service.LoginUser())
	r.POST("/update-password", service.UpdateUserPassword())
	r.POST("/update-username", service.UpdateUserName())
	r.GET("/get-user-flags", service.GetUserFlags())
}
