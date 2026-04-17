package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
	Provider         string
	APIKey           string
	BaseURL          string
	Models           []string
	Timeout          time.Duration
	RetryPerEndpoint int
	Region           string
	AllowCrossRegion bool
}

var planner *TaiFuLearningPlanner

// 初始化 planner（延迟初始化，等待 .env 加载）
func initPlanner() {
	planner = loadPlannerFromEnv()
	if planner.APIKey == "" {
		logrus.Error("E_AI_CONFIG: 未设置 API key（AI_API_KEY 或 APIKEY）")
	}
}

func loadPlannerFromEnv() *TaiFuLearningPlanner {
	provider := strings.ToLower(strings.TrimSpace(getEnv("AI_PROVIDER", "dashscope")))
	apiKey := strings.TrimSpace(getEnv("AI_API_KEY", strings.TrimSpace(os.Getenv("APIKEY"))))
	region := strings.ToLower(strings.TrimSpace(getEnv("DASHSCOPE_REGION", "cn")))

	defaultBaseURL := defaultBaseURLByProvider(provider, region)
	baseURL := strings.TrimSpace(getEnv("AI_BASE_URL", strings.TrimSpace(os.Getenv("DASHSCOPE_BASE_URL"))))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	models := parseCSV(getEnv("AI_MODELS", strings.Join(defaultModelsByProvider(provider), ",")))
	if len(models) == 0 {
		models = defaultModelsByProvider(provider)
	}

	timeoutSec := parsePositiveInt(getEnv("AI_TIMEOUT_SECONDS", "10"), 10)
	retryPerEndpoint := parsePositiveInt(getEnv("AI_RETRY_PER_ENDPOINT", "2"), 2)
	allowCrossRegion := parseBool(getEnv("AI_ALLOW_CROSS_REGION_FALLBACK", "true"), true)

	return &TaiFuLearningPlanner{
		Provider:         provider,
		APIKey:           apiKey,
		BaseURL:          baseURL,
		Models:           models,
		Timeout:          time.Duration(timeoutSec) * time.Second,
		RetryPerEndpoint: retryPerEndpoint,
		Region:           region,
		AllowCrossRegion: allowCrossRegion,
	}
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func parseCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parsePositiveInt(v string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseBool(v string, fallback bool) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return fallback
	}
	if v == "1" || v == "true" || v == "yes" || v == "on" {
		return true
	}
	if v == "0" || v == "false" || v == "no" || v == "off" {
		return false
	}
	return fallback
}

func defaultModelsByProvider(provider string) []string {
	switch provider {
	case "openrouter":
		return []string{"qwen/qwen3-next-80b-a3b-instruct:free", "openrouter/auto"}
	case "zhipu", "bigmodel", "glm":
		return []string{"glm-4.7-flash", "glm-4-flash-250414"}
	case "openai":
		return []string{"gpt-4o-mini"}
	default:
		return []string{"qwen3.5-flash", "qwen-plus"}
	}
}

