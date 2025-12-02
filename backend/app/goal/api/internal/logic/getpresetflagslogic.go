// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPresetFlagsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPresetFlagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPresetFlagsLogic {
	return &GetPresetFlagsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPresetFlagsLogic) GetPresetFlags() (resp *types.GetUserFlagsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
