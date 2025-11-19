package service

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 token - 支持 Authorization 头和 URL 参数（用于 WebSocket）
		var token string
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader != "" {
			// 从 Authorization 头获取
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
				log.Printf("[JWT] 从 Authorization 头获取 token")
			} else {
				log.Printf("[JWT] Authorization 格式错误: %s", authHeader)
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  "请求头中 Authorization 格式有误",
				})
				c.Abort()
				return
			}
		} else {
			// 从 URL 参数获取（用于 WebSocket 连接）
			token = c.Query("token")
			if token == "" {
				log.Printf("[JWT] 未找到 token - Authorization 头为空,URL 参数也为空")
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  "请求头中 Authorization 为空且 URL 中无 token 参数",
				})
				c.Abort()
				return
			}
			log.Printf("[JWT] 从 URL 参数获取 token: %s...", token[:min(10, len(token))])
		}

		// 解析 token
		claims, err := utils.ParseToken(token)
		if err != nil {
			log.Printf("[JWT] Token 解析失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "无效的 Token",
			})
			c.Abort()
			return
		}

		log.Printf("[JWT] Token 验证成功 - 用户ID: %d, 用户名: %s", claims.UserID, claims.Username)

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("token", token)

		c.Next()
	}
}

func getCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	// 类型断言
	id, ok := userID.(uint)
	if !ok {
		return 0, false
	}

	return id, true
}

// 用户注册
func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": "注册失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		// 检查邮箱是否已注册
		user_exist, _ := repository.GetUserByEmail(user.Email)
		if user_exist.ID != 0 {
			c.JSON(401, gin.H{"error": "该邮箱已被注册,请更换邮箱..."})
			log.Print("Email already exists")
			return
		}
		// 检查用户名是否已存在
		name_exist, _ := repository.GetUserByName(user.Name)
		if name_exist.ID != 0 {
			c.JSON(401, gin.H{"error": "该用户名已被使用,请更换用户名..."})
			log.Print("Username already exists")
			return
		}
		password, err := utils.HashPassword(user.Password)
		user.Password = password
		if err != nil {
			c.JSON(402, gin.H{"error": "注册失败,请重新再试..."})
		}
		//验证码机制
		code := utils.GenerateCode()
		err = utils.SentEmail(user.Email, "知序验证码", "您的验证码是："+code+"\n该验证码5分钟内有效,请尽快使用。")
		if err != nil {
			c.JSON(403, gin.H{"error": "验证码发送失败,请重新再试..."})
			utils.LogError("验证码发送失败", logrus.Fields{"user_email": user.Email})
			return
		}
		repository.SaveEmailCodeToDB(code, user.Email)
		user.Exist = false
		// 初始化用户成就表
		user = InitAchievementTable(user)
		AddUserCronJob(user)
		if err := repository.AddUserToDB(user); err != nil {
			c.JSON(405, gin.H{"error": "注册失败,请重新再试..."})
			utils.LogError("数据库添加用户失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户注册成功", logrus.Fields{"user_email": user.Email})
		c.JSON(http.StatusOK, gin.H{"message": "注册成功!"})
	}
}

// 用户登录
func LoginUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user_login struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&user_login); err != nil {
			c.JSON(400, gin.H{"error": "登录失败,请重新再试..."})
			return
		}
		user, err := repository.GetUserByEmail(user_login.Email)
		// 检查用户是否存在
		if err != nil || user.ID == 0 {
			c.JSON(401, gin.H{"error": "用户名或密码错误,请重新再试..."})
			return
		}
		// 检查邮箱是否已验证
		if !user.Exist {
			err := repository.DeleteUserByEmail(user_login.Email)
			if err != nil {
				utils.LogError("删除未验证用户失败", logrus.Fields{"user_email": user_login.Email})
			}
			c.JSON(403, gin.H{"error": "邮箱未验证,请前往验证..."})
			return
		}
		// 检查密码是否正确
		if !utils.CheckPasswordHash(user_login.Password, user.Password) {
			c.JSON(401, gin.H{"error": "用户名或密码错误,请重新再试..."})
			return
		}
		token, err := utils.GenerateToken(user.ID, user.Name, user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "生成 Token 失败",
			})
			utils.LogError("生成token失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户登录成功", logrus.Fields{"user_id": user.ID, "user_email": user.Email})
		c.JSON(http.StatusOK, gin.H{
			"message":          "登录成功!",
			"user_id":          user.ID,
			"name":             user.Name,
			"email":            user.Email,
			"head_show":        user.HeadShow,
			"daka":             user.Daka,
			"flag_number":      user.FlagNumber,
			"count":            user.Count,
			"month_learn_time": user.MonthLearntime,
			"token":            token,
		})
	}
}

