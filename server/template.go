package server

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	_ "embed"

	"github.com/daodao97/xgo/xlog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"rogchap.com/v8go"
)

type TemplateOptions struct {
	Cache *TemplateCache
}

func WithCache(cache *TemplateCache) func(*TemplateOptions) {
	return func(options *TemplateOptions) {
		options.Cache = cache
	}
}

func CreateTemplateRenderer(opts ...func(*TemplateOptions)) render.HTMLRender {
	tmpl := template.New("").Funcs(functions)

	tmpl, err := tmpl.ParseFS(Templates, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	xlog.Info("CreateTemplateRenderer", "tmpl", tmpl)

	options := &TemplateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	cache := options.Cache

	return &TemplateRenderer{
		templates:  tmpl,
		ginContext: nil,
		cache:      cache,
	}
}

type TemplateRenderer struct {
	templates  *template.Template
	ginContext *gin.Context
	cache      *TemplateCache
}

func (t *TemplateRenderer) SetGinContext(c *gin.Context) {
	t.ginContext = c
}

func (t *TemplateRenderer) Close() {
}

func (t *TemplateRenderer) RenderReact(c *gin.Context, fragment string, data any) (template.HTML, error) {
	reactContent, err := os.ReadFile(fmt.Sprintf("./build/server/%s", fragment))
	if err != nil {
		return template.HTML(""), err
	}

	start := time.Now()

	defer func() {
		xlog.Debug("RenderReact render end", xlog.String("path", c.Request.URL.Path), xlog.Any("fragment", fragment), xlog.Any("data", data), xlog.Any("time", time.Since(start)))
	}()

	cacheKey, err := t.cache.GenerateKey(fragment, data)
	if err == nil && t.cache != nil {
		if cachedHTML, found := t.cache.Load(cacheKey); found {
			xlog.Debug("Using cached render result", xlog.String("path", c.Request.URL.Path), xlog.Any("fragment", fragment), xlog.Any("cacheKey", cacheKey))
			return cachedHTML, nil
		}
	}

	// 从池中获取 isolate
	isolate := v8go.NewIsolate()
	defer isolate.Dispose()

	// 使用获取的 isolate 创建上下文
	global := v8go.NewObjectTemplate(isolate)
	ctx := v8go.NewContext(isolate, global)
	defer ctx.Close()

	render := &ReactRenderer{
		ctx:     ctx,
		content: string(reactContent),
		name:    fragment,
	}

	// 执行渲染
	html, err := render.Ctx(c).Render(data)
	if err != nil {
		return html, err
	}

	if t.cache != nil {
		if err := t.cache.Save(cacheKey, html); err != nil {
			if err := t.cache.Save(cacheKey, html); err != nil {
				xlog.Warn("Failed to save render result to cache", xlog.Any("error", err))
			}
		}
	}

	return html, nil
}

// Instance 实现 gin.HTMLRender 接口的方法
func (t *TemplateRenderer) Instance(name string, data any) render.Render {
	parts := strings.Split(name, ":")
	templateName := parts[0]
	componentName := ""

	if len(parts) > 1 {
		componentName = parts[1]
	}

	return &HTMLRender{
		Template:      t.templates,
		TemplateName:  templateName,
		ComponentName: componentName,
		Data:          data,
		renderer:      t,
	}
}
