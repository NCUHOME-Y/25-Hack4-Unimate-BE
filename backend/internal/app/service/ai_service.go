package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 学习计划请求
type LearningPlanRequest struct {
	Flag       string `json:"flag" binding:"required"` // 学习目标标识
	Background string `json:"background,omitempty"`    // 用户背景
	Difficulty int    `json:"difficulty,omitempty"`    // 难度分数: 50=简单, 150=中等, 200=困难
}

// 学习计划响应
type LearningPlanResponse struct {
	Success bool   `json:"success"`
	Flag    string `json:"flag"`
	Count   int    `json:"difficulty"` // 难度评分: 1,2,3
	Plan    string `json:"plan"`
	Error   string `json:"error,omitempty"`
}

// 太傅AI学习
type TaiFuLearningPlanner struct {
	APIKey  string
	BaseURL string
}

var planner *TaiFuLearningPlanner

// 初始化 planner（延迟初始化，等待 .env 加载）
func initPlanner() {
	if planner == nil {
		apiKey := os.Getenv("APIKEY")
		if apiKey == "" {
			logrus.Error("APIKEY 环境变量未设置")
		}
		planner = &TaiFuLearningPlanner{
			APIKey:  apiKey,
			BaseURL: "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		}
	}
}

// 检测输入是否为有效的学习目标
func isValidLearningGoal(input string) bool {
	// 去除空格
	trimmed := strings.TrimSpace(input)

	// 长度检查
	if len(trimmed) < 2 || len(trimmed) > 200 {
		return false
	}

	// 检查是否包含有意义的汉字、英文或数字
	hasValidContent := false
	for _, r := range trimmed {
		if (r >= '\u4e00' && r <= '\u9fa5') || // 汉字
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || // 英文
			(r >= '0' && r <= '9') { // 数字
			hasValidContent = true
			break
		}
	}
	if !hasValidContent {
		return false
	}

	// 检查是否全是重复字符（如：aaaaaa）
	if isRepeatingChars(trimmed) {
		return false
	}

	// 检查是否全是无意义符号
	invalidPatterns := []string{
		"!!!!!", "?????", ".....", "-----", "*****",
		"asdfg", "qwert", "12345", "abcde",
	}
	for _, pattern := range invalidPatterns {
		if strings.Contains(strings.ToLower(trimmed), pattern) {
			return false
		}
	}

	return true
}

// 检查是否为重复字符
func isRepeatingChars(s string) bool {
	if len(s) < 3 {
		return false
	}
	firstChar := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != firstChar {
			return false
		}
	}
	return true
}

func GenerateLearningPlan(c *gin.Context) {
	initPlanner()
	id, _ := utils.GetCurrentUserID(c)
	var req LearningPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LearningPlanResponse{
			Success: false,
			Error:   fmt.Sprintf("请求格式错误: %v", err),
		})
		return
	}

	// 输入合法性检测
	if !isValidLearningGoal(req.Flag) {
		c.JSON(http.StatusBadRequest, LearningPlanResponse{
			Success: false,
			Error:   "输入内容无效，请输入有意义的学习目标（如：学习Python编程、提升英语口语等）",
		})
		return
	}

	flag, plan, difficulty, err := planner.GenerateLearningPlan(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"goal":  req.Flag,
		}).Error("AI计划生成失败")
		if strings.Contains(err.Error(), "智谱API限流或额度不足") {
			c.JSON(http.StatusTooManyRequests, LearningPlanResponse{
				Success: false,
				Error:   "智谱API限流或额度不足，请稍后重试或更换APIKEY",
			})
			return
		}
		if strings.Contains(err.Error(), "AI服务繁忙") {
			c.JSON(http.StatusServiceUnavailable, LearningPlanResponse{
				Success: false,
				Error:   "AI服务繁忙，请稍后重试",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, LearningPlanResponse{
			Success: false,
			Error:   fmt.Sprintf("生成计划失败: %v", err),
		})
		return
	}

	// 埋点：生成计划（不添加Flag，让前端决定）
	// 注意：埋点失败不影响主功能，仅记录错误日志
	if err := repository.AddTrackPointToDB(id, "生成AI计划"); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": id,
			"error":   err.Error(),
		}).Warn("埋点记录失败（不影响功能）")
	}

	c.JSON(http.StatusOK, LearningPlanResponse{
		Success: true,
		Flag:    flag,
		Count:   difficulty,
		Plan:    plan,
	})
}

// CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// 生成学习计划的核心方法
func (p *TaiFuLearningPlanner) GenerateLearningPlan(req LearningPlanRequest) (string, string, int, error) {
	// 构建系统提示词
	systemPrompt := `你是"太傅AI计划生成器"，专门为用户制定科学、合理、可执行的目标达成计划。你可以帮助用户规划学习、健康、习惯养成、兴趣培养、工作效率等各类目标。

【核心要求】
1. 必须返回标准JSON格式（不要包含markdown代码块标记）
2. plan字段第一行必须是【目标概述】
3. 所有任务参数必须用中文格式：（每日完成：X次）
4. 绝对禁止英文参数名：count、limit、total、daily等
5. 难度分数与时间对应：
   - 100分 = 入门级（1-3天，3-5个任务，每日1-2次）
   - 200分 = 进阶级（1-2周，5-6个任务，每日1-4次）
   - 300分 = 专家级（1-2月，6-8个任务，每日2-5次）

【目标分类指导】
根据目标内容，在flag和plan中使用对应类型的关键词：
- 学习提升类：使用"学习、掌握、理解、练习、阅读、复习、知识、技能、能力"等词
- 健康运动类：使用"锻炼、运动、健身、跑步、健康、睡眠、饮食、体能、早起"等词
- 工作效率类：使用"工作、项目、任务、开发、实现、完成、效率、时间管理"等词
- 兴趣爱好类：使用"兴趣、爱好、娱乐、音乐、绘画、摄影、游戏、创作"等词
- 生活习惯类：使用"习惯、日常、生活、坚持、打卡、规律、整理、记录"等词

【内容要求】
每个阶段必须包含完整的四个部分：
- 阶段标题（如：阶段一：基础入门（预计1-3天））
- 【阶段目标】说明核心能力和预期成果（2-3句话）
- 【行动要点】列出3-5个关键要点，每个用"-"开头，说明具体内容和应用
- 【实践建议】给出具体的方法、资源推荐、时间安排（2-3句话）
- 可执行任务列表：列出3个以上可执行任务，每个格式：序号. 任务描述（每日完成：X次）

注意：每个阶段的四个部分【阶段目标】【行动要点】【实践建议】和任务列表都不能省略

【每日完成次数】
- 入门级：每任务1次
- 进阶级：每任务1-2次
- 专家级：每任务2-3次

【返回JSON格式】
{
    "flag": "计划标题（8-15字）",
    "difficulty": 100或200或300,
    "plan": "三阶段计划文本"
}

【plan字段格式】
第1行：【目标概述】目标说明和预期成果

然后包含3个阶段，每个阶段格式：
- 阶段标题：阶段一/阶段二/阶段三（含时间）
- 【阶段目标】核心能力或成果说明
- 【行动要点】关键点，用"-"标记
- 【实践建议】方法和资源
- 可执行任务列表（不显示标题），格式：序号. 任务描述（每日完成：X次）

【学习类示例】
{
    "flag": "Python编程入门",
    "difficulty": 200,
    "plan": "【目标概述】系统掌握Python编程，从基础到实战，具备独立开发能力。

阶段一：基础入门（预计1-3天）
【阶段目标】掌握Python基础语法，能编写简单程序，理解编程思维。
【行动要点】
- 变量与数据类型（整数、浮点数、字符串、布尔值），理解数据存储。
- 条件语句（if-elif-else）和循环（for/while），掌握流程控制。
- 函数定义、调用和参数传递，理解模块化编程。
【实践建议】
使用Python官方教程配合视频课程学习，每天编写2-3个小程序巩固知识，在LeetCode刷入门题。
1. 完成Python基础教程前5章并做笔记（每日完成：1次）
2. 编写基础练习程序（变量、循环、函数）并添加注释（每日完成：1次）
3. 总结学习心得并分享（每日完成：1次）"
}

【错误示例】
❌ 不能有：count: 10, limit: 2, total: 5, daily: 1
❌ 不能有：[count:10] [limit:2] {total:5}
❌ plan开头不能直接是"阶段一"，必须先有【目标概述】
❌ 不能把【学习要点】写成【行动要点】以外的名称

【必须遵守】
✅ 必须是：（每日完成：X次）
✅ plan第一行必须是：【目标概述】...
✅ 任务列表直接跟在【实践建议】后面，绝对不能有"【具体任务】"标题
✅ 严格按照标准示例格式
✅ 根据目标类型灵活调整内容，学习类用【学习要点】或【行动要点】都可以，健康/习惯类优先用【行动要点】`

	// 构建用户提示词
	userPrompt := fmt.Sprintf("目标: %s\n", req.Flag)
	if req.Background != "" {
		userPrompt += fmt.Sprintf("个人背景: %s\n", req.Background)
	}
	if req.Difficulty != 0 {
		userPrompt += fmt.Sprintf("期望难度分数: %d\n", req.Difficulty)
	}
	userPrompt += "\n请根据以上信息生成计划,返回标准JSON格式。"

	// 调用AI
	response, err := p.callOpenAI(systemPrompt, userPrompt)
	if err != nil {
		return "", "", 0, err
	}

	// 解析AI响应
	flag, plan, difficulty, err := p.parseAIResponse(response)
	if err != nil {
		return "", "", 0, err
	}

	// 验证结果
	if plan == "" {
		return "", "", 0, fmt.Errorf("AI返回的学习计划为空")
	}
	if difficulty == 0 {
		difficulty = req.Difficulty // 使用请求的难度作为默认值
		if difficulty == 0 {
			difficulty = 150 // 默认中等难度
		}
	}

	return flag, plan, difficulty, nil
}

