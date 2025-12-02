// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginWithOTPLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginWithOTPLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginWithOTPLogic {
	return &LoginWithOTPLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginWithOTPLogic) LoginWithOTP(req *types.LoginWithOTPReq) (resp *types.LoginResp, err error) {
	// todo: add your logic here and delete this line

	return
}
