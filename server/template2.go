package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/daodao97/goreact/base/login"
	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/goreact/i18n"
	"github.com/daodao97/xgo/xlog"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

const memoryThreshold = 500 * 1024 * 1024 // 500MB

// GojaTemplateRenderer 是使用 goja 作为 JavaScript 引擎的模板渲染器
type GojaTemplateRenderer struct {
	templates  *template.Template
	reactCache map[string]*GojaReactRenderer
	ginContext *gin.Context
	pool       *ManagedRuntimePool // 运行时池
}

// GojaReactRenderer 是 ReactRenderer 的 goja 实现版本
type GojaReactRenderer struct {
	pool    *ManagedRuntimePool // 运行时池
	content string              // 组件的 JavaScript 内容
	name    string              // 组件的名称
	ginCtx  *gin.Context        // Gin 的上下文
}

// CreateGojaTemplateRenderer 创建一个新的基于 goja 的模板渲染器
func CreateGojaTemplateRenderer() render.HTMLRender {
	tmpl := template.New("").Funcs(functions)

	tmpl, err := tmpl.ParseFS(Templates, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	return &GojaTemplateRenderer{
		templates:  tmpl,
		reactCache: map[string]*GojaReactRenderer{},
		ginContext: nil,
		pool:       NewManagedRuntimePool(5),
	}
}

func (t *GojaTemplateRenderer) SetGinContext(c *gin.Context) {
	t.ginContext = c
}

// Close 清理所有React渲染器资源
func (t *GojaTemplateRenderer) Close() {
	for fragment, renderer := range t.reactCache {
		renderer.Close()
		delete(t.reactCache, fragment)
	}
}

// Ctx 设置 gin.Context
func (render *GojaReactRenderer) Ctx(c *gin.Context) *GojaReactRenderer {
	render.ginCtx = c
	return render
}

// Close 释放资源 (goja 不需要显式释放)
func (r *GojaReactRenderer) Close() {
	// goja 使用 Go 的 GC，不需要显式清理
	// r.runtime = nil
	// r.pool.Put(r.runtime)
}

func (t *GojaTemplateRenderer) RenderReact(c *gin.Context, fragment string, data any) (template.HTML, error) {
	start := time.Now()

	defer func() {
		xlog.Debug("GojaRenderReact render end", xlog.Any("fragment", fragment), xlog.Any("time", time.Since(start)))
	}()

	reactFiles, err := os.ReadFile(fmt.Sprintf("./build/server/%s", fragment))
	if err != nil {
		return template.HTML(""), err
	}

	render := &GojaReactRenderer{
		pool:    t.pool,
		content: string(reactFiles),
		name:    fragment,
	}

	return render.Ctx(c).Render(data)
}

// Render 渲染 React 组件
func (renderer *GojaReactRenderer) Render(data any) (template.HTML, error) {
	params, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		return "", err
	}

	runtime := renderer.pool.Get()
	defer renderer.pool.Put(runtime)

	// 执行 JS 内容
	_, err = runtime.RunString(renderer.content)
	if err != nil {
		fmt.Printf("错误：运行组件时出错 %+v\n", err)
		return "", err
	}

	// 在设置window.location前，检查URL是否完整
	var scheme, host, path, query, fragment string
	if renderer.ginCtx.Request.URL != nil {
		scheme = renderer.ginCtx.Request.URL.Scheme
		if scheme == "" {
			scheme = "http" // 提供默认值
		}
		host = renderer.ginCtx.Request.Host
		path = renderer.ginCtx.Request.URL.Path
		query = renderer.ginCtx.Request.URL.RawQuery
		fragment = renderer.ginCtx.Request.URL.Fragment
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
	`, host, scheme, fmt.Sprintf("%s://%s", scheme, host), query, path, fragment, renderer.ginCtx.Request.URL.Port(), host, renderer.ginCtx.Request.URL.String())

	_, err = runtime.RunString(locationScript)
	if err != nil {
		fmt.Printf("错误：设置 location 时出错 %+v\n", err)
	}

	// 设置初始属性
	_, err = runtime.RunString("window.INITIAL_PROPS = " + string(params))
	if err != nil {
		fmt.Printf("错误：设置属性时出错 %+v\n", err)
		return "", err
	}

	// 设置翻译
	translations := i18n.GetTranslations(renderer.ginCtx)
	translationsJSON, err := json.Marshal(translations)
	if err != nil {
		fmt.Printf("错误：序列化翻译时出错 %+v\n", err)
		return "", err
	}

	_, err = runtime.RunString("window.TRANSLATIONS = " + string(translationsJSON))
	if err != nil {
		fmt.Printf("错误：设置翻译时出错 %+v\n", err)
		return "", err
	}

	// 设置网站配置
	websiteJSON, err := json.Marshal(conf.Get().Website)
	if err != nil {
		fmt.Printf("错误：序列化网站配置时出错 %+v\n", err)
		return "", err
	}

	_, err = runtime.RunString("window.WEBSITE = " + string(websiteJSON))
	if err != nil {
		fmt.Printf("错误：设置网站配置时出错 %+v\n", err)
		return "", err
	}

	// 设置语言
	lang := renderer.ginCtx.GetString("lang")
	_, err = runtime.RunString("window.LANG = '" + lang + "'")
	if err != nil {
		fmt.Printf("错误：设置语言时出错 %+v\n", err)
		return "", err
	}

	// 设置用户信息
	userInfo, err := login.GetUserInfo(renderer.ginCtx)
	if err == nil {
		userInfoJSON, _ := json.Marshal(userInfo)
		_, err = runtime.RunString("window.USER_INFO = " + string(userInfoJSON))
		if err != nil {
			return "", err
		}
	}

	// 设置 SSR 标志
	runtime.RunString("window.ssr = true")

	// 执行渲染
	val, err := runtime.RunString("Render()")
	if err != nil {
		return "", err
	}

	html := template.HTML(val.String())
	return html, nil
}

// Instance 实现 gin.HTMLRender 接口的方法
func (t *GojaTemplateRenderer) Instance(name string, data any) render.Render {
	parts := strings.Split(name, ":")
	templateName := parts[0]
	componentName := ""

	if len(parts) > 1 {
		componentName = parts[1]
	}

	return &GojaHTMLRender{
		Template:      t.templates,
		TemplateName:  templateName,
		ComponentName: componentName,
		Data:          data,
		renderer:      t,
	}
}

// GojaHTMLRender 是 HTML 渲染器的自定义实现
type GojaHTMLRender struct {
	Template      *template.Template
	TemplateName  string
	ComponentName string
	Data          any
	renderer      *GojaTemplateRenderer
}

// Render 实现 render.Render 接口
func (r *GojaHTMLRender) Render(w http.ResponseWriter) error {
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
	data.Head = i18n.GetHead(r.renderer.ginContext, strings.ToLower(strings.TrimSuffix(r.ComponentName, ".js")))

	data.Version = conf.Get().GitTag

	// 先执行模板渲染，然后再释放资源
	err = r.Template.ExecuteTemplate(w, r.TemplateName, data)

	r.renderer.Close()

	return err
}

// WriteContentType 设置内容类型
func (r *GojaHTMLRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/html; charset=utf-8"}
	}
}

type ManagedRuntimePool struct {
	pool         chan *goja.Runtime
	maxUsage     int
	usageCounter map[*goja.Runtime]int
	mu           sync.Mutex
}

func NewManagedRuntimePool(size int) *ManagedRuntimePool {
	p := &ManagedRuntimePool{
		pool:         make(chan *goja.Runtime, size),
		maxUsage:     50, // 最大使用次数
		usageCounter: make(map[*goja.Runtime]int),
	}

	// 预先创建运行时实例
	for i := 0; i < size; i++ {
		vm := goja.New()
		p.usageCounter[vm] = 0
		p.pool <- vm
	}

	// 启动内存监控
	go p.monitorMemory()

	return p
}

func (p *ManagedRuntimePool) Get() *goja.Runtime {
	vm := <-p.pool

	p.mu.Lock()
	p.usageCounter[vm]++
	count := p.usageCounter[vm]
	p.mu.Unlock()

	// 如果使用次数超过阈值，创建新实例替换
	if count > p.maxUsage {
		// 创建新实例
		newVM := goja.New()

		p.mu.Lock()
		delete(p.usageCounter, vm)
		p.usageCounter[newVM] = 0
		p.mu.Unlock()

		// 返回新实例
		return newVM
	}

	return vm
}

func (p *ManagedRuntimePool) Put(vm *goja.Runtime) {
	// 清理资源
	cleanupRuntime(vm)

	// 放回池中
	p.pool <- vm
}

func cleanupRuntime(vm *goja.Runtime) {
	// 清除全局对象上的自定义属性
	for _, key := range vm.GlobalObject().Keys() {
		if !isBuiltinGlobalProperty(key) {
			vm.GlobalObject().Delete(key)
		}
	}

	// 清除中断状态
	vm.ClearInterrupt()

	// 强制执行垃圾回收
	runtime.GC()
}

// isBuiltinGlobalProperty 判断属性是否为JavaScript内置全局属性
func isBuiltinGlobalProperty(key string) bool {
	builtins := map[string]bool{
		"Object": true, "Function": true, "Array": true, "String": true,
		"Boolean": true, "Number": true, "Date": true, "RegExp": true,
		"Error": true, "Math": true, "JSON": true, "console": true,
		"parseInt": true, "parseFloat": true, "isNaN": true, "isFinite": true,
		"decodeURI": true, "decodeURIComponent": true,
		"encodeURI": true, "encodeURIComponent": true,
		"eval": true, "Infinity": true, "NaN": true, "undefined": true,
		"globalThis": true, "global": true, "window": true,
	}
	return builtins[key]
}

func (p *ManagedRuntimePool) monitorMemory() {
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// 如果内存使用超过阈值，重建池
		if m.Alloc > memoryThreshold {
			p.rebuildPool()
		}

		time.Sleep(5 * time.Minute)
	}
}

func (p *ManagedRuntimePool) rebuildPool() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 创建新的运行时实例
	newPool := make(chan *goja.Runtime, cap(p.pool))
	newCounter := make(map[*goja.Runtime]int)

	// 清空旧池
	for len(p.pool) > 0 {
		<-p.pool
	}

	// 创建新实例
	for i := 0; i < cap(p.pool); i++ {
		vm := goja.New()
		newCounter[vm] = 0
		newPool <- vm
	}

	// 替换旧池
	p.pool = newPool
	p.usageCounter = newCounter

	// 强制执行垃圾回收
	runtime.GC()
}
