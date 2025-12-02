// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/social/api/internal/logic"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/social/api/internal/svc"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/social/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteChatRoomHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteChatRoomReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewDeleteChatRoomLogic(r.Context(), svcCtx)
		resp, err := l.DeleteChatRoom(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
