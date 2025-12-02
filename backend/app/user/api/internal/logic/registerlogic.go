// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/types"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// 检查邮箱是否已注册
	user_exist, _ := repository.GetUserByEmail(req.Email)
	if user_exist.ID != 0 {
		return nil, errors.New("该邮箱已被注册,请更换邮箱")
	}
	// 检查用户名是否已存在
	name_exist, _ := repository.GetUserByName(req.Name)
	if name_exist.ID != 0 {
		return nil, errors.New("该用户名已被使用,请更换用户名")
	}
	password, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("注册失败,请重新再试")
	}

	// 验证码机制
	code := utils.GenerateCode()
	err = utils.SentEmail(req.Email, "知序验证码", "您的验证码是："+code+"\n该验证码5分钟内有效,请尽快使用。")
	if err != nil {
		return nil, errors.New("验证码发送失败,请重新再试")
	}

	repository.SaveEmailCodeToDB(code, req.Email)

	user := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: password,
	}

	if err := repository.AddUserToDB(user); err != nil {
		return nil, errors.New("数据库添加用户失败")
	}

	return &types.RegisterResp{
		Message: "注册成功!",
	}, nil
}
