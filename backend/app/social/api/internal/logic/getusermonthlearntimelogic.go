// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/social/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/social/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserMonthLearnTimeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserMonthLearnTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserMonthLearnTimeLogic {
	return &GetUserMonthLearnTimeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserMonthLearnTimeLogic) GetUserMonthLearnTime() (resp *types.RankingResp, err error) {
	// todo: add your logic here and delete this line

	return
}
