package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/daodao97/goreact/base/login"
	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/goreact/i18n"
	"github.com/daodao97/xgo/xapp"
)

// HTMLRender 是 HTML 渲染器的自定义实现
// react js 生成的 html 通过 go template 渲染
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
				"IsDev":         xapp.IsDev(),
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
	data.Head = i18n.GetHead(r.renderer.ginContext, strings.ToLower(strings.TrimSuffix(r.ComponentName, ".js")))

	data.Version = conf.Get().GitTag

	// 先执行模板渲染，然后再释放资源
	err = r.Template.ExecuteTemplate(w, r.TemplateName, data)

	return err
}

// WriteContentType 设置内容类型
func (r *HTMLRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/html; charset=utf-8"}
	}
}