// callOpenAI 发起 AI 请求并返回原始响应字符串
func (p *TaiFuLearningPlanner) callOpenAI(systemPrompt, userPrompt string) (string, error) {
	models := []string{
		"glm-4.7-flash",
		"glm-4-flash-250414",
	}

	var lastErr error
	for _, modelName := range models {
		response, err := p.callOpenAIWithModel(modelName, systemPrompt, userPrompt)
		if err == nil {
			return response, nil
		}
		lastErr = err

		// 上游限流或超时时，尝试切换备用模型。
		if strings.Contains(err.Error(), "状态码429") || strings.Contains(err.Error(), "context deadline exceeded") {
			logrus.WithFields(logrus.Fields{
				"model": modelName,
				"error": err.Error(),
			}).Warn("主模型不可用，尝试切换备用模型")
			continue
		}

		// 非限流错误直接返回，避免掩盖配置类问题（如 key 无效）。
		return "", err
	}

	if lastErr != nil {
		if strings.Contains(lastErr.Error(), "code\":\"1302\"") {
			return "", fmt.Errorf("智谱API限流或额度不足，请稍后重试或更换APIKEY")
		}
		if strings.Contains(lastErr.Error(), "状态码429") || strings.Contains(lastErr.Error(), "context deadline exceeded") {
			return "", fmt.Errorf("AI服务繁忙，请稍后重试")
		}
		return "", lastErr
	}
	return "", fmt.Errorf("AI请求失败")
}

func (p *TaiFuLearningPlanner) callOpenAIWithModel(modelName, systemPrompt, userPrompt string) (string, error) {
	payload := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", p.BaseURL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	client := &http.Client{Timeout: 8 * time.Second}

	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		req, err := http.NewRequest("POST", p.BaseURL, bytes.NewReader(b))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		if p.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+p.APIKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if strings.Contains(err.Error(), "context deadline exceeded") && attempt < 2 {
				time.Sleep(time.Duration(attempt) * 700 * time.Millisecond)
				continue
			}
			return "", err
		}

		bodyBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			return "", readErr
		}

		if resp.StatusCode == 200 {
			return string(bodyBytes), nil
		}

		if resp.StatusCode == 429 {
			// 限流场景不在同一模型内重试，交由上层切换备用模型。
			return "", fmt.Errorf("AI API错误(状态码%d): %s", resp.StatusCode, string(bodyBytes))
		}

		apiErr := fmt.Errorf("AI API错误(状态码%d): %s", resp.StatusCode, string(bodyBytes))
		lastErr = apiErr

		// 5xx通常是上游拥塞，短暂退避后重试。
		if resp.StatusCode >= 500 && attempt < 2 {
			time.Sleep(time.Duration(attempt) * 700 * time.Millisecond)
			continue
		}

		return "", apiErr
	}

	if lastErr != nil {
		return "", lastErr
	}

	return "", fmt.Errorf("AI请求失败")
}

// parseAIResponse 尝试从 AI 响应中解析 JSON 格式的 flag/plan/difficulty
func (p *TaiFuLearningPlanner) parseAIResponse(response string) (string, string, int, error) {
	// 优先尝试 OpenAI 样式的 wrapper
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(response), &wrapper); err == nil && len(wrapper.Choices) > 0 {
		content := wrapper.Choices[0].Message.Content
		var out struct {
			Flag       string `json:"flag"`
			Plan       string `json:"plan"`
			Difficulty int    `json:"difficulty"`
		}
		if err2 := json.Unmarshal([]byte(content), &out); err2 == nil {
			return out.Flag, out.Plan, out.Difficulty, nil
		}

		if jsonText := extractJSONObject(content); jsonText != "" {
			if err3 := json.Unmarshal([]byte(jsonText), &out); err3 == nil {
				return out.Flag, out.Plan, out.Difficulty, nil
			}
		}
	}

	// 直接尝试将响应解析为目标结构
	var out struct {
		Flag       string `json:"flag"`
		Plan       string `json:"plan"`
		Difficulty int    `json:"difficulty"`
	}
	if err := json.Unmarshal([]byte(response), &out); err == nil {
		if out.Plan != "" {
			return out.Flag, out.Plan, out.Difficulty, nil
		}
	}

	if jsonText := extractJSONObject(response); jsonText != "" {
		if err := json.Unmarshal([]byte(jsonText), &out); err == nil && out.Plan != "" {
			return out.Flag, out.Plan, out.Difficulty, nil
		}
	}

	return "", "", 0, fmt.Errorf("无法解析AI响应")
}

