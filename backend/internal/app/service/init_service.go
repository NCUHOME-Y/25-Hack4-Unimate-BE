package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

var cronScheduler *cron.Cron
var userReminderJobs = make(map[uint]cron.EntryID)
var reminderMutex sync.Mutex

func Init() {
	cronScheduler = cron.New(cron.WithSeconds())

	utils.LogInfo(" 开始初始化定时任务", nil)

	users, err := repository.GetAllUser()
	if err != nil {
		utils.LogError("获取用户列表失败", logrus.Fields{"error": err.Error()})
		return
	}

	// 🐛 修复：注释掉有问题的凌晨4点清零任务
	// 这个任务会把前一天的学习时长也清零，导致数据丢失
	// TODO: 如果需要限制通宵学习，应该只清空前端 localStorage，不应修改已保存的数据库记录
	/*
		_, err = cronScheduler.AddFunc("0 0 4 * * *", func() {
			err := repository.InvalidateAllTodayLearnTime()
			if err != nil {
				utils.LogError("凌晨4点自动停止学习计时失败", logrus.Fields{"error": err.Error()})
			} else {
				utils.LogInfo("凌晨4点自动停止学习计时成功", nil)
			}
		})
		if err != nil {
			utils.LogError("添加凌晨4点自动停止学习计时任务失败", logrus.Fields{"error": err.Error()})
		}
	*/

	for _, u := range users {
		user := u
		// 🐛 修复闭包陷阱：捕获用户ID值，避免引用循环变量
		userID := user.ID

		// 每日任务
		_, err := cronScheduler.AddFunc("@daily", func() {
			// 注意：GetUserByID 默认不 Preload DaKaNumber；这里直接按 user_id 查询，确保每日任务完整生效。
			var dakaRecords []model.Daka_number
			if err := repository.DB.Where("user_id = ?", userID).Find(&dakaRecords).Error; err != nil {
				utils.LogError("每日任务获取打卡记录失败", logrus.Fields{"user_id": userID, "error": err.Error()})
				return
			}

			var flags []model.Flag
			if err := repository.DB.Where("user_id = ?", userID).Find(&flags).Error; err != nil {
				utils.LogError("每日任务获取Flag失败", logrus.Fields{"user_id": userID, "error": err.Error()})
				return
			}

			InitDakaNumberRecord(dakaRecords, userID)
			InitDaliyLearnTimeRecord(userID)
			InitDaliyFlag(flags)
			utils.LogInfo("执行每日初始化任务", logrus.Fields{"user_id": userID})
		})
		if err != nil {
			utils.LogError("添加每日任务失败", logrus.Fields{"user_id": userID, "error": err.Error()})
		}

		// 每月任务
		_, err = cronScheduler.AddFunc("@monthly", func() {
			InitMonthlyDakaRecord(userID)
			// 重置月学习时长（原子更新，避免 Save 覆盖其它字段）
			if err := repository.DB.Model(&model.User{}).Where("id = ?", userID).Update("month_learntime", 0).Error; err != nil {
				utils.LogError("重置用户每月学习时长失败", logrus.Fields{"user_id": userID, "error": err.Error()})
				return
			}
			utils.LogInfo("执行每月初始化任务", logrus.Fields{"user_id": userID})
		})
		if err != nil {
			utils.LogError("添加每月任务失败", logrus.Fields{"user_id": userID, "error": err.Error()})
		}

		// 学习提醒任务（用户级学习提醒）
		if user.IsStudyRemind {
			hour := user.StudyRemindHour
			min := user.StudyRemindMin
			// 修复：使用正确的 cron 格式（秒 分 时 日 月 周）
			cronStr := fmt.Sprintf("0 %d %d * * *", min, hour)
			// 🐛 修复闭包陷阱：捕获当前用户ID/时间副本，避免闭包引用外部变量
			userID := user.ID
			reminderHour := hour
			reminderMin := min
			entryID, err := cronScheduler.AddFunc(cronStr, func() {
				// 发送时只需要邮箱/姓名，用轻量查询即可
				currentUser, err := repository.GetUserBasicByID(userID)
				if err != nil {
					utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID, "error": err.Error()})
					return
				}

				utils.LogInfo("发送定时提醒邮件", logrus.Fields{
					"user_id": userID,
					"email":   currentUser.Email,
					"time":    fmt.Sprintf("%02d:%02d", reminderHour, reminderMin),
				})

				err = SendStudyReminderEmail(currentUser.Email, currentUser.Name, reminderHour, reminderMin)
				if err != nil {
					utils.LogError("发送提醒邮件失败", logrus.Fields{
						"user_id": userID,
						"error":   err.Error(),
					})
				} else {
					utils.LogInfo("提醒邮件发送成功", logrus.Fields{"user_id": userID})
				}
			})

			if err != nil {
				utils.LogError("添加提醒任务失败", logrus.Fields{
					"user_id":  user.ID,
					"cron_str": cronStr,
					"error":    err.Error(),
				})
			} else {
				// 保存任务ID，方便后续更新
				reminderMutex.Lock()
				userReminderJobs[user.ID] = entryID
				reminderMutex.Unlock()

				utils.LogInfo("✅ 添加提醒任务成功", logrus.Fields{
					"user_id": user.ID,
					"time":    fmt.Sprintf("%02d:%02d", hour, min),
				})
			}
		}
	}

	// 验证码清理任务 - 每5分钟执行一次
	_, err = cronScheduler.AddFunc("0 */5 * * * *", func() {
		err := repository.DeleteExpiredEmailCodes()
		if err != nil {
			utils.LogError("清理过期验证码失败", logrus.Fields{"error": err.Error()})
		} else {
			utils.LogInfo("✅ 清理过期验证码成功", nil)
		}
	})
	if err != nil {
		utils.LogError("添加验证码清理任务失败", logrus.Fields{"error": err.Error()})
	} else {
		utils.LogInfo("✅ 验证码清理任务已启动(每5分钟执行)", nil)
	}

	// Flag提醒任务 - 每分钟检查一次
	_, err = cronScheduler.AddFunc("0 * * * * *", func() {
		sendFlagReminders()
	})
	if err != nil {
		utils.LogError("添加Flag提醒任务失败", logrus.Fields{"error": err.Error()})
	} else {
		utils.LogInfo("✅ Flag提醒任务已启动(每分钟检查)", nil)
	}

	cronScheduler.Start()
	utils.LogInfo("初始化定时任务成功", logrus.Fields{
		"total_users": len(users),
		"total_jobs":  len(cronScheduler.Entries()),
	})
}

