// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/types"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	user, err := repository.GetUserByEmail(req.Email)
	// 检查用户是否存在
	if err != nil || user.ID == 0 {
		return nil, errors.New("用户名或密码错误")
	}
	// 检查密码是否正确
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}
	token, err := utils.GenerateToken(user.ID, user.Name, user.Email)
	if err != nil {
		return nil, errors.New("生成 Token 失败")
	}

	return &types.LoginResp{
		Message:        "登录成功!",
		UserId:         user.ID,
		Name:           user.Name,
		Email:          user.Email,
		HeadShow:       user.HeadShow,
		Daka:           user.Daka,
		FlagNumber:     user.FlagNumber,
		Count:          user.Count,
		MonthLearnTime: user.MonthLearntime,
		Token:          token,
	}, nil
}
