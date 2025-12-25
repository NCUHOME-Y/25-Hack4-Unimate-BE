package service

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 输入验证辅助函数
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validateEmail(email string) bool {
	return len(email) <= 100 && emailRegex.MatchString(email)
}

func validateUsername(name string) bool {
	return len(name) >= 2 && len(name) <= 20
}

func validatePassword(password string) bool {
	// 🔒 安全加固：强制密码强度要求（匹配前端）
	if len(password) < 8 || len(password) > 100 {
		return false
	}

	// 至少包含3种类型字符
	typeCount := 0
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		typeCount++ // 小写字母
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		typeCount++ // 大写字母
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		typeCount++ // 数字
	}
	if regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		typeCount++ // 特殊字符
	}

	return typeCount >= 3
}

// 安全比较验证码（防止时序攻击）
func secureCompareCode(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

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

// 用户注册（第一步：发送验证码）
func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": "请求参数错误"})
			utils.LogError("注册参数绑定错误", logrus.Fields{"error": err.Error()})
			return
		}

		// 输入验证
		if !validateEmail(user.Email) {
			c.JSON(400, gin.H{"error": "邮箱格式不正确"})
			return
		}
		if !validateUsername(user.Name) {
			c.JSON(400, gin.H{"error": "用户名长度需在2-20个字符之间"})
			return
		}
		if !validatePassword(user.Password) {
			c.JSON(400, gin.H{"error": "密码长度需在6-100个字符之间"})
			return
		}

		// 检查邮箱是否已注册
		user_exist, _ := repository.GetUserBasicByEmail(user.Email) // 🔧 性能优化
		if user_exist.ID != 0 {
			c.JSON(409, gin.H{"error": "该邮箱已被注册，请直接登录或使用其他邮箱"})
			utils.LogInfo("注册失败-邮箱已存在", logrus.Fields{"email": user.Email})
			return
		}

		// 检查用户名是否已存在
		name_exist, _ := repository.GetUserByName(user.Name)
		if name_exist.ID != 0 {
			c.JSON(409, gin.H{"error": "该用户名已被使用，请更换用户名"})
			utils.LogInfo("注册失败-用户名已存在", logrus.Fields{"name": user.Name})
			return
		}

		// 生成并发送验证码
		code := utils.GenerateCode()
		emailBody := fmt.Sprintf(`尊敬的用户，您好！

			您正在注册知序学习平台账号，验证码为：

    					%s

			该验证码将在5分钟内有效，请及时使用。

			请登录知序平台（http://139.199.157.76）查看详情。

				——知序平台`, code)
		err := utils.SentEmail(user.Email, "【知序】账号注册验证码", emailBody)
		if err != nil {
			c.JSON(500, gin.H{"error": "验证码发送失败，请稍后重试"})
			utils.LogError("验证码发送失败", logrus.Fields{"email": user.Email, "error": err.Error()})
			return
		}

		// 保存验证码到数据库
		repository.SaveEmailCodeToDB(code, user.Email)

		utils.LogInfo("注册验证码已发送", logrus.Fields{"email": user.Email})
		c.JSON(http.StatusOK, gin.H{
			"message": "验证码已发送到您的邮箱，请查收并输入验证码完成注册",
			"email":   user.Email,
		})
	}
}

