package handler

import (
	"Heckweek/internal/app/service"

	"github.com/gin-gonic/gin"
)

func BasicFlag(r *gin.Engine) {
	r.POST("/api/register", service.RegisterUser())
	r.POST("/api/login", service.LoginUser())
	e := r.Group("/logined", service.JWTAuth())
	e.POST("/update-password", service.UpdateUserPassword())
	e.POST("/update-username", service.UpdateUserName())
	e.POST("/api/add-flag", service.PostUserFlags())
	e.GET("/api/get-user-flags", service.GetUserFlags())
	e.POST("/api/doneFlag", service.DoneUserFlags())
}
