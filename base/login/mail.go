package login

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/xgo/xadmin"
	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xredis"

	"github.com/gin-gonic/gin"
)

var VerificationCodeKey = "verification_code:%s"
var VerificationCodeSubject = "注册验证码"
var VerificationCodePlainTextContent = "您好，邮箱验证码为: %s\n验证码10分钟有效期。如非本人操作，请忽略本邮件"

type MailCallbackRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"verificationCode"`
	Mode     string `json:"mode"`
}

func MailCallbackHandler(c *gin.Context) {
	var request MailCallbackRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request",
		})
		return
	}

	var user xdb.Record
	var err error

	if request.Mode == "register" {
		user, err = Register(request.Email, request.Password, request.Code)
	} else if request.Mode == "login" {
		user, err = Login(request.Email, request.Password, request.Code)
	}

	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	payload := map[string]string{
		"id":         user.GetString("id"),
		"email":      user.GetString("email"),
		"user_name":  user.GetString("user_name"),
		"avatar_url": user.GetString("avatar_url"),
	}

	if err := handleUserLogin(c, payload, conf.Get().JwtSecret); err != nil {
		c.JSON(500, gin.H{"error": "Failed to handle user login"})
		return
	}

	c.JSON(200, payload)
}

func Register(email string, password string, code string) (_user xdb.Record, err error) {
	err = VerifyCode(email, code)
	if err != nil {
		return nil, err
	}

	user, err := UserModel.First(xdb.WhereEq("email", email), xdb.WhereEq("appid", conf.Get().AppID))
	if err != nil && err != xdb.ErrNotFound {
		return nil, err
	}
	if user != nil {
		return nil, errors.New("user already exists")
	}

	id, err := UserModel.Insert(xdb.Record{
		"email":     email,
		"user_name": email,
		"password":  xadmin.PasswordHash(password),
		"appid":     conf.Get().AppID,
		"channel":   "mail",
	})
	if err != nil {
		return nil, err
	}

	return xdb.Record{
		"id":         id,
		"email":      email,
		"user_name":  email,
		"avatar_url": "",
	}, nil
}

func Login(email string, password string, code string) (_user xdb.Record, err error) {
	user, err := UserModel.First(xdb.WhereEq("email", email), xdb.WhereEq("appid", conf.Get().AppID))
	if err != nil && err != xdb.ErrNotFound {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if !xadmin.PasswordVerify(password, user.GetString("password")) {
		return nil, errors.New("invalid password")
	}

	return xdb.Record{
		"id":         user.GetInt("id"),
		"email":      user.GetString("email"),
		"user_name":  user.GetString("user_name"),
		"avatar_url": user.GetString("avatar_url"),
	}, nil
}

func SendVerificationCodeHandler(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request",
		})
		return
	}

	err := SendVerificationCode(request.Email)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
	})
}

func VerifyCode(email string, code string) error {
	exists, err := xredis.Get().Exists(context.Background(), fmt.Sprintf(VerificationCodeKey, email)).Result()
	if err != nil {
		return fmt.Errorf("failed to check if verification code exists: %w", err)
	}
	if exists == 0 {
		return errors.New("verification code does not exist")
	}
	cacheCode, err := xredis.Get().Get(context.Background(), fmt.Sprintf(VerificationCodeKey, email)).Result()
	if err != nil {
		return fmt.Errorf("failed to get verification code: %w", err)
	}
	xlog.Debug("VerifyCode", xlog.Any("email", email), xlog.Any("cacheCode", cacheCode), xlog.Any("code", code))
	if cacheCode != code {
		return errors.New("invalid code")
	}
	return nil
}

func GenerateVerificationCode() string {
	const letters = "0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, 6)
	for i := range code {
		code[i] = letters[r.Intn(len(letters))]
	}
	return string(code)
}

func SendVerificationCode(email string) error {
	code := GenerateVerificationCode()
	err := GetMailSender().SendVerificationCode(email, VerificationCodeSubject, fmt.Sprintf(VerificationCodePlainTextContent, code))
	if err != nil {
		return err
	}
	xlog.Debug("SendVerificationCode", xlog.Any("email", email), xlog.Any("code", code))
	xredis.Get().Set(context.Background(), fmt.Sprintf(VerificationCodeKey, email), code, time.Minute*10)
	return nil
}