// 更新用户密码
func UpdateUserPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Password    string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		id, _ := getCurrentUserID(c)
		user, _ := repository.GetUserByID(id)
		new_token, _ := utils.GenerateToken(user.ID, user.Name, user.Email)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(401, gin.H{"error": "请求失败,请重新再试..."})
			utils.LogError("请求绑定失败", logrus.Fields{})
			return
		}
		if !utils.CheckPasswordHash(req.Password, user.Password) {
			c.JSON(400, gin.H{"error": "原密码错误,请重新再试..."})
			return
		}
		req.NewPassword, _ = utils.HashPassword(req.NewPassword)
		err := repository.UpdatePassword(user.ID, req.NewPassword)
		if err != nil {
			c.JSON(500, gin.H{"message": "密码更新失败，请重新再试!"})
			utils.LogError("数据库更新用户数据失败", logrus.Fields{})
			return
		}

		utils.LogInfo("用户密码更新成功", logrus.Fields{"user_id": id})
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"new_token": new_token,
		})
	}
}

// 更新用户名
func UpdateUserName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			NewName string `json:"new_name"`
		}
		id, _ := getCurrentUserID(c)
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("UpdateUserName: 请求绑定失败: %v", err)
			c.JSON(400, gin.H{"error": "请求体格式错误, 请以 {new_name: string} 提交"})
			return
		}
		log.Printf("UpdateUserName: user_id=%d 请求新用户名=%q", id, req.NewName)
		user, _ := repository.GetUserByID(id)
		if req.NewName == user.Name {
			log.Printf("UpdateUserName: 新用户名与原用户名相同 (user_id=%d)", id)
			c.JSON(400, gin.H{"error": "新用户名与原用户名相同,请重新再试..."})
			return
		}
		if strings.TrimSpace(req.NewName) == "" {
			log.Printf("UpdateUserName: 新用户名为空 (user_id=%d)", id)
			c.JSON(400, gin.H{"error": "用户名不能为空,请重新再试..."})
			return
		}
		// 检查新用户名是否已被其他用户使用
		name_exist, _ := repository.GetUserByName(req.NewName)
		if name_exist.ID != 0 && name_exist.ID != id {
			log.Printf("UpdateUserName: 新用户名已被占用 (user_id=%d new_name=%s taken_by=%d)", id, req.NewName, name_exist.ID)
			c.JSON(400, gin.H{"error": "该用户名已被使用,请更换用户名..."})
			return
		}
		if err := repository.UpdateUserName(id, req.NewName); err != nil {
			utils.LogError("数据库更新用户名失败", logrus.Fields{"user_id": id, "new_name": req.NewName, "error": err.Error()})
			log.Printf("UpdateUserName: repository.UpdateUserName 返回错误: %v", err)
			c.JSON(500, gin.H{"message": "用户名更新失败，请稍后重试"})
			return
		}
		utils.LogInfo("用户用户名更新成功", logrus.Fields{"user_id": id, "new_name": req.NewName})
		c.JSON(http.StatusOK, gin.H{
			"success": true})
	}
}

// 更新用户状态
func UpdateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Status string `json:"status"`
		}
		id, _ := getCurrentUserID(c)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "更新状态失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		err := repository.UpdateUserStatus(id, req.Status)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新状态失败,请重新再试..."})
			utils.LogError("数据库更新用户数据失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户状态更新成功", logrus.Fields{"user_id": id, "new_status": req.Status})
		c.JSON(200, gin.H{
			"message": "状态更新成功",
			"状态":      req.Status})
	}
}

