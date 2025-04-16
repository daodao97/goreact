package auth

import (
	"context"
	"net/http"

	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/goreact/dao"
	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xjwt"
	"github.com/daodao97/xgo/xlog"
	"github.com/gin-gonic/gin"
)

type authContextKey struct{}

// gin middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从 cookie 获取 token
		cookieToken, _ := c.Cookie("session_token")

		// 尝试从 header 获取 API token
		headerToken := c.GetHeader("X-API-KEY")

		// 如果 cookie 和 header 都没有 token，返回未授权错误
		if cookieToken == "" && headerToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Need Login"})
			c.Abort()
			return
		}

		var payload map[string]interface{}
		var verifyErr error

		// 尝试验证 cookie token
		if cookieToken != "" {
			payload, verifyErr = xjwt.VerifyHMacToken(cookieToken, conf.Get().JwtSecret)
			if verifyErr != nil {
				xlog.ErrorCtx(c, "auth", xlog.Any("verifyErr", verifyErr))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session token"})
				c.Abort()
				return
			}
		}

		// 如果 cookie token 验证失败，尝试验证 API token
		if headerToken != "" {
			apiToken, apiErr := dao.GetApiTokenByToken(headerToken)
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

		xlog.InfoCtx(c, "auth", xlog.Any("payload", payload))
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
