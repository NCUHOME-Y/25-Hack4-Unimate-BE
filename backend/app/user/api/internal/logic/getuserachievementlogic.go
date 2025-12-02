// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserAchievementLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserAchievementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserAchievementLogic {
	return &GetUserAchievementLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserAchievementLogic) GetUserAchievement() (resp *types.GetUserAchievementResp, err error) {
	// todo: add your logic here and delete this line

	return
}