// 新增：获取指定用户统计（支持查看他人）
// 返回：打卡天数、完成flag数量、总积分、用户名、头像索引
func GetUserStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 当前用户ID（用于鉴权，至少需要登录）
		_, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}

		// 目标用户ID，可选，默认查看自己
		var targetID uint
		if q := c.Query("user_id"); q != "" {
			var parsed uint
			if _, err := fmt.Sscanf(q, "%d", &parsed); err == nil {
				targetID = parsed
			}
		}
		if targetID == 0 {
			if id, exists := c.Get("user_id"); exists {
				if vid, ok2 := id.(uint); ok2 {
					targetID = vid
				}
			}
		}

		user, err := repository.GetUserByID(targetID)
		if err != nil || user.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}

		// 已完成 flag 数量（可能与用户表中 flag_number 含义不同，这里返回两者）
		doneFlags, _ := repository.GetDoneFlagsByUserID(targetID)
		// 打卡天数使用 user.Daka
		dakaDays := user.Daka

		c.JSON(http.StatusOK, gin.H{
			"user_id":          user.ID,
			"name":             user.Name,
			"head_show":        user.HeadShow,
			"avatar_index":     user.HeadShow,
			"total_points":     user.Count,
			"month_learn_time": user.MonthLearntime,
			"completed_flags":  len(doneFlags),
			"flag_number":      user.FlagNumber,
			"daka_days":        dakaDays,
		})
	}
}

// 获取用户信息
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		user, err := repository.GetUserByID(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取用户状态失败,请重新再试..."})
			utils.LogError("数据库获取用户数据失败", logrus.Fields{})
			return
		}
		utils.LogInfo("获取用户信息成功", logrus.Fields{"user_id": id})
		c.JSON(http.StatusOK, gin.H{
			"id":               user.ID,
			"user_id":          user.ID,
			"username":         user.Name,
			"name":             user.Name,
			"email":            user.Email,
			"phone":            user.Email,
			"head_show":        user.HeadShow,
			"daka":             user.Daka,
			"flag_number":      user.FlagNumber,
			"count":            user.Count,
			"month_learn_time": user.MonthLearntime,
			"user":             user,
		})
	}
}

// 获取今日获得积分
func GetTodayPoints() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户ID失败"})
			return
		}
		total, err := repository.GetTodayPoints(id)
		if err != nil {
			utils.LogError("获取今日积分失败", logrus.Fields{"user_id": id, "error": err.Error()})
			c.JSON(500, gin.H{"today_points": 0})
			return
		}
		utils.LogInfo("获取今日积分成功", logrus.Fields{"user_id": id, "today_points": total})
		c.JSON(200, gin.H{"today_points": total})
	}
}

// 打卡
func DoDaKa() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		err := repository.DakaNumberToDB(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "打卡失败,请重新再试..."})
			utils.LogError("数据库更新用户打卡数据失败", logrus.Fields{"error": err.Error()})
			return
		}
		utils.LogInfo("用户打卡成功", logrus.Fields{"user_id": id})
		c.JSON(http.StatusOK, gin.H{"message": "打卡成功!"})
	}
}

// 获取打卡此月天的打卡记录
func GetDaKaRecords() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		dakaRecords, err := repository.GetMonthDakaRecords(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取打卡记录失败,请重新再试..."})
			utils.LogError("获取打卡记录失败", logrus.Fields{"user_id": id, "error": err.Error()})
			return
		}

		// 转换为前端需要的日期格式数组
		var dates []map[string]string
		for _, record := range dakaRecords {
			dates = append(dates, map[string]string{
				"date": record.DaKaDate.Format("2006-01-02"),
			})
		}

		utils.LogInfo("获取打卡记录成功", logrus.Fields{"user_id": id, "count": len(dates)})
		c.JSON(200, dates)
	}
}

// 用户选择的时间定时提醒
func UpdateUserRemindTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		var Remind struct {
			RemindHour int `json:"time_remind"`
			ReminMin   int `json:"min_remind"`
		}
		if err := c.ShouldBindJSON(&Remind); err != nil {
			c.JSON(400, gin.H{"error": "获取用户提醒时间失败,请重新再试..."})
			utils.LogError("获取用户提醒时间失败", logrus.Fields{})
			return
		}
		id, _ := getCurrentUserID(c)

		// 先开启提醒状态（如果还未开启）
		user, _ := repository.GetUserByID(id)
		if !user.IsRemind {
			repository.UpdateUserRemindStatus(id, true)
			utils.LogInfo("自动开启提醒功能", logrus.Fields{"user_id": id})
		}

		// 更新提醒时间
		err := repository.UpdateUserRemindTime(id, Remind.RemindHour, Remind.ReminMin)
		if err != nil {
			c.JSON(500, gin.H{"error": "更新用户提醒时间失败,请重新再试..."})
			utils.LogError("更新用户提醒时间失败", logrus.Fields{})
			return
		}

		// 更新定时任务（提醒状态为 true）
		UpdateUserReminderJob(id, Remind.RemindHour, Remind.ReminMin, true)

		utils.LogInfo("更新用户提醒时间成功", logrus.Fields{"user_id": id, "remind_hour": Remind.RemindHour, "remin_min": Remind.ReminMin})
		c.JSON(200, gin.H{"message": "更新用户提醒时间成功!"})
	}
}

