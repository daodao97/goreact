package login

import (
	"fmt"
	"time"

	"crypto/rsa"
	"encoding/base64"
	"math/big"

	"github.com/daodao97/goreact/conf"

	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xrequest"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Google OAuth2 配置
const (
	GoogleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"
)

// GoogleClaims 定义 JWT claims 结构
type GoogleClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func GoogleCallbackHandler(c *gin.Context) {
	authProvider := GetProvider("google")
	if authProvider == nil {
		c.JSON(401, gin.H{"error": "Google auth provider not found"})
		return
	}

	// 定义请求体结构
	type GoogleCallback struct {
		ClientID   string `json:"clientId"`
		Credential string `json:"credential"`
		SelectBy   string `json:"select_by"`
	}

	var callback GoogleCallback
	if err := c.ShouldBindJSON(&callback); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	certsResp, err := xrequest.New().Get(GoogleCertsURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get Google certs"})
		return
	}

	// 解析证书
	var certs struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := certsResp.Json().Unmarshal(&certs); err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse Google certs"})
		return
	}

	// 解析未验证的 token
	token, err := jwt.ParseWithClaims(callback.Credential, &GoogleClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证算法
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// 获取 token header 中的 kid
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}

		// 查找对应的公钥
		for _, cert := range certs.Keys {
			if cert.Kid == kid {
				// 解码 modulus
				nBytes, err := base64.RawURLEncoding.DecodeString(cert.N)
				if err != nil {
					return nil, fmt.Errorf("failed to decode modulus: %v", err)
				}
				n := new(big.Int).SetBytes(nBytes)

				// 解码 exponent
				eBytes, err := base64.RawURLEncoding.DecodeString(cert.E)
				if err != nil {
					return nil, fmt.Errorf("failed to decode exponent: %v", err)
				}
				e := new(big.Int).SetBytes(eBytes)

				// 构造 RSA 公钥
				return &rsa.PublicKey{
					N: n,
					E: int(e.Int64()),
				}, nil
			}
		}
		return nil, fmt.Errorf("key not found")
	})

	xlog.Info("token", xlog.Any("token", token), xlog.Any("err", err))

	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	if claims, ok := token.Claims.(*GoogleClaims); ok && token.Valid {
		// 验证 iss
		if claims.Issuer != "https://accounts.google.com" {
			c.JSON(401, gin.H{"error": "Invalid issuer"})
			return
		}

		// 验证 aud
		if claims.Audience[0] != authProvider.ClientID {
			c.JSON(401, gin.H{"error": "Invalid audience"})
			return
		}

		// 验证过期时间
		if claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(401, gin.H{"error": "Token expired"})
			return
		}

		userInfo := map[string]string{
			"email":      claims.Email,
			"name":       claims.Name,
			"avatar_url": claims.Picture,
			"channel":    "google",
		}

		token, err := handleUserLogin(c, userInfo, conf.Get().JwtSecret)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to handle user login"})
			return
		}

		userInfo["token"] = token

		c.JSON(200, userInfo)
		return
	}

	c.JSON(401, gin.H{"error": "Invalid token"})
}
