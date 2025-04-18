package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	_ "embed"

	"github.com/daodao97/goreact/base/login"
	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/goreact/i18n"
	"github.com/daodao97/goreact/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"rogchap.com/v8go"
)

//go:embed templates
var Templates embed.FS

var functions template.FuncMap = template.FuncMap{
	"convertToJson": convertToJson,
}

func convertToJson(a any) string {
	s, _ := json.Marshal(a)
	return string(s)
}

func CreateTemplateRenderer() render.HTMLRender {
	tmpl := template.New("").Funcs(functions)

	tmpl, err := tmpl.ParseFS(Templates, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	return &TemplateRenderer{
		templates:  tmpl,
		reactCache: map[string]*ReactRenderer{},
		ginContext: nil,
	}
}

type TemplateRenderer struct {
	templates  *template.Template
	reactCache map[string]*ReactRenderer
	ginContext *gin.Context
}

func (t *TemplateRenderer) SetGinContext(c *gin.Context) {
	t.ginContext = c
}

func (t *TemplateRenderer) RenderReact(c *gin.Context, fragment string, data any) (template.HTML, error) {
	reactFiles, err := os.ReadFile(fmt.Sprintf("./build/server/%s", fragment))
	if err != nil {
		return template.HTML(""), err
	}
	if _, ok := t.reactCache[fragment]; !ok {
		isolate := v8go.NewIsolate()
		global := v8go.NewObjectTemplate(isolate)
		ctx := v8go.NewContext(isolate, global)

		t.reactCache[fragment] = &ReactRenderer{
			ctx:     ctx,
			content: string(reactFiles),
			name:    fragment,
		}
	}

	return t.reactCache[fragment].Ctx(c).Render(data)
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

// HTMLRender 是 HTML 渲染器的自定义实现
type HTMLRender struct {
	Template      *template.Template
	TemplateName  string
	ComponentName string
	Data          any
	renderer      *TemplateRenderer
}

// Render 实现 render.Render 接口
func (r *HTMLRender) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	var htmlContent template.HTML
	var err error

	if r.ComponentName != "" {
		htmlContent, err = r.renderer.RenderReact(r.renderer.ginContext, r.ComponentName, r.Data)
		if err != nil {

			return r.Template.ExecuteTemplate(w, "error.html", map[string]any{
				"Title":         "服务端渲染失败",
				"ErrorMessage":  fmt.Sprintf("错误：渲染 React 时出错 %+v\n", err),
				"ComponentName": r.ComponentName,
				"RequestInfo":   r.renderer.ginContext.Request.URL.Path,
			})
		}
	}

	data := extendPayload(r.Data, r.TemplateName, r.ComponentName, htmlContent)

	data.Translations = i18n.GetTranslations(r.renderer.ginContext)
	data.Lang = r.renderer.ginContext.GetString("lang")
	data.Website = &conf.Get().Website
	userInfo, err := login.GetUserInfo(r.renderer.ginContext)
	if err == nil {
		data.UserInfo = userInfo
	}

	data.GoogleAdsTxt = conf.Get().GoogleAdsTxt
	data.GoogleAdsJS = conf.Get().GoogleAdsJS
	data.GoogleAnalytics = conf.Get().GoogleAnalytics
	data.MicrosoftClarityId = conf.Get().MicrosoftClarityId
	if val, ok := r.renderer.ginContext.Get("head"); ok {
		if head, ok := val.(*model.Head); ok {
			data.Head = head
		}
	}

	data.Version = conf.Get().GitTag

	return r.Template.ExecuteTemplate(w, r.TemplateName, data)
}

// WriteContentType 设置内容类型
func (r *HTMLRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/html; charset=utf-8"}
	}
}

type GeneralPayload struct {
	Translations       any
	Payload            any
	Template           string
	TemplateID         string
	ServerURL          string
	Component          string
	InnerHtmlContent   template.HTML
	Lang               string
	Website            *conf.Website
	UserInfo           any
	GoogleAdsTxt       string
	GoogleAdsJS        string
	GoogleAnalytics    string
	MicrosoftClarityId string
	Head               *model.Head
	Version            string
}

func extendPayload(
	data any,
	name string,
	component string,
	htmlContent template.HTML,
) *GeneralPayload {
	templateID := strings.ReplaceAll(name, "/", "-")
	templateID = strings.ReplaceAll(templateID, ".html", "")

	return &GeneralPayload{
		Payload:          data,
		Template:         name,
		TemplateID:       templateID,
		Component:        component,
		InnerHtmlContent: htmlContent,
	}
}

// 需要添加的Close方法
func (r *ReactRenderer) Close() {
	if r.ctx != nil {
		r.ctx.Close()
		r.ctx = nil
	}
}
