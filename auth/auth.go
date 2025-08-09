package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/goreact/dao"
	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xjwt"
	"github.com/daodao97/xgo/xlog"
	"github.com/gin-gonic/gin"
)

type authContextKey struct{}

type AuthOption struct {
	NotAbort bool
}

type AuthOptionFunc func(*AuthOption)

func WithAuthOption(notAbort bool) AuthOptionFunc {
	return func(o *AuthOption) {
		o.NotAbort = notAbort
	}
}

// gin middleware
func AuthMiddleware(option ...AuthOptionFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authOption := AuthOption{}
		for _, option := range option {
			option(&authOption)
		}
		// 尝试从 cookie 获取 token
		var token string
		cookieToken, _ := c.Cookie("session_token")
		authHeader := c.GetHeader("Authorization")

		if cookieToken != "" {
			token = cookieToken
		}
		if authHeader != "" {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 尝试从 header 获取 API token
		apiKey := c.GetHeader("X-API-KEY")

		xlog.DebugCtx(c, "auth", xlog.String("cookieToken", cookieToken), xlog.String("authHeader", authHeader), xlog.String("apiKey", apiKey))

		// 如果 cookie 和 header 都没有 token，返回未授权错误
		if token == "" && apiKey == "" {
			// 如果配置了终止，则返回未授权错误
			if !authOption.NotAbort {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Need Login"})
				c.Abort()
				return
			}
			// 如果配置了不终止，则继续执行
			c.Next()
			return
		}

		var payload map[string]any
		var verifyErr error

		// 尝试验证 cookie token
		if token != "" {
			xlog.DebugCtx(c, "auth", xlog.String("token", token), xlog.String("apiid", conf.Get().AppID), xlog.String("jwt_secret", conf.Get().JwtSecret))
			payload, verifyErr = xjwt.VerifyHMacToken(token, conf.Get().JwtSecret)
			if verifyErr != nil {
				xlog.ErrorCtx(c, "auth", xlog.Any("verifyErr", verifyErr), xlog.String("apiid", conf.Get().AppID), xlog.String("jwt_secret", conf.Get().JwtSecret))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session token"})
				c.Abort()
				return
			}
		}

		// 如果 cookie token 验证失败，尝试验证 API token
		if apiKey != "" {
			apiToken, apiErr := dao.GetApiTokenByToken(apiKey)
			if apiErr != nil {
				xlog.ErrorCtx(c, "auth", xlog.Any("apiErr", apiErr))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
				c.Abort()
				return
			}

			user, err := dao.GetUserById(apiToken.GetInt("uid"))
			if err != nil {
				xlog.ErrorCtx(c, "auth", xlog.Any("err", err))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
				c.Abort()
				return
			}

			payload = xdb.Record{
				"id":         user.GetString("id"),
				"user_name":  user.GetString("user_name"),
				"email":      user.GetString("email"),
				"avatar_url": user.GetString("avatar_url"),
			}

		}

		if payload == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		xlog.DebugCtx(c, "auth", xlog.Map("payload", payload))
		c.Set("auth", payload)
		// 将 payload 添加到请求上下文
		ctx := context.WithValue(c.Request.Context(), authContextKey{}, xdb.Record(payload))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetAuthFromContext(ctx context.Context) xdb.Record {
	if _ctx, ok := ctx.(*gin.Context); ok {
		return GetAuth(_ctx)
	}
	user, _ := ctx.Value(authContextKey{}).(xdb.Record)
	return user
}

func GetAuth(c *gin.Context) xdb.Record {
	return GetAuthFromContext(c.Request.Context())
}

func GetAccountId(c *gin.Context) string {
	auth := GetAuth(c)
	return auth.GetString("accountId")
}

func IsLogin(c *gin.Context) bool {
	cookieToken, _ := c.Cookie("session_token")
	return cookieToken != ""
}
