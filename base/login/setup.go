package login

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(app *gin.RouterGroup) {
	app.POST("/login/google", GoogleCallbackHandler)
	app.POST("/login/github", GithubCallbackHandler)
	app.GET("/logout", LogoutHandler)
}
