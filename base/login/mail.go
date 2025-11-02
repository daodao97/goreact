package login

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/goreact/dao"
	"github.com/daodao97/goreact/util/mail"
	"github.com/daodao97/xgo/xadmin"
	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xredis"

	"github.com/gin-gonic/gin"
)

var VerificationCodeKey = "verification_code:%s"
var VerificationCodeSubject = "注册验证码"
var VerificationCodePlainTextContent = "您好，邮箱验证码为: %s\n验证码10分钟有效期。如非本人操作，请忽略本邮件"

var VerificationCodeMailSender *CodeSender

func SetVerificationCodeMailSender(sender *CodeSender) {
	if sender.Subject == "" {
		sender.Subject = VerificationCodeSubject
	}
	if sender.PlainTextContent == "" {
		sender.PlainTextContent = VerificationCodePlainTextContent
	}
	VerificationCodeMailSender = sender
}

func GetVerificationCodeMailSender() *CodeSender {
	return VerificationCodeMailSender
}

type CodeSender struct {
	From             string
	Subject          string
	PlainTextContent string
	HtmlContent      string
	MailSender       mail.MailSender
}

func (c *CodeSender) Send(to string, code string) error {
	return c.MailSender.SendEmail(c.From, to, c.Subject, fmt.Sprintf(c.PlainTextContent, code), c.HtmlContent)
}

type MailCallbackRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"verificationCode"`
	Mode     string `json:"mode"` // register or login or reset
}

func MailCallbackHandler(c *gin.Context) {
	var request MailCallbackRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request",
		})
		return
	}

	request.Email = strings.TrimSpace(request.Email)
	if request.Mode == "register" && isEmailBlacklisted(request.Email) {
		c.JSON(400, gin.H{
			"message": "邮箱暂不支持注册",
		})
		return
	}

	var user xdb.Record
	var err error

	if request.Mode == "register" {
		user, err = Register(request.Email, request.Password, request.Code)
	} else if request.Mode == "login" {
		user, err = Login(request.Email, request.Password, request.Code)
	} else if request.Mode == "reset" {
		user, err = ResetPassword(request.Email, request.Password, request.Code)
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

	token, err := handleUserLogin(c, payload, conf.Get().JwtSecret)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to handle user login"})
		return
	}

	payload["token"] = token

	c.JSON(200, payload)
}

func isEmailBlacklisted(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}

	cfg := conf.Get()
	if cfg == nil {
		return false
	}

	blacklist := cfg.EmailBlacklist
	lowerEmail := strings.ToLower(email)

	for _, exact := range blacklist.Exact {
		if strings.EqualFold(strings.TrimSpace(exact), email) {
			return true
		}
	}

	for _, suffix := range blacklist.Suffixes {
		suffix = strings.TrimSpace(suffix)
		if suffix == "" {
			continue
		}
		lowerSuffix := strings.ToLower(suffix)
		if !strings.HasPrefix(lowerSuffix, "@") {
			lowerSuffix = "@" + lowerSuffix
		}
		if strings.HasSuffix(lowerEmail, lowerSuffix) {
			return true
		}
	}

	for _, keyword := range blacklist.Keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}
		if strings.Contains(lowerEmail, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

func Register(email string, password string, code string) (_user xdb.Record, err error) {
	err = VerifyCode(email, code)
	if err != nil {
		return nil, err
	}

	user, err := dao.UserModel.First(xdb.WhereEq("email", email), xdb.WhereEq("appid", conf.Get().AppID))
	if err != nil && err != xdb.ErrNotFound {
		return nil, err
	}
	if user != nil {
		return nil, errors.New("user already exists")
	}

	id, err := dao.UserModel.Insert(xdb.Record{
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
	user, err := dao.UserModel.First(xdb.WhereEq("email", email), xdb.WhereEq("appid", conf.Get().AppID))
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

func ResetPassword(email string, password string, code string) (_user xdb.Record, err error) {
	err = VerifyCode(email, code)
	if err != nil {
		return nil, err
	}

	user, err := dao.UserModel.First(xdb.WhereEq("email", email), xdb.WhereEq("appid", conf.Get().AppID))
	if err != nil && err != xdb.ErrNotFound {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	_, err = dao.UserModel.Update(xdb.Record{
		"password": xadmin.PasswordHash(password),
	}, xdb.WhereEq("id", user.GetInt("id")))
	if err != nil {
		return nil, err
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

	request.Email = strings.TrimSpace(request.Email)
	if isEmailBlacklisted(request.Email) {
		c.JSON(400, gin.H{
			"message": "邮箱暂不支持注册",
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
	email = strings.TrimSpace(email)
	if isEmailBlacklisted(email) {
		return errors.New("邮箱暂不支持注册")
	}
	if VerificationCodeMailSender == nil {
		return errors.New("verification code mail sender not set")
	}
	code := GenerateVerificationCode()

	err := VerificationCodeMailSender.Send(email, code)
	if err != nil {
		return err
	}
	xlog.Debug("SendVerificationCode", xlog.Any("email", email), xlog.Any("code", code))
	xredis.Get().Set(context.Background(), fmt.Sprintf(VerificationCodeKey, email), code, time.Minute*10)
	return nil
}
