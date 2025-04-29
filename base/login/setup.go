package login

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(app *gin.RouterGroup) {
	app.POST("/login/google", GoogleCallbackHandler)
	app.POST("/login/github", GithubCallbackHandler)
	app.POST("/login/mail", MailCallbackHandler)
	app.POST("/login/send-verification-code", SendVerificationCodeHandler)
	app.GET("/logout", LogoutHandler)
}
