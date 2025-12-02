// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLearnTimeMonthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLearnTimeMonthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLearnTimeMonthLogic {
	return &GetLearnTimeMonthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLearnTimeMonthLogic) GetLearnTimeMonth() (resp *types.GetLearnTimeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