// 为新用户添加定时任务
func AddUserCronJob(user model.User) {
	if cronScheduler == nil {
		utils.LogError("定时任务调度器未初始化", nil)
		return
	}

	// 🐛 修复闭包陷阱：捕获用户ID值
	userID := user.ID

	// 每日任务
	cronScheduler.AddFunc("@daily", func() {
		var dakaRecords []model.Daka_number
		if err := repository.DB.Where("user_id = ?", userID).Find(&dakaRecords).Error; err != nil {
			return
		}

		var flags []model.Flag
		if err := repository.DB.Where("user_id = ?", userID).Find(&flags).Error; err != nil {
			return
		}

		InitDakaNumberRecord(dakaRecords, userID)
		InitDaliyLearnTimeRecord(userID)
		InitDaliyFlag(flags)
	})

	// 每月任务
	cronScheduler.AddFunc("@monthly", func() {
		InitMonthlyDakaRecord(userID)
		_ = repository.DB.Model(&model.User{}).Where("id = ?", userID).Update("month_learntime", 0).Error
	})

	// 学习提醒任务（用户级学习提醒）
	if user.IsStudyRemind {
		hour := user.StudyRemindHour
		min := user.StudyRemindMin

		cronStr := fmt.Sprintf("0 %d %d * * *", min, hour)
		// 🐛 修复闭包陷阱：捕获用户ID和时间副本，发送时用轻量查询获取最新邮箱/姓名
		userID := user.ID
		reminderHour := hour
		reminderMin := min

		// 避免重复注册：若已存在旧任务，先移除
		reminderMutex.Lock()
		if oldJobID, exists := userReminderJobs[userID]; exists {
			cronScheduler.Remove(oldJobID)
			delete(userReminderJobs, userID)
		}
		reminderMutex.Unlock()

		entryID, err := cronScheduler.AddFunc(cronStr, func() {
			currentUser, err := repository.GetUserBasicByID(userID)
			if err != nil {
				return
			}
			SendStudyReminderEmail(currentUser.Email, currentUser.Name, reminderHour, reminderMin)
		})
		if err != nil {
			utils.LogError("为新用户添加提醒任务失败", logrus.Fields{"user_id": userID, "cron_str": cronStr, "error": err.Error()})
		} else {
			reminderMutex.Lock()
			userReminderJobs[userID] = entryID
			reminderMutex.Unlock()
		}
		utils.LogInfo("为新用户添加提醒任务", logrus.Fields{
			"user_id": user.ID,
			"time":    fmt.Sprintf("%02d:%02d", hour, min),
		})
	}

	utils.LogInfo("✅ 为新用户添加定时任务", logrus.Fields{"user_id": user.ID})
}

