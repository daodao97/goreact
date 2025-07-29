package login

import (
	"net/http"

	"github.com/daodao97/goreact/conf"

	"github.com/daodao97/xgo/xrequest"
	"github.com/gin-gonic/gin"
)

func GithubCallbackHandler(c *gin.Context) {
	authProvider := GetProvider("github")
	if authProvider == nil {
		c.JSON(http.StatusOK, gin.H{"error": "github auth provider not found"})
		return
	}

	code := c.Query("code")

	resp, err := xrequest.New().
		SetHeader("Accept", "application/json").
		SetFormData(map[string]string{
			"client_id":     authProvider.ClientID,
			"client_secret": authProvider.ClientSecret,
			"code":          code,
		}).
		Post("https://github.com/login/oauth/access_token")

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to get access token"})
		return
	}

	if resp.Error() != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to get access token"})
		return
	}

	errMsg := resp.Json().Get("error_description").String()
	if errMsg != "" {
		c.JSON(http.StatusOK, gin.H{"error": errMsg})
		return
	}

	accessToken := resp.Json().Get("access_token").String()

	resp, err = xrequest.New().
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Accept", "application/json").
		Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to get user info"})
		return
	}

	emailResp, err := xrequest.New().
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Accept", "application/json").
		Get("https://api.github.com/user/emails")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to get user emails"})
		return
	}

	email := emailResp.Json().Get("0.email").String()
	if email == "" {
		c.JSON(http.StatusOK, gin.H{"error": "failed to get user email"})
		return
	}

	var userInfo struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := resp.Json().Unmarshal(&userInfo); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to parse user info"})
		return
	}

	userLoginInfo := map[string]string{
		"email":      email,
		"user_name":  userInfo.Name,
		"avatar_url": userInfo.AvatarURL,
		"channel":    "github",
	}

	token, err := handleUserLogin(c, userLoginInfo, conf.Get().JwtSecret)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to handle user login"})
		return
	}

	userLoginInfo["token"] = token

	// c.JSON(http.StatusOK, userLoginInfo)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}
