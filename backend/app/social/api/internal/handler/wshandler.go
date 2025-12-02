// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/social/api/internal/logic"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/social/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func WsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewWsLogic(r.Context(), svcCtx)
		err := l.Ws()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
