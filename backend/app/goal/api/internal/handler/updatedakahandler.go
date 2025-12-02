// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/logic"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdateDakaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUpdateDakaLogic(r.Context(), svcCtx)
		resp, err := l.UpdateDaka()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
