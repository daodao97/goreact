package server

import (
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/daodao97/goreact/base/login"
	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/goreact/i18n"
	"github.com/gin-gonic/gin"
)

// 将 react js 转换为 html
type ReactRenderer struct {
	engine  JsEngine
	content string       // 组件的 JavaScript 内容
	name    string       // 组件的名称
	ginCtx  *gin.Context // Gin 的上下文
}

func (render *ReactRenderer) Ctx(c *gin.Context) *ReactRenderer {
	render.ginCtx = c
	return render
}

func (r *ReactRenderer) Close() {
	r.engine.Close()
}

// Render 渲染 React 组件
func (renderer *ReactRenderer) Render(data any) (template.HTML, error) {
	params, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		return "", err
	}

	_, err = renderer.engine.RunScript(renderer.content, renderer.name)
	if err != nil {
		return "", fmt.Errorf("render component failed: name=%s\n, err=%w", renderer.name, err)
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

	_, err = renderer.engine.RunScript(locationScript, "set-location.js")
	if err != nil {
		return "", fmt.Errorf("set location failed: err=%w", err)
	}

	_, err = renderer.engine.RunScript("window.INITIAL_PROPS = "+string(params), "params.js")
	if err != nil {
		return "", fmt.Errorf("set initial props failed: err=%w", err)
	}

	translations := i18n.GetTranslations(renderer.ginCtx)
	translationsJSON, err := json.Marshal(translations)
	if err != nil {
		return "", fmt.Errorf("serialize translations failed: err=%w", err)
	}

	_, err = renderer.engine.RunScript("window.TRANSLATIONS = "+string(translationsJSON), "translations.js")
	if err != nil {
		return "", fmt.Errorf("set translations failed: err=%w", err)
	}

	websiteJSON, err := json.Marshal(conf.Get().Website)
	if err != nil {
		return "", fmt.Errorf("serialize website failed: err=%w", err)
	}

	_, err = renderer.engine.RunScript("window.WEBSITE = "+string(websiteJSON), "website.js")
	if err != nil {
		return "", fmt.Errorf("set website failed: err=%w", err)
	}

	lang := renderer.ginCtx.GetString("lang")

	_, err = renderer.engine.RunScript("window.LANG = '"+lang+"'", "lang.js")
	if err != nil {
		return "", fmt.Errorf("set language failed: err=%w", err)
	}

	userInfo, err := login.GetUserInfo(renderer.ginCtx)
	if err == nil {
		userInfoJSON, _ := json.Marshal(userInfo)
		_, err = renderer.engine.RunScript("window.USER_INFO = "+string(userInfoJSON), "user_info.js")
		if err != nil {
			return "", fmt.Errorf("set user info failed: err=%w", err)
		}
	}

	_, err = renderer.engine.RunScript("window.ssr = true", "ssr.js")
	if err != nil {
		return "", fmt.Errorf("set SSR failed: err=%w", err)
	}

	_, err = renderer.engine.RunScript("Render()", "render.js")
	if err != nil {
		return "", fmt.Errorf("render failed: err=%w", err)
	}

	html := template.HTML(renderer.engine.String())

	return html, nil
}
