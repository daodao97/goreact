package server

import (
	"net/http"

	"github.com/daodao97/xgo/xapp"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Gin() *gin.Engine {
	r := xapp.NewGin()

	r.Static("/assets", "build")

	// 设置模板渲染器
	var opts []func(*TemplateOptions)
	// if !xapp.IsDev() {
	// 	opts = append(opts, WithCache(NewTemplateCache()))
	// }
	r.HTMLRender = CreateTemplateRenderer(opts...)

	r.Use(SetRendererContextMiddleware(r.HTMLRender.(*TemplateRenderer)))

	// CORS 配置
	corsConfig := cors.Config{
		AllowAllOrigins: true,
		AllowHeaders: []string{
			"Accept",
			"Content-Type",
			"Content-Length",
			"X-Custom-Header",
			"Origin",
			"Authorization",
			"X-Trace-ID",
			"Trace-Id",
			"x-request-id",
			"X-Request-ID",
			"TraceID",
			"ParentID",
			"Uber-Trace-ID",
			"uber-trace-id",
			"traceparent",
			"tracestate",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
			http.MethodOptions,
		},
	}
	r.Use(cors.New(corsConfig))

	// 安全中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	})

	if xapp.IsDev() {
		setupDev(r)
	}

	return r
}

func SetRendererContextMiddleware(renderer *TemplateRenderer) gin.HandlerFunc {
	return func(c *gin.Context) {
		renderer.SetGinContext(c)
		c.Next()
	}
}
