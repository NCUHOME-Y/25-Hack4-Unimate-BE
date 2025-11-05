package handler

import (
	"Heckweek/internal/app/service"

	"github.com/gin-gonic/gin"
)

func BasicFlag(r *gin.Engine) {
	r.POST("/register", service.RegisterUser())
	r.POST("/login", service.LoginUser())
	e := r.Group("/logined", service.JWTAuth())
	e.POST("/update-password", service.UpdateUserPassword())
	e.POST("/update-username", service.UpdateUserName())
	e.POST("/add-flag", service.PostUserFlags())
	e.GET("/get-user-flags", service.GetUserFlags())
}