// 完成注册（第二步：验证验证码并创建用户）
func CompleteRegistration() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
			Code     string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "请求参数错误"})
			utils.LogError("完成注册参数绑定错误", logrus.Fields{"error": err.Error()})
			return
		}

		// 输入验证
		if !validateEmail(req.Email) {
			c.JSON(400, gin.H{"error": "邮箱格式不正确"})
			return
		}
		if !validateUsername(req.Name) {
			c.JSON(400, gin.H{"error": "用户名长度需在2-20个字符之间"})
			return
		}
		if !validatePassword(req.Password) {
			c.JSON(400, gin.H{"error": "密码长度需在6-100个字符之间"})
			return
		}

		// 验证验证码
		emailCode, err := repository.GetEmailCodeByEmail(req.Email)
		if err != nil {
			c.JSON(400, gin.H{"error": "验证码不存在或已过期"})
			utils.LogError("获取邮箱验证码失败", logrus.Fields{"email": req.Email, "error": err.Error()})
			return
		}

		// 使用安全比较防止时序攻击
		if !secureCompareCode(emailCode.Code, req.Code) {
			c.JSON(400, gin.H{"error": "验证码错误"})
			utils.LogWarn("验证码错误", logrus.Fields{"email": req.Email})
			return
		}

		if emailCode.Expires.Before(time.Now()) {
			c.JSON(400, gin.H{"error": "验证码已过期，请重新获取"})
			utils.LogWarn("验证码已过期", logrus.Fields{"email": req.Email})
			return
		}

		// 再次检查邮箱和用户名是否被占用（防止并发注册）
		user_exist, _ := repository.GetUserBasicByEmail(req.Email) // 🔧 性能优化
		if user_exist.ID != 0 {
			c.JSON(409, gin.H{"error": "该邮箱已被注册"})
			return
		}
		name_exist, _ := repository.GetUserByName(req.Name)
		if name_exist.ID != 0 {
			c.JSON(409, gin.H{"error": "该用户名已被使用"})
			return
		}

		// 创建用户对象并哈希密码
		user := &model.User{
			Name:  req.Name,
			Email: req.Email,
		}
		user.Password, err = utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(500, gin.H{"error": "密码加密失败"})
			utils.LogError("密码哈希失败", logrus.Fields{"email": req.Email, "error": err.Error()})
			return
		}

		// 添加用户到数据库，获取自动生成的ID
		if err := repository.AddUserToDB(user); err != nil {
			c.JSON(500, gin.H{"error": "创建用户失败，请稍后重试"})
			utils.LogError("数据库添加用户失败", logrus.Fields{"email": req.Email, "error": err.Error()})
			return
		}

		// 删除已使用的验证码（防止重复使用）
		if err := repository.DeleteEmailCode(req.Email); err != nil {
			utils.LogWarn("删除验证码失败", logrus.Fields{"email": req.Email, "error": err.Error()})
		}

		// 初始化成就表
		achievements := []model.Achievement{
			{UserID: user.ID, Name: "首次完成", Description: "第一次设置flag", HadDone: false},
			{UserID: user.ID, Name: "7天连卡", Description: "连续打卡7天", HadDone: false},
			{UserID: user.ID, Name: "任务大师", Description: "完成50个flag", HadDone: false},
			{UserID: user.ID, Name: "目标达成", Description: "积分超过1000", HadDone: false},
			{UserID: user.ID, Name: "学习之星", Description: "累计学习时间超过1000分钟", HadDone: false},
			{UserID: user.ID, Name: "坚持不懈", Description: "累计打卡30天", HadDone: false},
			{UserID: user.ID, Name: "效率达人", Description: "单日完成5个flag", HadDone: false},
			{UserID: user.ID, Name: "专注大师", Description: "单日学习时长超过4小时", HadDone: false},
			{UserID: user.ID, Name: "早起鸟", Description: "早上6点前打卡5次", HadDone: false},
			{UserID: user.ID, Name: "夜猫子", Description: "晚上10点后打卡5次", HadDone: false},
			{UserID: user.ID, Name: "完美主义", Description: "连续10次满分完成flag", HadDone: false},
			{UserID: user.ID, Name: "全能选手", Description: "完成5种不同标签的flag", HadDone: false},
			{UserID: user.ID, Name: "学习狂人", Description: "累计学习时间超过5000分钟", HadDone: false},
			{UserID: user.ID, Name: "社交达人", Description: "发布10条动态", HadDone: false},
			{UserID: user.ID, Name: "时间管理者", Description: "连续30天完成至少1个flag", HadDone: false},
			{UserID: user.ID, Name: "成就收集者", Description: "解锁10个徽章", HadDone: false},
		}

		// 批量插入成就记录
		if err := repository.BatchCreateAchievements(achievements); err != nil {
			utils.LogError("初始化成就表失败", logrus.Fields{"user_id": user.ID, "error": err.Error()})
			// 不阻断注册流程，只记录错误
		}

		// 添加定时任务
		AddUserCronJob(*user)

		// 生成JWT token
		token, err := utils.GenerateToken(user.ID, user.Name, user.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": "生成token失败"})
			utils.LogError("生成token失败", logrus.Fields{"user_id": user.ID, "error": err.Error()})
			return
		}

		utils.LogInfo("用户注册成功", logrus.Fields{"user_id": user.ID, "email": user.Email, "name": user.Name})

		c.JSON(http.StatusOK, gin.H{
			"message":          "注册成功！",
			"token":            token,
			"user_id":          user.ID,
			"name":             user.Name,
			"email":            user.Email,
			"head_show":        user.HeadShow,
			"daka":             user.Daka,
			"flag_number":      user.FlagNumber,
			"count":            user.Count,
			"month_learn_time": user.MonthLearntime,
		})
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

		// 🔒 安全加固：检查登录失败次数限制
		loginKey := fmt.Sprintf("login_fail:%s", user_login.Email)
		lockKey := fmt.Sprintf("login_lock:%s", user_login.Email)

		// 检查是否被锁定
		if repository.RedisClient != nil {
			locked, err := repository.RedisClient.Get(repository.Ctx, lockKey).Result()
			if err == nil && locked == "1" {
				ttl := repository.RedisClient.TTL(repository.Ctx, lockKey).Val()
				c.JSON(429, gin.H{
					"error":   "登录失败次数过多，账户已被临时锁定",
					"message": fmt.Sprintf("请等待 %d 分钟后再试", int(ttl.Minutes())+1),
				})
				utils.LogWarn("账户登录被锁定", logrus.Fields{"email": user_login.Email})
				return
			}
		}

		user, err := repository.GetUserBasicByEmail(user_login.Email) // 🔧 性能优化：登录验证只需要基本信息
		// 检查用户是否存在
		if err != nil || user.ID == 0 {
			// 记录失败次数
			if repository.RedisClient != nil {
				repository.RedisClient.Incr(repository.Ctx, loginKey)
				repository.RedisClient.Expire(repository.Ctx, loginKey, 15*time.Minute)
			}
			c.JSON(401, gin.H{"error": "用户名或密码错误,请重新再试..."})
			return
		}
		// 检查密码是否正确
		if !utils.CheckPasswordHash(user_login.Password, user.Password) {
			// 🔒 安全加固：记录失败次数并检查是否需要锁定
			if repository.RedisClient != nil {
				failCount, _ := repository.RedisClient.Incr(repository.Ctx, loginKey).Result()
				repository.RedisClient.Expire(repository.Ctx, loginKey, 15*time.Minute)

				// 5次失败后锁定15分钟
				if failCount >= 5 {
					repository.RedisClient.Set(repository.Ctx, lockKey, "1", 15*time.Minute)
					repository.RedisClient.Del(repository.Ctx, loginKey)
					c.JSON(429, gin.H{
						"error":   "登录失败次数过多，账户已被锁定",
						"message": "请15分钟后再试",
					})
					utils.LogWarn("账户因多次登录失败被锁定", logrus.Fields{"email": user_login.Email, "fail_count": failCount})
					return
				}

				utils.LogWarn("登录密码错误", logrus.Fields{"email": user_login.Email, "fail_count": failCount})
			}
			c.JSON(401, gin.H{"error": "用户名或密码错误,请重新再试..."})
			return
		}

		// 🔒 登录成功，清除失败记录
		if repository.RedisClient != nil {
			repository.RedisClient.Del(repository.Ctx, loginKey)
			repository.RedisClient.Del(repository.Ctx, lockKey)
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
		id, _ := utils.GetCurrentUserID(c)
		user, _ := repository.GetUserBasicByID(id) // 🔧 性能优化：使用轻量级查询
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
		id, _ := utils.GetCurrentUserID(c)
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("UpdateUserName: 请求绑定失败: %v", err)
			c.JSON(400, gin.H{"error": "请求体格式错误, 请以 {new_name: string} 提交"})
			return
		}
		log.Printf("UpdateUserName: user_id=%d 请求新用户名=%q", id, req.NewName)
		user, _ := repository.GetUserBasicByID(id) // 🔧 性能优化
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
		id, _ := utils.GetCurrentUserID(c)
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
		_, ok := utils.GetCurrentUserID(c)
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
		id, ok := utils.GetCurrentUserID(c)
		if !ok {
			c.JSON(401, gin.H{"error": "未授权，请先登录"})
			return
		}
		if id == 0 {
			c.JSON(400, gin.H{"error": "无效的用户ID"})
			return
		}
		user, err := repository.GetUserBasicByID(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取用户状态失败,请重新再试..."})
			utils.LogError("数据库获取用户数据失败", logrus.Fields{"user_id": id, "error": err.Error()})
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
		id, ok := utils.GetCurrentUserID(c)
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
		id, _ := utils.GetCurrentUserID(c)
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
		id, _ := utils.GetCurrentUserID(c)
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
		id, _ := utils.GetCurrentUserID(c)

		// 先开启学习提醒状态（如果还未开启）
		user, _ := repository.GetUserByID(id)
		if !user.IsStudyRemind {
			repository.UpdateUserStudyRemindStatus(id, true)
			utils.LogInfo("自动开启学习提醒功能", logrus.Fields{"user_id": id})
		}

		// 更新学习提醒时间（同时兼容旧字段）
		err := repository.UpdateUserRemindTime(id, Remind.RemindHour, Remind.ReminMin)
		if err != nil {
			c.JSON(500, gin.H{"error": "更新用户提醒时间失败,请重新再试..."})
			utils.LogError("更新用户提醒时间失败", logrus.Fields{})
			return
		}

		// 更新学习提醒定时任务（学习提醒状态为 true）
		UpdateUserReminderJob(id, Remind.RemindHour, Remind.ReminMin, true)

		utils.LogInfo("更新用户提醒时间成功", logrus.Fields{"user_id": id, "remind_hour": Remind.RemindHour, "remin_min": Remind.ReminMin})
		c.JSON(200, gin.H{"message": "更新用户提醒时间成功!"})
	}
}

