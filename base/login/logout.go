package login

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LogoutHandler(ctx *gin.Context) {
	ctx.SetCookie("session_token", "deleted", -3600, "/", "", false, true)
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Expires", "0")
	ctx.Redirect(http.StatusSeeOther, "/")
}
