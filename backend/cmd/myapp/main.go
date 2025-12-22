package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/handler"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/service"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 首先加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("警告: 加载 .env 文件失败: %v", err)
	}

	// 🔧 阶段5：配置日志级别和模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug" // 默认开发模式
	}
	gin.SetMode(ginMode)

	if ginMode == "release" {
		// 生产环境：关闭Gin的调试日志
		gin.DisableConsoleColor()
	}

	// 数据库连接
	if err := repository.DBconnect(); err != nil {
		log.Fatalf("❌ 数据库连接失败，程序退出: %v", err)
	}

	service.Init()                   //初始化每天学习时间记录
	service.StartHitokotoScheduler() //启动一言定时任务
	r := gin.Default()

	// Panic Recovery 中间件（防止服务器崩溃）
	r.Use(gin.Recovery())

	// 请求记录中间件
	r.Use(func(c *gin.Context) {
		utils.LogInfo("请求开始", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		})
		c.Next()
		utils.LogInfo("请求完成", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": c.Writer.Status(),
		})
	})

	// 添加全局 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 静态文件服务 - 提供前端头像访问
	// 优先检查本地开发环境路径，然后是生产环境路径
	assetsPath := "../frontend/src/assets"
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		// 如果本地路径不存在，尝试使用生产环境路径
		assetsPath = "./assets"
		if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
			// 如果都不存在，使用相对于可执行文件的路径
			execPath, _ := os.Executable()
			assetsPath = filepath.Join(filepath.Dir(execPath), "assets")
		}
	}
	r.Static("/assets", assetsPath)
	utils.LogInfo("静态文件服务启动成功", map[string]interface{}{
		"route": "/assets",
		"path":  assetsPath,
	})

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "unimate-backend",
		})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	handler.BasicUser(r) //用户相关
	utils.LogInfo("服务器启动成功", nil)
	handler.Flag(r) //签到相关
	utils.LogInfo("签到模块加载成功", nil)
	handler.BasicPost(r) //帖子相关
	utils.LogInfo("帖子模块加载成功", nil)
	handler.BasicFlag(r)
	utils.LogInfo("Flag模块加载成功", nil)
	handler.ChatWebSocket(r) //聊天相关
	utils.LogInfo("聊天模块加载成功", nil)
	handler.Ranking(r) //封神榜相关
	utils.LogInfo("封神榜模块加载成功", nil)
	handler.Search(r) //搜索相关
	utils.LogInfo("搜索模块加载成功", nil)
	handler.LearnTime(r) //学习时长相关
	utils.LogInfo("学习时长模块加载成功", nil)
	handler.Achievement(r) //成就相关
	utils.LogInfo("成就模块加载成功", nil)
	handler.AI(r) //AI学习计划
	utils.LogInfo("AI模块加载成功", nil)

	// 🔧 阶段5：优雅关闭机制
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 在goroutine中启动服务器
	go func() {
		utils.LogInfo("服务器启动", map[string]interface{}{
			"port": port,
			"env":  os.Getenv("GIN_MODE"),
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.LogInfo("服务器正在关闭...", nil)

	// 给服务器5秒时间完成正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器强制关闭: %v", err)
	}

	utils.LogInfo("服务器已安全退出", nil)
}
