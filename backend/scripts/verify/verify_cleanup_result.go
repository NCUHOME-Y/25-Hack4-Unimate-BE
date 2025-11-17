package main

import (
	"fmt"
	"log"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("警告: 加载 .env 文件失败: %v", err)
	}

	// 连接数据库
	repository.DBconnect()

	// 获取所有用户
	users, err := repository.GetAllUser()
	if err != nil {
		log.Fatalf("获取用户列表失败: %v", err)
	}

	fmt.Println("=== 成就清理结果验证 ===")
	for _, user := range users {
		achievements, err := repository.GetAchievementsByUserID(user.ID)
		if err != nil {
			log.Printf("获取用户 %s 的成就失败: %v", user.Name, err)
			continue
		}

		fmt.Printf("用户 %s (ID: %d): %d 个成就\n", user.Name, user.ID, len(achievements))

		// 检查成就名称是否有效
		validNames := []string{
			"首次完成", "7天连卡", "任务大师", "目标达成", "学习之星", "坚持不懈",
			"效率达人", "专注大师", "早起鸟", "夜猫子", "完美主义", "全能选手",
			"学习狂人", "社交达人", "时间管理者", "成就收集者",
		}
		validSet := make(map[string]bool)
		for _, name := range validNames {
			validSet[name] = true
		}

		for _, achievement := range achievements {
			if !validSet[achievement.Name] {
				fmt.Printf("  ❌ 无效成就: %s\n", achievement.Name)
			}
		}
	}
}