// 用户选择是否开启提醒
func UpdateUserRemind() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := utils.GetCurrentUserID(c)
		user, _ := repository.GetUserByID(id)
		// 切换学习提醒开关
		user.IsStudyRemind = !user.IsStudyRemind
		err := repository.UpdateUserStudyRemindStatus(id, user.IsStudyRemind)
		if err != nil {
			c.JSON(500, gin.H{"error": "更新用户提醒状态失败,请重新再试..."})
			utils.LogError("更新用户提醒状态失败", logrus.Fields{})
			return
		}

		// 更新学习提醒定时任务
		UpdateUserReminderJob(id, user.StudyRemindHour, user.StudyRemindMin, user.IsStudyRemind)

		utils.LogInfo("更新用户学习提醒状态成功", logrus.Fields{"user_id": id, "is_study_remind": user.IsStudyRemind})
		c.JSON(200, gin.H{"message": "更新用户学习提醒状态成功!",
			"状态": user.IsStudyRemind})
	}
}

// 更新用户 Flag 提醒（用户级总开关）
func UpdateUserFlagRemind() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := utils.GetCurrentUserID(c)
		user, _ := repository.GetUserByID(id)
		user.IsFlagRemind = !user.IsFlagRemind
		err := repository.UpdateUserFlagRemindStatus(id, user.IsFlagRemind)
		if err != nil {
			c.JSON(500, gin.H{"error": "更新用户 Flag 提醒状态失败,请重新再试..."})
			utils.LogError("更新用户 Flag 提醒状态失败", logrus.Fields{})
			return
		}

		utils.LogInfo("更新用户 Flag 提醒状态成功", logrus.Fields{"user_id": id, "is_flag_remind": user.IsFlagRemind})
		c.JSON(200, gin.H{"message": "更新用户 Flag 提醒状态成功!", "状态": user.IsFlagRemind})
	}
}

// 头像切换
func SwithHead() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Number int `json:"number"`
		}
		id, _ := utils.GetCurrentUserID(c)
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
		err := repository.UpdateUserHeadShow(id, req.Number)
		if err != nil {
			log.Printf("更新头像失败: %v", err)
			c.JSON(500, gin.H{"error": "更新头像失败"})
			return
		}
		c.JSON(200, gin.H{"success": true})
	}
}

// 新增：添加积分接口
func AddPointsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
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
