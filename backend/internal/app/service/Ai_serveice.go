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

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	"github.com/gin-gonic/gin"
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
		planner = &TaiFuLearningPlanner{
			APIKey:  os.Getenv("APIKEY"),
			BaseURL: "https://api.siliconflow.cn/v1/chat/completions",
		}
		fmt.Printf("planner配置完成")
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

	// 输入合法性检测
	if !isValidLearningGoal(req.Flag) {
		fmt.Printf("⚠️ 检测到无效输入: %s\n", req.Flag)
		c.JSON(http.StatusBadRequest, LearningPlanResponse{
			Success: false,
			Error:   "输入内容无效，请输入有意义的学习目标（如：学习Python编程、提升英语口语等）",
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

	// 埋点：生成学习计划（不添加Flag，让前端决定）
	repository.AddTrackPointToDB(id, "生成学习计划")
	fmt.Printf("✅ 成功生成学习计划，难度: %d，计划长度: %d\n", difficulty, len(plan))
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
	systemPrompt := `你是"太傅AI学习计划生成器",专门为用户制定科学合理的学习路径。请根据用户的学习目标和个人背景,生成详细的三阶段学习计划,并自动拆解为具体可执行的Flag任务。

难度评分标准：
50分 - 入门级,适合零基础,1-2周可掌握,拆解为3-5个简单Flag
150分 - 基础级,需要一些预备知识,1个月左右,拆解为5-6个中等Flag  
200分 - 专家级,需要大量时间和实践,半年以上深度钻研,拆解为6-8个挑战Flag

注意：Flag数量必须控制在1-8个之间，确保每个Flag都有明确的可执行性

请严格按照以下JSON格式返回(不要包含markdown代码块标记):
{
	"flag": "根据用户目标生成的具体精炼标题(10-20字)",
	"difficulty": 分数(50/150/200),
	"plan": "详细的三阶段学习计划..."
}

plan字段格式要求:
1. 必须包含3个明确的阶段,每个阶段用"阶段一:"或"第一阶段:"标识
2. 每个阶段必须包含:
   - 阶段目标（该阶段要达成的核心能力）
   - 学习要点（2-4个关键知识点，详细说明学习内容）
   - 实践建议（具体的练习方法和资源推荐）
   - 时间规划（建议的学习时长和进度安排）
3. 每个阶段下生成2-3个具体的、可执行的Flag任务
4. 任务必须用数字或符号标记(如"1. "、"- "、"• ")
5. 任务描述要具体可执行,包含明确的完成标准
6. 总共生成的任务数量控制在1-8个之间

示例格式:
阶段一:基础入门（预计1-2周）
【阶段目标】掌握Python基础语法，能够编写简单程序
【学习要点】
- 变量、数据类型（整数、浮点数、字符串、布尔值）
- 基本运算符和表达式
- 条件语句（if-elif-else）和循环（for/while）
- 函数定义和调用
【实践建议】
- 推荐资源：Python官方教程、菜鸟教程
- 每天编写2-3个小程序巩固知识点
- 使用在线编程平台（如LeetCode入门题）练习
【具体任务】
1. 完成Python语法基础教程前5章，并做笔记
2. 编写10个基础练习程序（变量、循环、函数各3个）

阶段二:进阶学习（预计2-3周）
【阶段目标】掌握Python核心数据结构和面向对象编程
【学习要点】
- 列表、元组、字典、集合的使用和常用方法
- 字符串处理和正则表达式
- 文件读写操作
- 面向对象编程：类、对象、继承、多态
【实践建议】
- 通过实际案例理解数据结构的应用场景
- 编写小工具来练习文件操作（如批量重命名）
- 设计简单的类来建模现实问题
【具体任务】
1. 掌握列表和字典操作，完成20道相关练习题
2. 编写一个简单的学生成绩管理系统（使用类和文件操作）

阶段三:项目实战（预计2-4周）
【阶段目标】独立完成完整项目，建立编程自信
【学习要点】
- 项目规划和模块划分
- 代码组织和注释规范
- 调试技巧和错误处理
- 第三方库的使用（如requests、pandas）
【实践建议】
- 从简单项目开始，逐步增加复杂度
- 使用Git进行版本控制
- 参考GitHub上的优秀开源项目
【具体任务】
1. 开发一个实用工具（计算器、待办清单或天气查询应用）
2. 总结学习笔记，整理知识脑图，分享学习心得`

	// 构建用户提示词
	userPrompt := fmt.Sprintf("学习目标: %s\n", req.Flag)
	if req.Background != "" {
		userPrompt += fmt.Sprintf("个人背景: %s\n", req.Background)
	}
	if req.Difficulty != 0 {
		userPrompt += fmt.Sprintf("期望难度分数: %d\n", req.Difficulty)
	}
	userPrompt += "\n请根据以上信息生成学习计划,返回标准JSON格式。"

	fmt.Printf("📋 系统提示: %s\n", systemPrompt)
	fmt.Printf("📋 用户提示: %s\n", userPrompt)

	// 调用AI
	response, err := p.callOpenAI(systemPrompt, userPrompt)
	if err != nil {
		fmt.Printf("❌ AI调用失败: %v\n", err)
		return "", "", 0, err
	}

	fmt.Printf("✅ AI返回成功,原始响应长度: %d\n", len(response))

	// 解析AI响应
	flag, plan, difficulty, err := p.parseAIResponse(response)
	if err != nil {
		fmt.Printf("❌ 解析AI响应失败: %v\n", err)
		return "", "", 0, err
	}

	// 验证结果
	if plan == "" {
		return "", "", 0, fmt.Errorf("AI返回的学习计划为空")
	}
	if difficulty == 0 {
		difficulty = req.Difficulty // 使用请求的难度作为默认值
	}

	fmt.Printf("✅ 解析成功,难度: %d, 计划长度: %d\n", difficulty, len(plan))
	return flag, plan, difficulty, nil
}

// 解析AI响应
func (p *TaiFuLearningPlanner) parseAIResponse(response string) (string, string, int, error) {
	// 清理响应（移除可能的markdown代码块标记）
	cleanResponse := response
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimPrefix(cleanResponse, "```")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	// 尝试解析JSON响应
	var result struct {
		Flag       string `json:"flag"`
		Difficulty int    `json:"difficulty"`
		Plan       string `json:"plan"`
	}

	err := json.Unmarshal([]byte(cleanResponse), &result)
	if err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		fmt.Printf("尝试解析的内容前100字符: %s\n", cleanResponse[:min(100, len(cleanResponse))])

		// 如果解析失败,返回原始响应作为计划
		return "", cleanResponse, 0, nil
	}

	// 验证必要字段
	if result.Plan == "" {
		fmt.Printf("⚠️ 解析的计划为空,使用原始响应\n")
		return result.Flag, cleanResponse, result.Difficulty, nil
	}

	fmt.Printf("✅ 成功解析: flag=%s, difficulty=%d, plan长度=%d\n",
		result.Flag, result.Difficulty, len(result.Plan))

	return result.Flag, result.Plan, result.Difficulty, nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
		"model": "Qwen/Qwen3-VL-30B-A3B-Instruct",
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
