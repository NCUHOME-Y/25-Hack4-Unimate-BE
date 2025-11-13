package handler

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/service"

	"github.com/gin-gonic/gin"
)

func BasicUser(r *gin.Engine) {
	r.POST("/api/register", service.RegisterUser())
	r.POST("/api/login", service.LoginUser())
	r.POST("/api/sendEmailCode", service.VerifyEmail())
	e := r.Use(service.JWTAuth())
	e.PUT("/updatePassword", service.UpdateUserPassword())
	e.PUT("/updateUsername", service.UpdateUserName())
	e.PUT("/api/UpdateStatus", service.UpdateStatus())
	e.GET("/api/getUser", service.GetUser())
	e.PUT("/api/updateDaka", service.DoDaKa())
	e.GET("/api/getDakaRecords", service.GetDaKaRecords())
}

func Flag(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.POST("/api/addFlag", service.PostUserFlags())
	e.GET("/api/getUserFlags", service.GetUserFlags())
	e.PUT("/api/updateFlagHide", service.UpdateFlagHide())
	e.PUT("/api/doneFlag", service.DoneUserFlags())
	e.POST("/api/finshDoneFlag", service.FinshDoneFlag())
	e.DELETE("/api/deleteFlag", service.DeleteUserFlags())
	e.GET("/api/getDoneFlags", service.GetDoneFlags())
	e.GET("/api/getUnDoneFlags", service.GetNotDoneFlags())
	r.GET("/api/getRecentDoFlagUsers", service.GetRecentDoFlagUsers())
}

func BasicPost(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.POST("/api/postUserPost", service.PostUserPost())
	e.DELETE("/api/deleteUserPost", service.DeleteUserPost())
	e.POST("/api/commentOnPost", service.CommentOnPost())
	e.DELETE("/api/deleteComment", service.DeleteUserPostComment())
	e.GET("/api/getAllPosts", service.GetAllPosts())
	e.GET("/api/getflag", service.GetVisibleFlags())
}

func ChatWebSocket(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.GET("/ws/chat", service.WsHandler())
}

func Ranking(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.GET("/api/ranking", service.GetUserCount())
}

func LearnTime(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.POST("/api/addLearnTime", service.RecordLearnTime())
	e.GET("/api/getLearnTime", service.GetLearnTimeRecords())
}

func Achievement(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.GET("/api/getUserAchievement", service.GetUserAchievement())
}
