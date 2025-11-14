package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	"github.com/gin-gonic/gin"
)

// 学习计划请求
type LearningPlanRequest struct {
	Flag       string `json:"flag" binding:"required"` // 学习目标标识
	Background string `json:"background,omitempty"`    // 用户背景
	Level      int    `json:"preferences,omitempty"`   // 学习偏好
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
		planner = &TaiFuLearningPlanner{
			APIKey:  os.Getenv("APIKEY"),
			BaseURL: "https://api.siliconflow.cn/v1/chat/completions",
		}
		fmt.Printf("planner配置完成")
	}
}

func GenerateLearningPlan(c *gin.Context) {
	// 确保 planner 已初始化
	initPlanner()
	id, _ := getCurrentUserID(c)
	var req LearningPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ 请求格式错误: %v\n", err)
		c.JSON(http.StatusBadRequest, LearningPlanResponse{
			Success: false,
			Error:   fmt.Sprintf("请求格式错误: %v", err),
		})
		return
	}

	fmt.Printf("📝 收到学习计划请求: %+v\n", req)

	// 生成学习计划
	flag, plan, difficulty, err := planner.GenerateLearningPlan(req)
	if err != nil {
		fmt.Printf("❌ 生成学习计划失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, LearningPlanResponse{
			Success: false,
			Error:   fmt.Sprintf("生成学习计划失败: %v", err),
		})
		return
	}
	repository.AddFlagToDB(id, model.Flag{
		Title:     req.Flag,
		Detail:    plan,
		CreatedAt: time.Now(),
		IsPublic:  true, // AI生成的Flag默认公开
	})
	//埋点
	repository.AddTrackPointToDB(id, "生成学习计划")
	fmt.Printf("✅ 成功生成学习计划，难度: %d\n", difficulty)
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
	systemPrompt := `你是"太傅AI学习计划生成器"，专门为用户制定科学合理的学习路径。请根据用户的学习目标(flag)生成详细的三阶段学习计划，并评估难度等级(1-5分)。

难度评分标准：
50分 - 入门级，适合零基础，1-2周可掌握
150分 - 基础级，需要一些预备知识，1个月左右
200分 - 专家级，需要大量时间和实践，半年以上深度钻研

请严格按照以下JSON格式返回，不要包含其他内容：
{
	"flag": "按照大致方向生成具体的flag目标",
	"difficulty": 分数,
	"plan": "学习几乎详细的三阶段学习计划内容"
}`

	// 构建用户提示词
	userPrompt := fmt.Sprintf("学习目标: %s\n", req.Flag)
	if req.Background != "" {
		userPrompt += fmt.Sprintf("用户背景: %s\n", req.Background)
	}
	if req.Level != 0 {
		userPrompt += fmt.Sprintf("学习偏好等级: %d\n", req.Level)
	}

	fmt.Printf("📋 系统提示: %s\n", systemPrompt)
	fmt.Printf("📋 用户提示: %s\n", userPrompt)

	// 调用AI
	response, err := p.callOpenAI(systemPrompt, userPrompt)
	if err != nil {
		fmt.Printf("❌ AI调用失败: %v\n", err)
		return "", "", 0, err
	}

	fmt.Printf("✅ AI返回成功\n")

	// 解析AI响应
	flag, plan, difficulty, err := p.parseAIResponse(response)
	if err != nil {
		fmt.Printf("❌ 解析AI响应失败: %v\n", err)
		return "", "", 0, err
	}

	fmt.Printf("✅ 解析成功，难度: %d\n", difficulty)
	return flag, plan, difficulty, nil
}

// 解析AI响应
func (p *TaiFuLearningPlanner) parseAIResponse(response string) (string, string, int, error) {
	// 尝试解析JSON响应
	var result struct {
		Flag string `json:"flag"`

		Difficulty int    `json:"difficulty"`
		Plan       string `json:"plan"`
	}

	err := json.Unmarshal([]byte(response), &result)
	if err != nil {
		fmt.Printf("❌ 解析失败，返回原始响应: %v\n", err)
		// 如果解析失败，返回原始响应作为计划
		return "", response, 3, nil
	}

	if result.Plan == "" {
		fmt.Printf("⚠️ 解析的计划为空\n")
		return "", response, result.Difficulty, nil
	}

	return result.Flag, result.Plan, result.Difficulty, nil
}

// 调用OpenAI API
func (p *TaiFuLearningPlanner) callOpenAI(systemPrompt, userPrompt string) (string, error) {
	// 检查API密钥
	fmt.Printf("🔍 检查API密钥...\n")
	fmt.Printf("API密钥: %s\n", p.APIKey)
	fmt.Printf("BaseURL: %s\n", p.BaseURL)

	if p.APIKey == "" {
		fmt.Printf("❌ API密钥为空\n")
		return "", fmt.Errorf("❌ API密钥未配置，请检查环境变量 APIKEY")
	}

	// 准备请求数据
	requestData := map[string]interface{}{
		"model": "Qwen/Qwen2.5-Coder-32B-Instruct",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens":  3000,
		"temperature": 0.3,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("❌ 序列化请求失败: %v\n", err)
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	fmt.Printf("📤 发送请求到: %s\n", p.BaseURL)
	fmt.Printf("📄 请求体: %s\n", string(requestBody))

	req, err := http.NewRequest("POST", p.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	// 调试：打印 Authorization header
	fmt.Printf("📍 Authorization Header: %s\n", req.Header.Get("Authorization"))

	// 创建带 60 秒超时的客户端（给 AI 充足时间响应）
	client := &http.Client{Timeout: 60 * time.Second}
	fmt.Printf("⏱️ 开始调用 SiliconFlow API...\n")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 发送请求失败: %v\n", err)
		return "", fmt.Errorf("❌ 发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📥 收到响应，状态码: %d\n", resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// ✅ 打印原始响应
	fmt.Printf("📝 API原始响应: %s\n", string(body))

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		} `json:"data"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		fmt.Printf("原始响应内容: %s\n", string(body))
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查是否有错误码
	if response.Code != 0 && response.Code != 200 {
		fmt.Printf("❌ SiliconFlow API 返回错误码: %d, 消息: %s\n", response.Code, response.Message)
		return "", fmt.Errorf("❌ SiliconFlow API 错误: %s (错误码: %d)", response.Message, response.Code)
	}

	if response.Error.Message != "" {
		fmt.Printf("❌ SiliconFlow API 错误字段: %s\n", response.Error.Message)
		return "", fmt.Errorf("❌ SiliconFlow API 错误: %s", response.Error.Message)
	}

	// 优先检查 Data 中的 Choices（某些版本 API）
	if len(response.Data.Choices) > 0 {
		content := response.Data.Choices[0].Message.Content
		fmt.Printf("✅ AI 返回内容 (从 data): %s\n", content)
		return content, nil
	}

	// 备选：检查顶级的 Choices
	if len(response.Choices) > 0 {
		content := response.Choices[0].Message.Content
		fmt.Printf("✅ AI 返回内容 (从 choices): %s\n", content)
		return content, nil
	}

	fmt.Printf("❌ 未收到有效的 AI 响应，响应结构: %+v\n", response)
	return "", fmt.Errorf("未收到有效的 AI 响应")
}