func defaultBaseURLByProvider(provider, region string) string {
	switch provider {
	case "openrouter":
		return "https://openrouter.ai/api/v1/chat/completions"
	case "zhipu", "bigmodel", "glm":
		return "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	case "openai":
		return "https://api.openai.com/v1/chat/completions"
	default:
		if region == "intl" || region == "international" {
			return "https://dashscope-intl.aliyuncs.com/compatible-mode/v1/chat/completions"
		}
		return "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
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
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "E_AI_CONFIG"):
			c.JSON(http.StatusInternalServerError, LearningPlanResponse{
				Success: false,
				Error:   errMsg,
			})
			return
		case strings.Contains(errMsg, "E_AI_REGION_MISMATCH"):
			c.JSON(http.StatusBadGateway, LearningPlanResponse{
				Success: false,
				Error:   errMsg,
			})
			return
		case strings.Contains(errMsg, "E_AI_AUTH"):
			c.JSON(http.StatusBadGateway, LearningPlanResponse{
				Success: false,
				Error:   errMsg,
			})
			return
		case strings.Contains(errMsg, "E_AI_RATE_LIMIT"):
			c.JSON(http.StatusTooManyRequests, LearningPlanResponse{
				Success: false,
				Error:   errMsg,
			})
			return
		case strings.Contains(errMsg, "E_AI_TIMEOUT"), strings.Contains(errMsg, "E_AI_UPSTREAM_5XX"):
			c.JSON(http.StatusServiceUnavailable, LearningPlanResponse{
				Success: false,
				Error:   errMsg,
			})
			return
		default:
			c.JSON(http.StatusBadGateway, LearningPlanResponse{
				Success: false,
				Error:   errMsg,
			})
			return
		}
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
	if p == nil {
		return "", fmt.Errorf("E_AI_CONFIG: planner is nil")
	}
	if strings.TrimSpace(p.APIKey) == "" {
		return "", fmt.Errorf("E_AI_CONFIG: 未设置 API key，请配置 AI_API_KEY 或 APIKEY")
	}
	models := p.Models
	if len(models) == 0 {
		models = defaultModelsByProvider(p.Provider)
	}

	var lastErr error
	for _, modelName := range models {
		response, err := p.callOpenAIWithModel(modelName, systemPrompt, userPrompt)
		if err == nil {
			return response, nil
		}
		lastErr = err

		if strings.Contains(err.Error(), "E_AI_AUTH") || strings.Contains(err.Error(), "E_AI_REGION_MISMATCH") {
			return "", err
		}

		// 上游限流或超时时，尝试切换备用模型。
		if strings.Contains(err.Error(), "E_AI_RATE_LIMIT") || strings.Contains(err.Error(), "E_AI_TIMEOUT") || strings.Contains(err.Error(), "E_AI_UPSTREAM_5XX") {
			logrus.WithFields(logrus.Fields{
				"provider": p.Provider,
				"model":    modelName,
				"error":    err.Error(),
			}).Warn("主模型不可用，尝试切换备用模型")
			continue
		}

		// 非限流错误直接返回，避免掩盖配置类问题（如 key 无效）。
		return "", err
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("E_AI_UNKNOWN: AI请求失败，未拿到任何可用响应")
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
		return "", fmt.Errorf("E_AI_CONFIG: 请求序列化失败: %v", err)
	}
	client := &http.Client{Timeout: p.Timeout}
	endpoints := p.resolveEndpoints()

	var lastErr error
	hasTimeoutErr := false
	hasInvalidKeyErr := false
	hasRateLimitErr := false
	for _, endpoint := range endpoints {
		for attempt := 1; attempt <= p.RetryPerEndpoint; attempt++ {
			req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
			if err != nil {
				return "", fmt.Errorf("E_AI_CONFIG: 创建请求失败: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
			if p.Provider == "openrouter" {
				req.Header.Set("HTTP-Referer", getEnv("OPENROUTER_HTTP_REFERER", "https://zhixu.online"))
				req.Header.Set("X-Title", getEnv("OPENROUTER_APP_NAME", "UniMate"))
			}
			if p.APIKey != "" {
				req.Header.Set("Authorization", "Bearer "+p.APIKey)
			}

			resp, err := client.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("E_AI_TIMEOUT: provider=%s model=%s endpoint=%s attempt=%d timeout=%s error=%v", p.Provider, modelName, endpoint, attempt, p.Timeout.String(), err)
				if strings.Contains(err.Error(), "context deadline exceeded") && attempt < p.RetryPerEndpoint {
					time.Sleep(time.Duration(attempt) * 700 * time.Millisecond)
					continue
				}
				if strings.Contains(err.Error(), "context deadline exceeded") {
					hasTimeoutErr = true
					logrus.WithFields(logrus.Fields{
						"provider": p.Provider,
						"model":    modelName,
						"endpoint": endpoint,
						"error":    err.Error(),
					}).Warn("模型接口超时，尝试下一个端点")
					break
				}
				return "", lastErr
			}

			bodyBytes, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = fmt.Errorf("E_AI_UPSTREAM_READ: provider=%s model=%s endpoint=%s status=%d error=%v", p.Provider, modelName, endpoint, resp.StatusCode, readErr)
				return "", lastErr
			}

			if resp.StatusCode == 200 {
				return string(bodyBytes), nil
			}

			if resp.StatusCode == 401 && (strings.Contains(string(bodyBytes), "invalid_api_key") || strings.Contains(string(bodyBytes), "Incorrect API key provided")) {
				hasInvalidKeyErr = true
				lastErr = fmt.Errorf("E_AI_AUTH: provider=%s model=%s endpoint=%s status=%d body=%s", p.Provider, modelName, endpoint, resp.StatusCode, compactForLog(string(bodyBytes), 500))
				logrus.WithFields(logrus.Fields{
					"provider": p.Provider,
					"model":    modelName,
					"endpoint": endpoint,
					"status":   resp.StatusCode,
				}).Warn("端点鉴权失败，尝试下一个端点")
				break
			}

			if resp.StatusCode == 429 {
				hasRateLimitErr = true
				return "", fmt.Errorf("E_AI_RATE_LIMIT: provider=%s model=%s endpoint=%s status=%d body=%s", p.Provider, modelName, endpoint, resp.StatusCode, compactForLog(string(bodyBytes), 500))
			}

			code, message, requestID := parseUpstreamError(bodyBytes)
			apiErr := fmt.Errorf("E_AI_UPSTREAM_%d: provider=%s model=%s endpoint=%s status=%d upstream_code=%s request_id=%s message=%s body=%s", resp.StatusCode, p.Provider, modelName, endpoint, resp.StatusCode, code, requestID, message, compactForLog(string(bodyBytes), 500))
			lastErr = apiErr

			// 5xx通常是上游拥塞，短暂退避后重试。
			if resp.StatusCode >= 500 && attempt < p.RetryPerEndpoint {
				time.Sleep(time.Duration(attempt) * 700 * time.Millisecond)
				continue
			}

			return "", apiErr
		}
	}

	if hasInvalidKeyErr {
		if hasTimeoutErr {
			return "", fmt.Errorf("E_AI_REGION_MISMATCH: provider=%s model=%s detail=主端点超时且备用端点鉴权失败，建议检查 DASHSCOPE_REGION 或 AI_BASE_URL 与 key 来源一致", p.Provider, modelName)
		}
		return "", fmt.Errorf("E_AI_AUTH: provider=%s model=%s detail=API key 无效或未开通对应模型权限", p.Provider, modelName)
	}

	if hasRateLimitErr {
		return "", fmt.Errorf("E_AI_RATE_LIMIT: provider=%s model=%s detail=触发上游限流，请降低频率或升级配额", p.Provider, modelName)
	}

	if lastErr != nil {
		return "", lastErr
	}

	return "", fmt.Errorf("E_AI_UNKNOWN: provider=%s model=%s detail=所有端点尝试后仍失败", p.Provider, modelName)
}

