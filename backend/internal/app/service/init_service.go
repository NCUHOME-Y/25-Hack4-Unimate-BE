package service

import (
	"fmt"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func Init() {
	cron := cron.New()
	users, _ := repository.GetAllUser()
	for _, u := range users {
		user := u
		cron.AddFunc("@daily", func() {
			InitDakaNumberRecord(user.DaKaNumber, user.ID)
			InitDaliyLearnTimeRecord(user.ID)
			InitDaliyFlag(user.Flags)
		})
		cron.AddFunc("@monthly", func() {
			InitMonthlyDakaRecord(user.ID)
		})
		if !user.IsRemind {
			continue
		}
		str := fmt.Sprintf("0 %d %d * * *", user.RemindMin, user.RemindHour)
		cron.AddFunc(str, func() {
			utils.SentEmail(user.Email, "知序：提醒您要好好自律哦", "灵魂的欲望是你命运的先知")
		})
	}
	cron.Start()
	utils.LogInfo("初始化定时任务成功", nil)
}

// 初始化每天学习时间记录
func InitDaliyLearnTimeRecord(id uint) {
	err := repository.AddNewLearnTimeToDB(id)
	if err != nil {
		utils.LogError("添加新的学习时间记录失败", logrus.Fields{"user_id": id})
		return
	}
	utils.LogInfo("添加新的学习时间记录成功", logrus.Fields{"user_id": id})
}

// 初始化每天学习时间记录
func InitDaliyFlag(flags []model.Flag) {
	for _, flag := range flags {
		err := repository.UpdateFlagHadDone(flag.ID, false)
		if err != nil {
			utils.LogError("初始化每日签到状态失败", logrus.Fields{"flag_id": flag.ID})
			return
		}
	}
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
}

// 每月建立打卡记录
func InitMonthlyDakaRecord(id uint) {
	err := repository.AddNewDakaNumberToDB(id)
	if err != nil {
		utils.LogError("添加新的打卡记录失败", logrus.Fields{"user_id": id})
		return
	}
}
