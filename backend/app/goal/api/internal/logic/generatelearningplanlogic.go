// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateLearningPlanLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateLearningPlanLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateLearningPlanLogic {
	return &GenerateLearningPlanLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateLearningPlanLogic) GenerateLearningPlan(req *types.GeneratePlanReq) (resp *types.GeneratePlanResp, err error) {
	// todo: add your logic here and delete this line

	return
}
