package login

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LogoutHandler(ctx *gin.Context) {
	ctx.SetCookie("session_token", "", -1, "/", "", false, true)
	ctx.Redirect(http.StatusSeeOther, "/")
}