func extractJSONObject(s string) string {
	start := -1
	depth := 0

	for i, r := range s {
		if r == '{' {
			if depth == 0 {
				start = i
			}
			depth++
			continue
		}
		if r == '}' {
			if depth > 0 {
				depth--
				if depth == 0 && start >= 0 {
					return s[start : i+1]
				}
			}
		}
	}

	return ""
}

// ==================== AI 历史记录相关接口 ====================

// 保存AI历史记录
func SaveAIHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok || id == 0 {
			c.JSON(401, gin.H{"error": "未授权，请先登录"})
			return
		}

		var req struct {
			Background    string `json:"background"`
			Goal          string `json:"goal"`
			Difficulty    string `json:"difficulty"`
			GeneratedPlan string `json:"generatedPlan"` // JSON字符串
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数格式错误"})
			utils.LogError("解析AI历史记录参数失败", logrus.Fields{"error": err.Error()})
			return
		}

		// 创建AI历史记录
		aiHistory := model.AIHistory{
			UserID:        id,
			Background:    req.Background,
			Goal:          req.Goal,
			Difficulty:    req.Difficulty,
			GeneratedPlan: req.GeneratedPlan,
			CreatedAt:     time.Now(),
		}

		if err := repository.DB.Create(&aiHistory).Error; err != nil {
			c.JSON(500, gin.H{"error": "保存AI历史记录失败"})
			utils.LogError("保存AI历史记录失败", logrus.Fields{"user_id": id, "error": err.Error()})
			return
		}

		utils.LogInfo("保存AI历史记录成功", logrus.Fields{"user_id": id, "goal": req.Goal})
		c.JSON(200, gin.H{"success": true, "message": "AI历史记录已保存"})
	}
}

// 获取AI历史记录
func GetAIHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok || id == 0 {
			c.JSON(401, gin.H{"error": "未授权，请先登录"})
			return
		}

		var aiHistories []model.AIHistory
		if err := repository.DB.Where("user_id = ?", id).Order("created_at desc").Limit(10).Find(&aiHistories).Error; err != nil {
			c.JSON(500, gin.H{"error": "获取AI历史记录失败"})
			utils.LogError("获取AI历史记录失败", logrus.Fields{"user_id": id, "error": err.Error()})
			return
		}

		// 转换为前端需要的格式
		var result []gin.H
		for _, history := range aiHistories {
			var plan interface{}
			if err := json.Unmarshal([]byte(history.GeneratedPlan), &plan); err != nil {
				// 如果解析失败，使用原始字符串
				plan = history.GeneratedPlan
			}

			result = append(result, gin.H{
				"id":            history.ID,
				"background":    history.Background,
				"goal":          history.Goal,
				"difficulty":    history.Difficulty,
				"generatedPlan": plan,
				"createdAt":     history.CreatedAt,
			})
		}

		utils.LogInfo("获取AI历史记录成功", logrus.Fields{"user_id": id, "count": len(result)})
		c.JSON(200, gin.H{"aiHistories": result})
	}
}

// 删除AI历史记录
func DeleteAIHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok || id == 0 {
			c.JSON(401, gin.H{"error": "未授权，请先登录"})
			return
		}

		var req struct {
			HistoryID uint `json:"historyId" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数格式错误"})
			utils.LogError("解析删除AI历史记录参数失败", logrus.Fields{"error": err.Error()})
			return
		}

		// 确保只能删除自己的记录
		if err := repository.DB.Where("id = ? AND user_id = ?", req.HistoryID, id).Delete(&model.AIHistory{}).Error; err != nil {
			c.JSON(500, gin.H{"error": "删除AI历史记录失败"})
			utils.LogError("删除AI历史记录失败", logrus.Fields{"user_id": id, "history_id": req.HistoryID, "error": err.Error()})
			return
		}

		utils.LogInfo("删除AI历史记录成功", logrus.Fields{"user_id": id, "history_id": req.HistoryID})
		c.JSON(200, gin.H{"success": true, "message": "AI历史记录已删除"})
	}
}
