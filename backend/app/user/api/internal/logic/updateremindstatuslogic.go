// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRemindStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRemindStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRemindStatusLogic {
	return &UpdateRemindStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRemindStatusLogic) UpdateRemindStatus() (resp *types.VerifyEmailResp, err error) {
	// todo: add your logic here and delete this line

	return
}
