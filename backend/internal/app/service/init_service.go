package service

import (
	"fmt"
	"sync"

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

	// 每天凌晨4点自动停止所有学习计时且本次不计入学习时长
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

	for _, u := range users {
		user := u

		// 每日任务
		_, err := cronScheduler.AddFunc("@daily", func() {
			InitDakaNumberRecord(user.DaKaNumber, user.ID)
			InitDaliyLearnTimeRecord(user.ID)
			InitDaliyFlag(user.Flags)
			utils.LogInfo("执行每日初始化任务", logrus.Fields{"user_id": user.ID})
		})
		if err != nil {
			utils.LogError("添加每日任务失败", logrus.Fields{"user_id": user.ID, "error": err.Error()})
		}

		// 每月任务
		_, err = cronScheduler.AddFunc("@monthly", func() {
			InitMonthlyDakaRecord(user.ID)
			user.MonthLearntime = 0
			err := repository.SaveUserToDB(user)
			if err != nil {
				utils.LogError("重置用户每月学习时长失败", logrus.Fields{"user_id": user.ID, "error": err.Error()})
			}
			utils.LogInfo("执行每月初始化任务", logrus.Fields{"user_id": user.ID})
		})
		if err != nil {
			utils.LogError("添加每月任务失败", logrus.Fields{"user_id": user.ID, "error": err.Error()})
		}

		// 提醒任务
		if user.IsRemind {
			// 修复：使用正确的 cron 格式（秒 分 时 日 月 周）
			cronStr := fmt.Sprintf("0 %d %d * * *", user.RemindMin, user.RemindHour)
			entryID, err := cronScheduler.AddFunc(cronStr, func() {
				utils.LogInfo("发送定时提醒邮件", logrus.Fields{
					"user_id": user.ID,
					"email":   user.Email,
					"time":    fmt.Sprintf("%02d:%02d", user.RemindHour, user.RemindMin),
				})

				err := utils.SentEmail(user.Email, "知序提醒您：要好好自律哦", "温馨提示:灵魂的欲望是你命运的先知")
				if err != nil {
					utils.LogError("发送提醒邮件失败", logrus.Fields{
						"user_id": user.ID,
						"error":   err.Error(),
					})
				} else {
					utils.LogInfo("提醒邮件发送成功", logrus.Fields{"user_id": user.ID})
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
					"time":    fmt.Sprintf("%02d:%02d", user.RemindHour, user.RemindMin),
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

	// 每日任务
	cronScheduler.AddFunc("@daily", func() {
		InitDakaNumberRecord(user.DaKaNumber, user.ID)
		InitDaliyLearnTimeRecord(user.ID)
		InitDaliyFlag(user.Flags)
	})

	// 每月任务
	cronScheduler.AddFunc("@monthly", func() {
		InitMonthlyDakaRecord(user.ID)
	})

	// 提醒任务
	if user.IsRemind {
		cronStr := fmt.Sprintf("0 %d %d * * *", user.RemindMin, user.RemindHour)
		cronScheduler.AddFunc(cronStr, func() {
			utils.SentEmail(user.Email, "知序提醒您：要好好自律哦", "灵魂的欲望是你命运的先知")
		})
		utils.LogInfo("为新用户添加提醒任务", logrus.Fields{
			"user_id": user.ID,
			"time":    fmt.Sprintf("%02d:%02d", user.RemindHour, user.RemindMin),
		})
	}

	utils.LogInfo("✅ 为新用户添加定时任务", logrus.Fields{"user_id": user.ID})
}

// 更新用户的提醒任务
func UpdateUserReminderJob(userID uint, hour, min int, isRemind bool) {
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

	// 如果开启提醒，添加新的任务
	if isRemind {
		// 获取用户信息
		user, err := repository.GetUserByID(userID)
		if err != nil {
			utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID, "error": err.Error()})
			return
		}

		cronStr := fmt.Sprintf("0 %d %d * * *", min, hour)
		entryID, err := cronScheduler.AddFunc(cronStr, func() {
			utils.LogInfo("⏰ 发送定时提醒邮件", logrus.Fields{
				"user_id": userID,
				"email":   user.Email,
				"time":    fmt.Sprintf("%02d:%02d", hour, min),
			})

			err := utils.SentEmail(user.Email, "知序提醒您：要好好自律哦", "灵魂的欲望是你命运的先知")
			if err != nil {
				utils.LogError("发送提醒邮件失败", logrus.Fields{
					"user_id": userID,
					"error":   err.Error(),
				})
			} else {
				utils.LogInfo("✅ 提醒邮件发送成功", logrus.Fields{"user_id": userID})
			}
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
	user, _ := repository.GetUserByID(id)
	Time, _ := repository.GetTodayLearnTime(id)
	user.MonthLearntime = user.MonthLearntime + Time.Duration
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
	for _, daka_record := range daka {
		err := repository.UpdateDakaHadDone(id)
		if err != nil {
			utils.LogError("初始化每日打卡状态失败", logrus.Fields{"daka_id": daka_record.ID})
			return
		}
	}
	user, _ := repository.GetUserByID(id)
	daka1, _ := repository.GetRecentDakaNumber(id)
	if daka1.HadDone {
		daka1.MonthDaka = daka1.MonthDaka + 1
		user.Daka = user.Daka + 1
	}
	err := repository.SaveUserToDB(user)
	if err != nil {
		utils.LogError("保存用户数据失败", logrus.Fields{"user_id": id})
		return
	}
}

// 每月建立打卡记录
func InitMonthlyDakaRecord(id uint) {
	err := repository.AddNewDakaNumberToDB(id)
	if err != nil {
		utils.LogError("添加新的打卡记录失败", logrus.Fields{"user_id": id})
		return
	}
}
