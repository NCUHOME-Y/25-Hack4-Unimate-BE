package handler

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/service"

	"github.com/gin-gonic/gin"
)

func BasicUser(r *gin.Engine) {
	// 公开接口：不需要认证
	r.POST("/api/register", service.RegisterUser())                     // 第一步：发送验证码
	r.POST("/api/completeRegistration", service.CompleteRegistration()) // 第二步：验证验证码并完成注册
	r.POST("/api/login", service.LoginUser())
	r.POST("/api/sendEmailCode", service.SendEmailCode()) // 修复：发送验证码
	r.POST("/api/verifyEmail", service.VerifyEmail())     // 新增：验证邮箱验证码
	r.POST("/api/loginWithOTP", service.LoginWithOTP())   // 新增：验证码登录
	r.POST("/api/forgetcode", service.ForgetPassword())

	// 需要认证的接口：创建路由组而不是污染全局路由器
	e := r.Group("/")
	e.Use(service.JWTAuth())
	e.PUT("/api/updatePassword", service.UpdateUserPassword())
	// 统一加上 /api 前缀，方便前端与 Nginx 代理规则一致
	e.PUT("/api/updateUsername", service.UpdateUserName())
	// 兼容历史大小写：保留 /api/UpdateStatus，同时提供 /api/updateStatus
	e.PUT("/api/UpdateStatus", service.UpdateStatus())
	e.PUT("/api/updateStatus", service.UpdateStatus())
	e.GET("/api/getUser", service.GetUser())
	e.GET("/api/getTodayPoints", service.GetTodayPoints())
	e.POST("/api/switchhead", service.SwitchHead())
	e.PUT("/api/updateDaka", service.DoDaKa())
	e.PUT("/api/updateRemindTime", service.UpdateUserRemindTime())
	e.PUT("/api/updateRemindStatus", service.UpdateUserRemind())
	e.PUT("/api/updateFlagRemindStatus", service.UpdateUserFlagRemind())
	e.GET("/api/getDakaRecords", service.GetDaKaRecords())
	e.PUT("/api/addPoints", service.AddPointsHandler())
	e.GET("/api/getUserStats", service.GetUserStats())
	// 管理接口：手动触发 flag_number 对齐（需要登录）
	e.POST("/api/admin/recalc_flag_numbers", service.RecalcFlagNumbers())
}

func Flag(r *gin.Engine) {
	// 公开接口：不需要认证
	r.GET("/api/getRecentDoFlagUsers", service.GetRecentDoFlagUsers())

	// 需要认证的接口：创建路由组
	e := r.Group("/")
	e.Use(service.JWTAuth())
	e.POST("/api/addFlag", service.PostUserFlags())
	e.GET("/api/getUserFlags", service.GetUserFlags())
	e.PUT("/api/updateFlagHide", service.UpdateFlagHide())
	e.PUT("/api/updateFlag", service.UpdateFlagInfo())
	e.PUT("/api/doneFlag", service.DoneUserFlags())
	e.POST("/api/finshDoneFlag", service.FinshDoneFlag())
	e.DELETE("/api/deleteFlag", service.DeleteUserFlags())
	e.GET("/api/getDoneFlags", service.GetDoneFlags())
	e.GET("/api/getUnDoneFlags", service.GetNotDoneFlags())
}

func BasicFlag(r *gin.Engine) {
	// 所有接口需要认证：创建路由组
	e := r.Group("/")
	e.Use(service.JWTAuth())
	e.POST("/api/likeFlag", service.LikeFlag())
	e.POST("/api/flagcomment", service.CommentOnFlag())
	e.DELETE("/api/flagdeletecomment", service.DeleteFlagComment())
	e.GET("/api/getflaglike", service.GetFlagLikes())

	// 新增接口：获取有日期的flag（用于日历高亮）
	e.GET("/api/flags/with-dates", service.GetFlagsWithDates())
	// 新增接口：获取预设flag
	e.GET("/api/flags/preset", service.GetPresetFlags())
	// 新增接口：获取过期flag
	e.GET("/api/flags/expired", service.GetExpiredFlags())
	// 新增接口：切换flag提醒状态
	e.POST("/api/toggleFlagNotification", service.ToggleFlagNotification())
}
func BasicPost(r *gin.Engine) {
	// 公开接口：不需要认证
	r.GET("/api/getAllPosts", service.GetAllPosts())
	r.GET("/api/getflag", service.GetVisibleFlags())
	r.POST("/api/triggerHitokoto", service.TriggerHitokoto()) // 测试接口：手动触发一言发布

	// 需要认证的接口：创建路由组
	e := r.Group("/")
	e.Use(service.JWTAuth())
	e.POST("/api/likepost", service.LikePost())
	e.GET("/api/getpostlike", service.GetPostLikes())
	e.GET("/api/getUserLikedPosts", service.GetUserLikedPosts())
	e.POST("/api/postUserPost", service.PostUserPost())
	e.DELETE("/api/deleteUserPost", service.DeleteUserPost())
	e.POST("/api/commentOnPost", service.CommentOnPost())
	e.DELETE("/api/deleteComment", service.DeleteUserPostComment())
}

