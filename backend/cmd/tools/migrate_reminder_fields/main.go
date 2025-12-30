package main

import (
	"log"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	"github.com/joho/godotenv"
)

// 一次性数据迁移工具：将旧的提醒字段数据复制到新字段
// 运行后可以安全删除旧字段的兼容逻辑
func main() {
	// 加载环境变量
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Println("警告: 未找到 .env 文件，使用系统环境变量")
	}

	// 连接数据库
	repository.DBconnect()
	if repository.DB == nil {
		log.Fatal("❌ 数据库连接失败")
	}
	log.Println("✅ 数据库连接成功")

	// 执行迁移 SQL
	// 只迁移那些"新字段还是默认值，但旧字段有值"的老用户
	sql := `
		UPDATE users 
		SET 
			is_study_remind = is_remind,
			study_remind_hour = remind_hour,
			study_remind_min = remind_min
		WHERE 
			-- 条件1：新字段还是默认值（说明是老用户）
			(is_study_remind = false AND study_remind_hour = 12 AND study_remind_min = 0)
			-- 条件2：旧字段有非默认值（说明用户设置过提醒）
			AND (is_remind = true OR remind_hour != 12 OR remind_min != 0)
	`

	result := repository.DB.Exec(sql)
	if result.Error != nil {
		log.Fatalf("❌ 迁移失败: %v", result.Error)
	}

	log.Printf("✅ 迁移完成！共更新 %d 个老用户的提醒字段", result.RowsAffected)
	log.Println("现在可以安全移除代码里的旧字段兼容逻辑了")
}
