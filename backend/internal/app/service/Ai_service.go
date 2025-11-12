package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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

var planner = &TaiFuLearningPlanner{
	APIKey:  os.Getenv("OPENAI_API_KEY"),
	BaseURL: "https://api.openai.com/v1/chat/completions",
}

func GenerateLearningPlan(c *gin.Context) {
	var req LearningPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LearningPlanResponse{
			Success: false,
			Error:   fmt.Sprintf("请求格式错误: %v", err),
		})
		return
	}

	// 生成学习计划
	plan, difficulty, err := planner.GenerateLearningPlan(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, LearningPlanResponse{
			Success: false,
			Error:   fmt.Sprintf("生成学习计划失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, LearningPlanResponse{
		Success: true,
		Flag:    req.Flag,
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
func (p *TaiFuLearningPlanner) GenerateLearningPlan(req LearningPlanRequest) (string, int, error) {
	// 构建系统提示词
	systemPrompt := `你是"太傅AI学习计划生成器"，专门为用户制定科学合理的学习路径。请根据用户的学习目标(flag)生成详细的三阶段学习计划，并评估难度等级(1-5分)。

难度评分标准：
1分 - 入门级，适合零基础，1-2周可掌握
2分 - 基础级，需要一些预备知识，1个月左右
3分 - 进阶级，需要扎实基础，2-3个月系统学习
5分 - 专家级，需要大量时间和实践，半年以上深度钻研

请严格按照以下JSON格式返回，不要包含其他内容：
{
	"difficulty": 分数,
	"plan": {
		"阶段一": "详细描述阶段一的学习内容和目标",
		"阶段二": "详细描述阶段二的学习内容和目标",
		"阶段三": "详细描述阶段三的学习内容和目标"
	}
}`

	// 构建用户提示词
	userPrompt := fmt.Sprintf("学习目标: %s\n", req.Flag)
	if req.Background != "" {
		userPrompt += fmt.Sprintf("用户背景: %s\n", req.Background)
	}
	if req.Level != 0 {
		userPrompt += fmt.Sprintf("学习偏好等级: %d\n", req.Level)
	}

	// 调用AI
	response, _ := p.callOpenAI(systemPrompt, userPrompt)
	// 解析AI响应
	return p.parseAIResponse(response)
}

// 解析AI响应
func (p *TaiFuLearningPlanner) parseAIResponse(response string) (string, int, error) {
	// 尝试解析JSON响应
	var result struct {
		Difficulty int    `json:"difficulty"`
		Plan       string `json:"plan"`
	}

	json.Unmarshal([]byte(response), &result)
	return result.Plan, result.Difficulty, nil
}

// 调用OpenAI API
func (p *TaiFuLearningPlanner) callOpenAI(systemPrompt, userPrompt string) (string, error) {
	// 准备请求数据
	requestData := map[string]interface{}{
		"model": "Qwen / Qwen2.5 Coder 32B Instruct",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens":  3000,
		"temperature": 0.3,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", p.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var response struct {
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
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("OpenAI API错误: %s", response.Error.Message)
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("未收到有效的AI响应")
}
