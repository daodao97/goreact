package server

import (
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/daodao97/goreact/base/login"
	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/goreact/i18n"
	"github.com/gin-gonic/gin"
	"rogchap.com/v8go"
)

// ReactRenderer 是 ReactRenderer 的 Gin 版本
type ReactRenderer struct {
	ctx     *v8go.Context // 使用 v8go 的上下文
	content string        // 组件的 JavaScript 内容
	name    string        // 组件的名称
	ginCtx  *gin.Context  // Gin 的上下文
}

// Ctx 设置 gin.Context
func (render *ReactRenderer) Ctx(c *gin.Context) *ReactRenderer {
	render.ginCtx = c
	return render
}

// Render 渲染 React 组件
func (renderer *ReactRenderer) Render(data any) (template.HTML, error) {
	params, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		return "", err
	}

	_, err = renderer.ctx.RunScript(renderer.content, renderer.name)
	if err != nil {
		fmt.Printf("错误：运行组件时出错 %+v\n", err)
		return "", err
	}

	locationScript := fmt.Sprintf(`
	globalThis.window = globalThis.window || {};
	globalThis.window.location = {
	  hostname: "%s",
	  protocol: "%s",
	  origin: "%s",
	  search: "%s",
	  pathname: "%s",
	  hash: "%s",
	  port: "%s",
	  host: "%s",
	  href: "%s",
	  assign: function(url) { console.log("Location assign:", url); },
	  replace: function(url) { console.log("Location replace:", url); },
	  reload: function(force) { console.log("Location reload:", force); }
	};
	`,
		renderer.ginCtx.Request.Host,
		renderer.ginCtx.Request.URL.Scheme,
		renderer.ginCtx.Request.URL.Host,
		renderer.ginCtx.Request.URL.RawQuery,
		renderer.ginCtx.Request.URL.Path,
		renderer.ginCtx.Request.URL.Fragment,
		renderer.ginCtx.Request.URL.Port(),
		renderer.ginCtx.Request.URL.Host,
		renderer.ginCtx.Request.URL.String())

	_, err = renderer.ctx.RunScript(locationScript, "set-location.js")
	if err != nil {
		fmt.Printf("错误：设置 location 时出错 %+v\n", err)
	}

	_, err = renderer.ctx.RunScript("window.INITIAL_PROPS = "+string(params), "params.js")
	if err != nil {
		fmt.Printf("错误：设置属性时出错 %+v\n", err)
		return "", err
	}

	translations := i18n.GetTranslations(renderer.ginCtx)
	translationsJSON, err := json.Marshal(translations)
	if err != nil {
		fmt.Printf("错误：序列化翻译时出错 %+v\n", err)
		return "", err
	}

	_, err = renderer.ctx.RunScript("window.TRANSLATIONS = "+string(translationsJSON), "translations.js")
	if err != nil {
		fmt.Printf("错误：设置翻译时出错 %+v\n", err)
		return "", err
	}

	websiteJSON, err := json.Marshal(conf.Get().Website)
	if err != nil {
		fmt.Printf("错误：序列化网站配置时出错 %+v\n", err)
		return "", err
	}

	_, err = renderer.ctx.RunScript("window.WEBSITE = "+string(websiteJSON), "website.js")
	if err != nil {
		fmt.Printf("错误：设置网站配置时出错 %+v\n", err)
		return "", err
	}

	lang := renderer.ginCtx.GetString("lang")

	_, err = renderer.ctx.RunScript("window.LANG = '"+lang+"'", "lang.js")
	if err != nil {
		fmt.Printf("错误：设置语言时出错 %+v\n", err)
		return "", err
	}

	userInfo, err := login.GetUserInfo(renderer.ginCtx)
	if err == nil {
		userInfoJSON, _ := json.Marshal(userInfo)
		_, err = renderer.ctx.RunScript("window.USER_INFO = "+string(userInfoJSON), "user_info.js")
		if err != nil {
			return "", err
		}
	}

	renderer.ctx.RunScript("window.ssr = true", "ssr.js")

	val, err := renderer.ctx.RunScript("Render()", "render.js")
	if err != nil {
		return "", err
	}

	html := template.HTML(val.String())

	return html, nil
}