func (p *TaiFuLearningPlanner) resolveEndpoints() []string {
	cnEndpoint := "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	intlEndpoint := "https://dashscope-intl.aliyuncs.com/compatible-mode/v1/chat/completions"

	var endpoints []string
	add := func(values ...string) {
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			exists := false
			for _, existing := range endpoints {
				if existing == value {
					exists = true
					break
				}
			}
			if !exists {
				endpoints = append(endpoints, value)
			}
		}
	}

	if customList := parseCSV(os.Getenv("AI_ENDPOINTS")); len(customList) > 0 {
		add(customList...)
		return endpoints
	}

	if p != nil && p.Provider != "dashscope" {
		add(p.BaseURL)
		return endpoints
	}

	region := strings.ToLower(strings.TrimSpace(os.Getenv("DASHSCOPE_REGION")))
	customBaseURL := strings.TrimSpace(os.Getenv("DASHSCOPE_BASE_URL"))
	if customBaseURL != "" {
		add(customBaseURL)
	}

	if p != nil && strings.TrimSpace(p.BaseURL) != "" {
		add(p.BaseURL)
	}

	switch region {
	case "intl", "international":
		add(intlEndpoint)
		if p == nil || p.AllowCrossRegion {
			add(cnEndpoint)
		}
	case "cn", "china", "domestic":
		add(cnEndpoint)
		if p == nil || p.AllowCrossRegion {
			add(intlEndpoint)
		}
	default:
		add(cnEndpoint)
		if p == nil || p.AllowCrossRegion {
			add(intlEndpoint)
		}
	}

	return endpoints
}

func parseUpstreamError(body []byte) (string, string, string) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", ""
	}

	code := ""
	message := ""
	requestID := ""

	if v, ok := payload["request_id"].(string); ok {
		requestID = v
	}
	if v, ok := payload["requestId"].(string); ok && requestID == "" {
		requestID = v
	}

	if e, ok := payload["error"].(map[string]interface{}); ok {
		if v, ok := e["code"].(string); ok {
			code = v
		}
		if v, ok := e["message"].(string); ok {
			message = v
		}
	}

	if code == "" {
		if v, ok := payload["code"].(string); ok {
			code = v
		}
	}
	if message == "" {
		if v, ok := payload["message"].(string); ok {
			message = v
		}
	}

	return code, message, requestID
}

func compactForLog(input string, maxLen int) string {
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")
	input = strings.TrimSpace(input)
	if maxLen <= 0 || len(input) <= maxLen {
		return input
	}
	return input[:maxLen] + "...(truncated)"
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
