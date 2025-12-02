// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddLearnTimeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddLearnTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLearnTimeLogic {
	return &AddLearnTimeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddLearnTimeLogic) AddLearnTime(req *types.AddLearnTimeReq) (resp *types.AddFlagResp, err error) {
	// todo: add your logic here and delete this line

	return
}
