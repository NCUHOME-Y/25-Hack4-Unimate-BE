package main

import (
	"log"
	"os"
	"path/filepath"

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

	repository.DBconnect() //数据库连接
	service.Init()         //初始化每天学习时间记录
	r := gin.Default()

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
	handler.Ranking(r) //排行榜相关
	utils.LogInfo("排行榜模块加载成功", nil)
	handler.Search(r) //搜索相关
	utils.LogInfo("搜索模块加载成功", nil)
	handler.LearnTime(r) //学习时长相关
	utils.LogInfo("学习时长模块加载成功", nil)
	handler.Achievement(r) //成就相关
	utils.LogInfo("成就模块加载成功", nil)
	handler.AI(r) //AI学习计划
	utils.LogInfo("AI模块加载成功", nil)
	// TODO: 实现这些函数后再启用
	// handler.ChatHistory(r) //聊天历史 // P1修复：聊天历史和房间管理
	// utils.LogInfo("聊天历史模块加载成功", nil)
	// handler.PostRESTful(r) //RESTful帖子 // P1修复：RESTful风格帖子接口
	// utils.LogInfo("RESTful帖子模块加载成功", nil)
	r.Run("0.0.0.0:8080")
	utils.LogInfo("服务器运行中，监听端口8080", nil)
}
