// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"encoding/json"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/user/api/internal/types"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLogic) GetUser() (resp *types.UserInfoResp, err error) {
	userID, ok := l.ctx.Value("user_id").(json.Number)
	var uid int64
	if ok {
		uid, _ = userID.Int64()
	} else {
		if id, ok := l.ctx.Value("user_id").(int); ok {
			uid = int64(id)
		} else if id, ok := l.ctx.Value("user_id").(float64); ok {
			uid = int64(id)
		}
	}

	if uid == 0 {
		return nil, nil
	}

	user, err := repository.GetUserByID(uint(uid))
	if err != nil {
		return nil, err
	}

	return &types.UserInfoResp{
		Id:             user.ID,
		UserId:         user.ID,
		Username:       user.Name,
		Name:           user.Name,
		Email:          user.Email,
		HeadShow:       user.HeadShow,
		Daka:           user.Daka,
		FlagNumber:     user.FlagNumber,
		Count:          user.Count,
		MonthLearnTime: user.MonthLearntime,
	}, nil
}