func ChatWebSocket(r *gin.Engine) {
	// 所有接口需要认证：创建路由组
	e := r.Group("/")
	e.Use(service.JWTAuth())
	e.GET("/ws/chat", service.WsHandler())

	// 谈玄斋管理接口（修复：添加认证）
	e.GET("/api/chat/rooms", service.GetChatRooms())
	e.POST("/api/chat/rooms", service.CreateChatRoom())
	e.DELETE("/api/chat/rooms/:roomId", service.DeleteChatRoom())

	// 聊天历史接口
	e.GET("/api/chat/history/:roomId", service.GetChatHistory())
	e.GET("/api/private-chat/history", service.GetPrivateChatHistory())
	e.GET("/api/private-chat/conversations", service.GetPrivateConversations())
}

func Ranking(r *gin.Engine) {
	// 封神榜应该是公开的，所有人都能看
	r.GET("/api/getUseflagrRank", service.GetUserByFlagNumber())
	r.GET("/api/countranking", service.GetUserCount())
	r.GET("/api/learnTimeRanking", service.GetUserMonthLearnTime())
	r.GET("/api/dakaRanking", service.GetUserTotalDaka())
}

func LearnTime(r *gin.Engine) {
	// 所有接口需要认证：创建路由组
	e := r.Group("/")
	e.Use(service.JWTAuth())
	e.POST("/api/addLearnTime", service.RecordLearnTime())
	e.GET("/api/getlabel", service.GetLabelByUserID())
	e.GET("/api/getLearnTimemonth", service.GetLearnTimeRecords())
	e.GET("/api/getdakatotal", service.GetUserDakaTotal())
	e.GET("/api/getmonthdaka", service.GetUserMonthDaka())
	e.GET("/api/get7daylearntime", service.GetLearnTimeLast7Days())
	e.GET("/api/getLearnTime180days", service.GetLearnTimeLast180Days())
	e.GET("/api/getLearnTimemonly", service.GetLearnTimeRecordsMonth())
	// 新增接口
	e.GET("/api/getCurrentMonthLearnTime", service.GetCurrentMonthLearnTime())
	e.GET("/api/getRecent6MonthsLearnTime", service.GetRecent6MonthsLearnTime())
	// 🔧 新增：获取今日学习时长
	e.GET("/api/getTodayLearnTime", service.GetTodayLearnTime())
}

func Achievement(r *gin.Engine) {
	// 所有接口需要认证：创建路由组
	e := r.Group("/")
	e.Use(service.JWTAuth())
	e.GET("/api/getUserAchievement", service.GetUserAchievement())
}

func Search(r *gin.Engine) {
	// 所有接口需要认证：创建路由组
	e := r.Group("/")
	e.Use(service.JWTAuth())
	e.POST("/api/searchUser", service.SearchUser())
	e.POST("/api/searchPosts", service.SearchPosts())
}

// AI 学习计划路由
func AI(r *gin.Engine) {
	// 所有接口需要认证：创建路由组
	e := r.Group("/")
	e.Use(service.JWTAuth())
	e.POST("/api/ai/generate-plan", service.GenerateLearningPlan)
	e.POST("/api/ai/save-history", service.SaveAIHistory())
	e.GET("/api/ai/get-history", service.GetAIHistory())
	e.DELETE("/api/ai/delete-history", service.DeleteAIHistory())
}
