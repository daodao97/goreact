package server

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "embed"

	"github.com/daodao97/xgo/xlog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
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

// TemplateRenderer 模板引擎
type TemplateRenderer struct {
	templates  *template.Template
	ginContext *gin.Context
	cache      *TemplateCache
}

func (t *TemplateRenderer) SetGinContext(c *gin.Context) {
	t.ginContext = c
}

func (t *TemplateRenderer) RenderReact(c *gin.Context, fragment string, data any) (template.HTML, error) {
	reactContent, err := os.ReadFile(filepath.Join(buildServerDir, fragment))
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

	render := &ReactRenderer{
		engine:  NewV8JsEngine(),
		content: string(reactContent),
		name:    fragment,
	}
	defer render.Close()

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
	componentName := ""
	templateName := ""

	if strings.Contains(name, ":") {
		parts := strings.Split(name, ":")
		templateName = parts[0]
		if len(parts) > 1 {
			componentName = parts[1]
		}
	} else {
		templateName = "index.html"
		componentName = name
	}

	return &HTMLRender{
		Template:      t.templates,
		TemplateName:  templateName,
		ComponentName: componentName,
		Data:          data,
		renderer:      t,
	}
}
