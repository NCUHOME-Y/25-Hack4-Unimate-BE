package handler

import (
	"Heckweek/internal/app/service"

	"github.com/gin-gonic/gin"
)

func BasicUser(r *gin.Engine) {
	r.POST("/api/register", service.RegisterUser())
	r.POST("/api/login", service.LoginUser())
	e := r.Use(service.JWTAuth())
	e.POST("/updatePassword", service.UpdateUserPassword())
	e.POST("/updateUsername", service.UpdateUserName())
	e.POST("/api/UpdateStatus", service.UpdateStatus())
	e.GET("/api/getUser", service.GetUser())
}

func Flag(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.POST("/api/addFlag", service.PostUserFlags())
	e.GET("/api/getUserFlags", service.GetUserFlags())
	e.POST("/api/doneFlag", service.DoneUserFlags())
	e.POST("/api/finshDoneFlag", service.FinshDoneFlag())
	e.DELETE("/api/deleteFlag", service.DeleteUserFlags())
}

func BasicPost(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.POST("/api/postUserPost", service.PostUserPost())
	e.DELETE("/api/deleteUserPost", service.DeleteUserPost())
	e.POST("/api/commentOnPost", service.CommentOnPost())
	e.DELETE("/api/deleteComment", service.DeleteUserPostComment())
}
