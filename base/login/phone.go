package login

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/xgo/xredis"
	"github.com/gin-gonic/gin"
)

var phoneVerificationCodeCacheKey = "phone_verification_code:%s"

type PhoneLoginRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type PhoneLoginResponse struct {
	Token string `json:"token"`
}

func SendPhoneVerificationCodeHandler(c *gin.Context) {
	var req PhoneLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is required"})
		return
	}

	if phoneLoginService == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone login service not set"})
		return
	}

	err := phoneLoginService.SendVerificationCode(req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = sendPhoneVerificationCode(req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent"})
}

func PhoneLoginHandler(c *gin.Context) {
	var req PhoneLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is required"})
		return
	}

	if req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
		return
	}

	if phoneLoginService == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone login service not set"})
		return
	}

	err := verifyPhoneCode(req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userInfo := map[string]string{
		"phone":   req.Phone,
		"channel": "phone",
	}

	token, err := handleUserLogin(c, userInfo, conf.Get().JwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle user login"})
		return
	}

	userInfo["token"] = token

	c.JSON(http.StatusOK, userInfo)
}

func sendPhoneVerificationCode(phone string, code string) error {
	cacheKey := fmt.Sprintf(phoneVerificationCodeCacheKey, phone)
	err := xredis.Get().Set(context.Background(), cacheKey, code, time.Minute*5).Err()
	if err != nil {
		return err
	}
	return nil
}

func verifyPhoneCode(phone string, code string) error {
	cacheKey := fmt.Sprintf(phoneVerificationCodeCacheKey, phone)
	cacheCode, err := xredis.Get().Get(context.Background(), cacheKey).Result()
	if err != nil {
		return err
	}
	if cacheCode != code {
		return errors.New("invalid code")
	}
	return nil
}

type PhoneLoginService interface {
	SendVerificationCode(phone string, code string) error
}

var phoneLoginService PhoneLoginService

func SetPhoneLoginService(service PhoneLoginService) {
	phoneLoginService = service
}
