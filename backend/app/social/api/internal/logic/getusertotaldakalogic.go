// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/social/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/social/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserTotalDakaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserTotalDakaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTotalDakaLogic {
	return &GetUserTotalDakaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserTotalDakaLogic) GetUserTotalDaka() (resp *types.RankingResp, err error) {
	// todo: add your logic here and delete this line

	return
}
