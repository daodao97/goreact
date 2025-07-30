package invite

import (
	"github.com/daodao97/goreact/auth"
	"github.com/gin-gonic/gin"
)

func SetupRouter(app *gin.Engine) {
	invite := app.Group("/invite")
	invite.Use(auth.AuthMiddleware())
	invite.GET("/get", GetUserInviteCode)
	invite.POST("/set", SetUserInviteCode)
	invite.GET("/list", InvitedList)
}