// 更新用户的学习提醒任务
func UpdateUserReminderJob(userID uint, hour, min int, isStudyRemind bool) {
	if cronScheduler == nil {
		utils.LogError("定时任务调度器未初始化", nil)
		return
	}

	reminderMutex.Lock()
	defer reminderMutex.Unlock()

	// 移除旧的提醒任务
	if oldJobID, exists := userReminderJobs[userID]; exists {
		cronScheduler.Remove(oldJobID)
		delete(userReminderJobs, userID)
		utils.LogInfo("🗑️ 移除旧的提醒任务", logrus.Fields{"user_id": userID})
	}

	// 如果开启学习提醒，添加新的任务
	if isStudyRemind {
		cronStr := fmt.Sprintf("0 %d %d * * *", min, hour)
		// 🐛 修复闭包陷阱：捕获 userID 和时间副本，发送时重新获取邮箱/姓名
		reminderHour := hour
		reminderMin := min
		entryID, err := cronScheduler.AddFunc(cronStr, func() {
			user, err := repository.GetUserBasicByID(userID)
			if err != nil {
				utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID, "error": err.Error()})
				return
			}

			utils.LogInfo("⏰ 发送定时提醒邮件", logrus.Fields{
				"user_id": userID,
				"email":   user.Email,
				"time":    fmt.Sprintf("%02d:%02d", reminderHour, reminderMin),
			})

			SendStudyReminderEmail(user.Email, user.Name, reminderHour, reminderMin)
		})

		if err != nil {
			utils.LogError("添加新提醒任务失败", logrus.Fields{
				"user_id":  userID,
				"cron_str": cronStr,
				"error":    err.Error(),
			})
		} else {
			userReminderJobs[userID] = entryID
			utils.LogInfo("✅ 更新提醒任务成功", logrus.Fields{
				"user_id": userID,
				"time":    fmt.Sprintf("%02d:%02d", hour, min),
			})
		}
	}
}

// 初始化每天学习时间记录
func InitDaliyLearnTimeRecord(id uint) {
	Time, _ := repository.GetTodayLearnTime(id)

	// 🐛 修复：用原子更新累加月学习时长，避免 Save 覆盖其他字段
	if Time.Duration > 0 {
		err := repository.DB.Exec("UPDATE users SET month_learntime = month_learntime + ? WHERE id = ?", Time.Duration, id).Error
		if err != nil {
			utils.LogError("更新用户月学习时长失败", logrus.Fields{"user_id": id, "error": err.Error()})
		}
	}

	err := repository.AddNewLearnTimeToDB(id)
	if err != nil {
		utils.LogError("添加新的学习时间记录失败", logrus.Fields{"user_id": id})
		return
	}
	utils.LogInfo("添加新的学习时间记录成功", logrus.Fields{"user_id": id})
}

