package handler

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/service"

	"github.com/gin-gonic/gin"
)

func BasicUser(r *gin.Engine) {
	r.POST("/api/register", service.RegisterUser())
	r.POST("/api/login", service.LoginUser())
	r.POST("/api/sendEmailCode", service.VerifyEmail())
	r.POST("/api/forgetcode", service.ForgetPassword())
	e := r.Use(service.JWTAuth())
	e.PUT("/updatePassword", service.UpdateUserPassword())
	e.PUT("/updateUsername", service.UpdateUserName())
	e.PUT("/api/UpdateStatus", service.UpdateStatus())
	e.GET("/api/getUser", service.GetUser())
	e.PUT("/api/updateDaka", service.DoDaKa())
	e.PUT("/api/updateRemindTime", service.UpdateUserRemindTime())
	e.PUT("/api/updateRemindStatus", service.UpdateUserRemind())
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

func BasicFlag(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.POST("/api/likeFlag", service.LikeFlag())
	e.POST("/api/flagcomment", service.CommentOnFlag())
	e.DELETE("/api/flagdeletecomment", service.DeleteFlagComment())
	e.GET("/api/getflaglike", service.GetFlagLikes())
}
func BasicPost(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.POST("/api/likepost", service.LikePost())
	e.GET("/api/getpostlike", service.GetPostLikes())
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
	e.GET("/api/countranking", service.GetUserCount())
	e.GET("/api/learnTimeRanking", service.GetUserCount())
	e.GET("/api/dakaRanking", service.GetUserTotalDaka())
}

func LearnTime(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.POST("/api/addLearnTime", service.RecordLearnTime())
	e.GET("/api/getLearnTimemonth", service.GetLearnTimeRecords())
	e.GET("/api/getdakatotal", service.GetUserDakaTotal())
	e.GET("/api/getmonthdaka", service.GetUserMonthDaka())
	e.GET("/api/get7daylearntime", service.GetLearnTimeLast7Days())
	e.GET("/api/getLearnTime180days", service.GetLearnTimeLast180Days())
	e.GET("/api/getLearnTimemonly", service.GetLearnTimeRecordsMonth())
}

func Achievement(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.GET("/api/getUserAchievement", service.GetUserAchievement())
}

func Search(r *gin.Engine) {
	e := r.Use(service.JWTAuth())
	e.POST("/api/searchUser", service.SearchUser())
	e.POST("/api/searchPosts", service.SearchPosts())
}
