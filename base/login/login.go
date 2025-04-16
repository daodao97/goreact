package login

import (
	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xjwt"
	"github.com/gin-gonic/gin"
)

func GetProvider(providerName string) *conf.AuthProvider {
	for _, provider := range conf.Get().Website.AuthProvider {
		if provider.Provider == conf.AuthProviderType(providerName) {
			return &provider
		}
	}
	return nil
}

func handleUserLogin(c *gin.Context, userInfo map[string]string, jwtSecret string) error {
	userId, err := CreateUserOrIgnore(xdb.Record{
		"email":      userInfo["email"],
		"user_name":  userInfo["name"],
		"avatar_url": userInfo["avatar_url"],
		"channel":    userInfo["channel"],
	})

	if err != nil {
		return err
	}

	payload := map[string]any{
		"id":         userId,
		"email":      userInfo["email"],
		"user_name":  userInfo["name"],
		"avatar_url": userInfo["avatar_url"],
	}

	token, err := xjwt.GenHMacToken(payload, jwtSecret)
	if err != nil {
		return err
	}

	c.SetCookie("session_token", token, 3600*24*30, "/", "", false, true)
	return nil
}

var UserModel xdb.Model

func SetUserMoel(m xdb.Model) {
	UserModel = m
}

func CreateUserOrIgnore(user xdb.Record) (int64, error) {
	existing, _ := UserModel.First(
		xdb.WhereEq("email", user.GetString("email")),
		xdb.WhereEq("appid", conf.Get().AppID),
	)
	if existing != nil {
		return int64(existing.GetInt("id")), nil
	}
	user["appid"] = conf.Get().AppID

	uid, err := UserModel.Insert(user)
	if err != nil {
		return 0, err
	}
	if OnNewRegisterFunc != nil {
		user["id"] = uid
		OnNewRegisterFunc(user)
	}
	return uid, nil
}

func GetUserInfo(c *gin.Context) (xdb.Record, error) {
	cookie, err := c.Cookie("session_token")
	if err != nil {
		return nil, err
	}

	claims, err := xjwt.VerifyHMacToken(cookie, conf.Get().JwtSecret)
	if err != nil {
		return nil, err
	}

	userInfo := xdb.Record{
		"email":      claims["email"],
		"name":       claims["user_name"],
		"avatar_url": claims["avatar_url"],
		"id":         claims["id"],
	}

	return userInfo, nil
}

type OnNewRegister func(user xdb.Record)

var OnNewRegisterFunc OnNewRegister

func SetOnNewRegister(fn OnNewRegister) {
	OnNewRegisterFunc = fn
}