// 用户选择是否开启提醒
func UpdateUserRemind() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		user, _ := repository.GetUserByID(id)
		user.IsRemind = !user.IsRemind
		err := repository.UpdateUserRemindStatus(id, user.IsRemind)
		if err != nil {
			c.JSON(500, gin.H{"error": "更新用户提醒状态失败,请重新再试..."})
			utils.LogError("更新用户提醒状态失败", logrus.Fields{})
			return
		}

		// 更新定时任务
		UpdateUserReminderJob(id, user.RemindHour, user.RemindMin, user.IsRemind)

		utils.LogInfo("更新用户提醒状态成功", logrus.Fields{"user_id": id, "is_remind": user.IsRemind})
		c.JSON(200, gin.H{"message": "更新用户提醒状态成功!",
			"状态": user.IsRemind})
	}
}

// 头像切换
func SwithHead() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Number int `json:"number"`
		}
		id, _ := getCurrentUserID(c)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "头像切换失败,请重新再试..."})
			log.Print("Binding error")
			return
		}

		// 验证头像编号必须在1-32之间（支持全部头像）
		if req.Number < 1 || req.Number > 32 {
			c.JSON(400, gin.H{"error": "头像编号必须在1-32之间"})
			log.Printf("Invalid avatar number: %d", req.Number)
			return
		}

		log.Printf("切换头像 - 用户ID: %d, 头像编号: %d", id, req.Number)
		user, _ := repository.GetUserByID(id)
		user.HeadShow = req.Number
		repository.SaveUserToDB(user)
		c.JSON(200, gin.H{"success": true})
	}
}

// 新增：添加积分接口
func AddPointsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "用户未登录"})
			utils.LogError("添加积分失败：用户未登录", nil)
			return
		}
		var req struct {
			Points int `json:"points"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			// 问题4修复：参数绑定失败时提供详细错误信息
			c.JSON(400, gin.H{"error": "参数错误：积分值必须是有效的数字"})
			utils.LogError("添加积分失败：参数绑定错误", logrus.Fields{"error": err.Error(), "body": c.Request.Body})
			return
		}

		// 问题2修复：验证积分值必须为正整数
		if req.Points <= 0 {
			c.JSON(400, gin.H{"error": "积分值必须大于0"})
			utils.LogError("添加积分失败：积分值无效", logrus.Fields{"user_id": id, "points": req.Points})
			return
		}

		utils.LogInfo("开始添加积分", logrus.Fields{"user_id": id, "points": req.Points})

		// 问题5&6修复：使用原子自增操作，直接传递增量
		err := repository.CountAddDB(id, req.Points)
		if err != nil {
			c.JSON(500, gin.H{"error": "积分添加失败，请稍后重试"})
			utils.LogError("积分添加失败：数据库更新错误", logrus.Fields{"user_id": id, "points": req.Points, "error": err.Error()})
			return
		}

		// 重新查询更新后的积分
		user, err := repository.GetUserByID(id)
		if err != nil {
			// 即使查询失败，积分已添加成功
			utils.LogError("查询用户失败（积分已添加）", logrus.Fields{"user_id": id, "error": err.Error()})
			c.JSON(200, gin.H{"message": "积分添加成功", "count": 0})
			return
		}

		// 问题1&3修复：确保返回正确的JSON结构，字段名小写
		utils.LogInfo("积分添加成功", logrus.Fields{
			"user_id":   id,
			"points":    req.Points,
			"old_count": user.Count - req.Points,
			"new_count": user.Count,
		})
		c.JSON(200, gin.H{"message": "积分添加成功", "count": user.Count})
	}
}
