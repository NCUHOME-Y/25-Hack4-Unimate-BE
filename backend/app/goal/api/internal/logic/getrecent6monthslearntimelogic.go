// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRecent6MonthsLearnTimeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRecent6MonthsLearnTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRecent6MonthsLearnTimeLogic {
	return &GetRecent6MonthsLearnTimeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRecent6MonthsLearnTimeLogic) GetRecent6MonthsLearnTime() (resp *types.GetLearnTimeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
