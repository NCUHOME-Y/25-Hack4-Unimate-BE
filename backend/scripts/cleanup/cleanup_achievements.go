package main

import (
	"fmt"
	"log"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("开始清理重复的成就数据...")

	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("警告: 加载 .env 文件失败: %v", err)
	}

	// 连接数据库
	repository.DBconnect()

	// 定义正确的16个成就名称映射
	// correctAchievements := map[string]string{
	// 	// 旧名称 -> 新名称 的映射
	// 	"学习达人":     "学习之星",
	// 	"专注家":      "专注大师",
	// 	"新手启程":    "首次完成",
	// 	"任务收藏家":   "任务大师",
	// }

	// 标准的16个成就名称
	validAchievementNames := []string{
		"首次完成", "7天连卡", "任务大师", "目标达成", "学习之星", "坚持不懈",
		"效率达人", "专注大师", "早起鸟", "夜猫子", "完美主义", "全能选手",
		"学习狂人", "社交达人", "时间管理者", "成就收集者",
	}

	// 创建成就名称集合以便快速查找
	validNamesSet := make(map[string]bool)
	for _, name := range validAchievementNames {
		validNamesSet[name] = true
	}

	// 获取所有用户
	users, err := repository.GetAllUser()
	if err != nil {
		log.Fatalf("获取用户列表失败: %v", err)
	}

	totalDeleted := 0

	for _, user := range users {
		fmt.Printf("正在处理用户: %s (ID: %d)\n", user.Name, user.ID)

		// 获取该用户的所有成就
		achievements, err := repository.GetAchievementsByUserID(user.ID)
		if err != nil {
			log.Printf("获取用户 %s 的成就失败: %v", user.Name, err)
			continue
		}

		// 使用map按成就名称分组，保留ID最小的记录
		achievementMap := make(map[string][]model.Achievement)
		for _, achievement := range achievements {
			achievementMap[achievement.Name] = append(achievementMap[achievement.Name], achievement)
		}

		fmt.Printf("  用户 %s 的成就分布:\n", user.Name)
		for name, achievementList := range achievementMap {
			fmt.Printf("    '%s': %d 条记录\n", name, len(achievementList))
		}

		// 第一步：清理无效的成就名称
		for _, achievement := range achievements {
			if !validNamesSet[achievement.Name] {
				fmt.Printf("  删除无效成就: '%s' (ID: %d)\n", achievement.Name, achievement.ID)
				err := repository.DB.Delete(&model.Achievement{}, achievement.ID).Error
				if err != nil {
					log.Printf("删除无效成就记录 ID=%d 失败: %v", achievement.ID, err)
				} else {
					totalDeleted++
				}
			}
		}

		// 第二步：重新获取清理后的成就，并处理重复项
		achievements, _ = repository.GetAchievementsByUserID(user.ID)

		// 重新构建成就映射
		achievementMap = make(map[string][]model.Achievement)
		for _, achievement := range achievements {
			achievementMap[achievement.Name] = append(achievementMap[achievement.Name], achievement)
		}
		var toDelete []uint
		for name, achievementList := range achievementMap {
			if len(achievementList) > 1 {
				fmt.Printf("  发现重复成就 '%s': %d 条记录\n", name, len(achievementList))

				// 保留ID最小的记录，删除其他的
				minID := achievementList[0].ID
				for _, achievement := range achievementList[1:] {
					if achievement.ID < minID {
						minID = achievement.ID
					}
				}

				// 收集需要删除的ID
				for _, achievement := range achievementList {
					if achievement.ID != minID {
						toDelete = append(toDelete, achievement.ID)
					}
				}
			}
		}

		// 删除重复记录 - 使用正确的删除方式
		for _, id := range toDelete {
			// 直接使用GORM删除
			err := repository.DB.Delete(&model.Achievement{}, id).Error
			if err != nil {
				log.Printf("删除成就记录 ID=%d 失败: %v", id, err)
			} else {
				fmt.Printf("  已删除重复成就记录 ID=%d\n", id)
				totalDeleted++
			}
		}
	}

	fmt.Printf("\n清理完成！共删除了 %d 条重复的成就记录\n", totalDeleted)

	// 验证清理结果
	fmt.Println("\n验证清理结果...")
	for _, user := range users {
		achievements, err := repository.GetAchievementsByUserID(user.ID)
		if err != nil {
			log.Printf("验证用户 %s 的成就失败: %v", user.Name, err)
			continue
		}

		// 检查是否还有重复
		nameCount := make(map[string]int)
		for _, achievement := range achievements {
			nameCount[achievement.Name]++
		}

		hasDuplicate := false
		for name, count := range nameCount {
			if count > 1 {
				fmt.Printf("  ⚠️ 用户 %s 仍有重复成就 '%s': %d 条\n", user.Name, name, count)
				hasDuplicate = true
			}
		}

		if !hasDuplicate {
			fmt.Printf("  ✅ 用户 %s 的成就数据正常，共 %d 个成就\n", user.Name, len(achievements))
		}
	}
}