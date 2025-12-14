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
		apiKey := os.Getenv("APIKEY")
		if apiKey == "" {
			fmt.Printf("❌ 警告：APIKEY环境变量未设置\n")
		} else {
			fmt.Printf("✅ API Key已加载，前缀: %s...\n", apiKey[:min(10, len(apiKey))])
		}

		planner = &TaiFuLearningPlanner{
			APIKey:  apiKey,
			BaseURL: "https://api.siliconflow.cn/v1/chat/completions",
		}
		fmt.Printf("✅ planner配置完成\n")
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

	fmt.Printf("📝 收到计划生成请求: %+v\n", req)

	// 生成计划
	flag, plan, difficulty, err := planner.GenerateLearningPlan(req)
	if err != nil {
		fmt.Printf("❌ 生成计划失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, LearningPlanResponse{
			Success: false,
			Error:   fmt.Sprintf("生成计划失败: %v", err),
		})
		return
	}

	// 埋点：生成计划（不添加Flag，让前端决定）
	repository.AddTrackPointToDB(id, "生成AI计划")
	fmt.Printf("✅ 成功生成计划，难度: %d，计划长度: %d\n", difficulty, len(plan))
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
3. 总结学习心得并分享（每日完成：1次）

阶段二：进阶学习（预计1-2周）
【阶段目标】掌握Python核心数据结构和面向对象编程，解决实际问题。
【行动要点】
- 列表、元组、字典、集合的使用，理解适用场景。
- 字符串处理和正则表达式，提升文本处理能力。
- 文件读写和数据持久化，掌握数据管理。
- 面向对象编程：类、对象、继承、多态。
【实践建议】
通过案例理解数据结构应用，编写小工具练习文件操作，设计类建模现实问题。
1. 完成列表和字典练习题，整理常见错误（每日完成：2次）
2. 开发学生成绩管理系统（使用类和文件）（每日完成：1次）
3. 参与线上编程挑战，分享代码心得（每日完成：1次）

阶段三：项目实战（预计1-2月）
【阶段目标】独立完成完整项目，建立编程自信，具备团队协作能力。
【行动要点】
- 项目规划和模块化设计，合理分工与进度管理。
- 代码规范和注释标准，提升团队协作效率。
- 调试技巧和异常处理，减少bug提升稳定性。
- 第三方库使用（requests、pandas等），扩展功能。
【实践建议】
从简单项目逐步增加复杂度，使用Git版本控制，参考GitHub开源项目学习。
1. 开发实用工具（计算器/待办清单/天气应用）并撰写使用手册（每日完成：2次）
2. 整理学习笔记和知识脑图并分享，收集反馈（每日完成：1次）
3. 参与开源项目贡献代码，记录成长历程（每日完成：1次）"
}

【健康类示例】
{
    "flag": "早睡早起养成计划",
    "difficulty": 100,
    "plan": "【目标概述】调整作息规律，建立健康睡眠习惯，提升精神状态和生活质量。

阶段一：适应期（预计1-3天）
【阶段目标】逐步调整生物钟，初步建立早睡早起意识，减轻身体不适感。
【行动要点】
- 固定睡眠时间窗口（23:00-7:00），保证8小时睡眠。
- 睡前1小时避免电子设备，放松身心准备入睡。
- 早晨使用闹钟并立即起床，避免赖床。
【实践建议】
可以设置多个闹钟，并把闹钟放在离床较远的位置。睡前可以读纸质书、听轻音乐放松，营造良好睡眠环境。
1. 在23:00前洗漱完毕并上床（每日完成：1次）
2. 7:00准时起床并开窗通风（每日完成：1次）
3. 记录当天睡眠质量和精神状态（每日完成：1次）

阶段二：巩固期（预计1-2周）
【阶段目标】稳定作息习惯，身体适应新节奏，减少对意志力的依赖。
【行动要点】
- 培养睡前仪式感（洗澡、护肤、冥想等），形成条件反射。
- 增加晨间活动（运动、早餐、晨读），强化早起动力。
- 午睡控制在20-30分钟，避免影响夜间睡眠。
【实践建议】
可以加入早起打卡群互相监督，设计奖励机制激励自己。白天增加户外活动，帮助夜晚更好入睡。
1. 完成睡前固定流程（洗澡-护肤-冥想）（每日完成：1次）
2. 早起后完成30分钟晨间活动（运动/阅读）（每日完成：1次）
3. 总结一天作息情况并调整计划（每日完成：1次）

阶段三：习惯养成期（预计1-2月）
【阶段目标】将早睡早起内化为生活方式，无需刻意坚持即可自然执行。
【行动要点】
- 根据季节调整作息时间，保持灵活性。
- 建立应对特殊情况的补救措施（偶尔晚睡如何快速恢复）。
- 扩展健康生活习惯（规律运动、健康饮食），形成良性循环。
【实践建议】
定期回顾身体变化和精神状态改善，强化习惯的正向反馈。可以记录睡眠日记，分析影响睡眠质量的因素。
1. 持续保持固定作息，允许周末±30分钟弹性（每日完成：1次）
2. 每周总结作息规律和健康改善情况（每日完成：1次）
3. 分享早睡早起心得，帮助他人建立习惯（每日完成：1次）"
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
		if difficulty == 0 {
			difficulty = 150 // 默认中等难度
		}
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

	// 准备请求数据 - 使用标准格式
	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	requestData := map[string]interface{}{
		"model": "Qwen/Qwen2.5-7B-Instruct", // 使用更稳定的模型
		"messages": []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		"max_tokens":  4000, // 增加token数，确保完整输出
		"temperature": 0.7,
		"stream":      false,
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
