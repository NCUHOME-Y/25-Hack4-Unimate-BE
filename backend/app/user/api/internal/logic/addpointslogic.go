// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddPointsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddPointsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPointsLogic {
	return &AddPointsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddPointsLogic) AddPoints(req *types.AddPointsReq) (resp *types.AddPointsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
