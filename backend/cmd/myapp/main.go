package main

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/handler"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/service"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
)

func main() {
	repository.DBconnect() //数据库连接
	service.Init()         //初始化每天学习时间记录
	r := gin.Default()
	handler.BasicUser(r) //用户相关
	utils.LogInfo("服务器启动成功", nil)
	handler.Flag(r) //签到相关
	utils.LogInfo("签到模块加载成功", nil)
	handler.BasicPost(r) //帖子相关
	utils.LogInfo("帖子模块加载成功", nil)
	handler.ChatWebSocket(r) //聊天相关
	utils.LogInfo("聊天模块加载成功", nil)
	handler.Ranking(r) //排行榜相关
	utils.LogInfo("排行榜模块加载成功", nil)
	handler.LearnTime(r) //学习时长相关
	utils.LogInfo("学习时长模块加载成功", nil)
	handler.Achievement(r) //成就相关
	utils.LogInfo("成就模块加载成功", nil)
	handler.AiService(r)
	utils.LogInfo("AI服务模块加载成功", nil)
	r.Run("0.0.0.0:8080")
	utils.LogInfo("服务器运行中，监听端口8080", nil)
}
