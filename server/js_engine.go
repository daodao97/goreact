package server

import (
	"rogchap.com/v8go"
)

type JsEngine interface {
	RunScript(source string, origin string) (string, error)
	String() string
	Close()
}

func NewV8JsEngine() JsEngine {
	isolate := v8go.NewIsolate()
	global := v8go.NewObjectTemplate(isolate)
	ctx := v8go.NewContext(isolate, global)

	return &v8JsEngine{
		isolate: isolate,
		engine:  ctx,
	}
}

type v8JsEngine struct {
	isolate *v8go.Isolate
	engine  *v8go.Context
	value   *v8go.Value
}

func (e *v8JsEngine) RunScript(source string, origin string) (string, error) {
	val, err := e.engine.RunScript(source, origin)
	if err != nil {
		return "", err
	}
	e.value = val
	return val.String(), nil
}

func (e *v8JsEngine) String() string {
	return e.value.String()
}

func (e *v8JsEngine) Close() {
	e.engine.Close()
	e.isolate.Dispose()
}