// 初始化每天flag（重置完成状态和计数）
func InitDaliyFlag(flags []model.Flag) {
	for _, flag := range flags {
		// 重置完成状态
		err := repository.UpdateFlagHadDone(flag.ID, false)
		if err != nil {
			utils.LogError("初始化每日flag完成状态失败", logrus.Fields{"flag_id": flag.ID})
			continue
		}
		// 重置每日完成次数为0
		err = repository.UpdateFlagDoneNumber(flag.ID, 0)
		if err != nil {
			utils.LogError("初始化每日flag计数失败", logrus.Fields{"flag_id": flag.ID})
			continue
		}
	}
	utils.LogInfo("每日flag初始化完成", logrus.Fields{"flag_count": len(flags)})
}

// 初始化打卡记录
func InitDakaNumberRecord(daka []model.Daka_number, id uint) {
	// 兼容旧调度入口：当前打卡统计在用户实际打卡时已即时落库，
	_ = daka
	_ = id
}

// 每月建立打卡记录
func InitMonthlyDakaRecord(id uint) {
	err := repository.AddNewDakaNumberToDB(id)
	if err != nil {
		utils.LogError("添加新的打卡记录失败", logrus.Fields{"user_id": id})
		return
	}
}

// 发送Flag提醒邮件
func sendFlagReminders() {
	// 获取所有启用了提醒的Flag
	var flags []model.Flag
	err := repository.DB.Where("enable_notification = ?", true).Find(&flags).Error
	if err != nil {
		utils.LogError("获取启用提醒的Flag失败", logrus.Fields{"error": err.Error()})
		return
	}

	if len(flags) == 0 {
		return
	}

	// 获取当前时间（小时和分钟）
	now := time.Now()
	currentHour := now.Hour()
	currentMinute := now.Minute()

	for _, flag := range flags {
		// 跳过已完成的flag
		if flag.Completed {
			continue
		}

		// 解析提醒时间
		var reminderHour, reminderMinute int
		if flag.ReminderTime == "" {
			flag.ReminderTime = "12:00"
		}
		_, err := fmt.Sscanf(flag.ReminderTime, "%d:%d", &reminderHour, &reminderMinute)
		if err != nil {
			utils.LogError("解析Flag提醒时间失败", logrus.Fields{
				"flag_id":       flag.ID,
				"reminder_time": flag.ReminderTime,
				"error":         err.Error(),
			})
			continue
		}

		// 检查是否到了提醒时间（精确到分钟）
		if currentHour == reminderHour && currentMinute == reminderMinute {
			// 获取用户信息
			user, err := repository.GetUserByID(flag.UserID)
			if err != nil {
				utils.LogError("获取用户信息失败", logrus.Fields{
					"user_id": flag.UserID,
					"error":   err.Error(),
				})
				continue
			}

			// 检查用户是否开启了 Flag 级别的提醒（用户级总开关）
			if !user.IsFlagRemind {
				// 跳过发送flag提醒
				continue
			}

			// 发送Flag提醒邮件
			err = SendFlagReminderEmail(user.Email, user.Name, flag.Title, flag.Detail, flag.ReminderTime, flag.Priority)
			if err == nil {
				utils.LogInfo("✅ Flag提醒邮件发送成功", logrus.Fields{
					"user_id":   flag.UserID,
					"flag_id":   flag.ID,
					"flag_name": flag.Title,
					"email":     user.Email,
					"time":      flag.ReminderTime,
				})
			}
		}
	}
}
