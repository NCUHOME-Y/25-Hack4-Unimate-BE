package service

import (
	"fmt"

	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/sirupsen/logrus"
)

const platformURL = "http://139.199.157.76"

// SendStudyReminderEmail 发送学习提醒邮件
func SendStudyReminderEmail(email, userName string, hour, min int) error {
	subject := "【知序】您的学习提醒"
	body := fmt.Sprintf(`尊敬的%s，您好！

这是您设定的每日学习提醒。

提醒时间：%02d:%02d

请记得保持良好的学习习惯，持续追求进步。自律是成功的基石，每一天的坚持都将成为您通往目标的阶梯。

请登录知序平台（%s）查看详情。

——知序平台`, userName, hour, min, platformURL)

	if err := utils.SentEmail(email, subject, body); err != nil {
		utils.LogError("学习提醒邮件发送失败", logrus.Fields{
			"email": email,
			"error": err.Error(),
		})
		return err
	}
	utils.LogInfo("学习提醒邮件发送成功", logrus.Fields{"email": email})
	return nil
}

// SendVerificationCodeEmail 发送验证码邮件
func SendVerificationCodeEmail(email, code string) error {
	subject := "【知序】账号注册验证码"
	body := fmt.Sprintf(`您好！

您正在注册知序账号，验证码为：

%s

验证码有效期为5分钟，请及时使用。

如非本人操作，请忽略此邮件。

——知序平台`, code)

	if err := utils.SentEmail(email, subject, body); err != nil {
		utils.LogError("验证码邮件发送失败", logrus.Fields{
			"email": email,
			"error": err.Error(),
		})
		return err
	}
	utils.LogInfo("验证码邮件发送成功", logrus.Fields{"email": email})
	return nil
}

// SendIdentityVerificationEmail 发送身份验证码邮件
func SendIdentityVerificationEmail(email, code string) error {
	subject := "【知序】身份验证码"
	body := fmt.Sprintf(`您好！

您正在进行身份验证，验证码为：

%s

验证码有效期为5分钟，请及时使用。

如非本人操作，请忽略此邮件。

——知序平台`, code)

	if err := utils.SentEmail(email, subject, body); err != nil {
		utils.LogError("身份验证码邮件发送失败", logrus.Fields{
			"email": email,
			"error": err.Error(),
		})
		return err
	}
	utils.LogInfo("身份验证码邮件发送成功", logrus.Fields{"email": email})
	return nil
}

// SendPostCommentNotification 发送帖子评论通知邮件
func SendPostCommentNotification(receiverEmail, receiverName, postTitle, commenterName, commentContent string) {
	subject := "【知序】您收到了新的消息"
	body := fmt.Sprintf(`尊敬的%s，您好！

您的帖子"%s"收到了新的评论：

评论者：%s
评论内容：%s

请登录知序平台（%s）查看详情。

——知序平台`, receiverName, postTitle, commenterName, commentContent, platformURL)

	go func() {
		if err := utils.SentEmail(receiverEmail, subject, body); err != nil {
			utils.LogError("帖子评论通知邮件发送失败", logrus.Fields{
				"email": receiverEmail,
				"error": err.Error(),
			})
		} else {
			utils.LogInfo("帖子评论通知邮件发送成功", logrus.Fields{"email": receiverEmail})
		}
	}()
}

// SendFlagCommentNotification 发送Flag评论通知邮件
func SendFlagCommentNotification(receiverEmail, receiverName, flagTitle, commenterName, commentContent string) {
	subject := "【知序】您收到了新的消息"
	body := fmt.Sprintf(`尊敬的%s，您好！

您的目标"%s"收到了新的评论：

评论者：%s
评论内容：%s

请登录知序平台（%s）查看详情。

——知序平台`, receiverName, flagTitle, commenterName, commentContent, platformURL)

	go func() {
		if err := utils.SentEmail(receiverEmail, subject, body); err != nil {
			utils.LogError("目标评论通知邮件发送失败", logrus.Fields{
				"email": receiverEmail,
				"error": err.Error(),
			})
		} else {
			utils.LogInfo("目标评论通知邮件发送成功", logrus.Fields{"email": receiverEmail})
		}
	}()
}

// SendFlagReminderEmail 发送Flag提醒邮件
func SendFlagReminderEmail(email, userName, flagTitle, flagDetail, reminderTime string, priority int) error {
	subject := "知序 - Flag目标提醒通知"
	body := fmt.Sprintf(`尊敬的 %s 用户：您好！

这是来自知序平台的Flag目标提醒通知。

【Flag详情】
标题：%s
详细说明：%s
提醒时间：%s
优先级：%d

请及时完成您设定的目标任务，保持良好的学习和生活习惯。持续的自律和坚持将帮助您实现更好的自己。

请登录知序平台（%s）查看详情。

——知序平台
`, userName, flagTitle, flagDetail, reminderTime, priority, platformURL)

	if err := utils.SentEmail(email, subject, body); err != nil {
		utils.LogError("Flag提醒邮件发送失败", logrus.Fields{
			"email": email,
			"error": err.Error(),
		})
		return err
	}
	utils.LogInfo("Flag提醒邮件发送成功", logrus.Fields{"email": email})
	return nil
}

// SendPrivateMessageNotification 发送私信通知邮件
func SendPrivateMessageNotification(receiverEmail, receiverName, senderName, messageContent string) {
	subject := "【知序】您收到了新的私信"
	body := fmt.Sprintf(`尊敬的%s，您好！

您收到了来自 %s 的新私信：

%s

请登录知序平台（%s）查看详情。

——知序平台`, receiverName, senderName, messageContent, platformURL)

	go func() {
		if err := utils.SentEmail(receiverEmail, subject, body); err != nil {
			utils.LogError("私信通知邮件发送失败", logrus.Fields{
				"email": receiverEmail,
				"error": err.Error(),
			})
		} else {
			utils.LogInfo("私信通知邮件发送成功", logrus.Fields{"email": receiverEmail})
		}
	}()
}
