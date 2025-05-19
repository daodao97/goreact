package server

import (
	"os"
	"path/filepath"
	"time"

	"github.com/daodao97/goreact/i18n"
	"github.com/daodao97/xgo/xlog"
	"github.com/gin-gonic/gin"

	"github.com/fsnotify/fsnotify"
)

func setupDev(r *gin.Engine) {
	BuildJS()
	xlog.Debug("HMR init: BuildJS done")

	// 创建一个全局通道用于广播 HMR 事件
	hmrBroadcast := make(chan string, 10)
	xlog.Debug("HMR init: create channel")

	// 监听 frontend 目录, 有变动就重新构建
	go func() {
		xlog.Debug("HMR init: start watch frontend dir")
		frontendDir := filepath.Join(".", "frontend")
		watchDir(frontendDir, func(event fsnotify.Event) {
			BuildJS()
			xlog.Debug("frontend dir changed, prepare send hmr event", xlog.Any("event", event))
			hmrBroadcast <- "hmr"
			xlog.Debug("frontend dir changed, send hmr event")
		})
	}()

	// 监听 locale 目录, 有变动就重新构建
	go func() {
		xlog.Debug("HMR init: start watch locales dir")
		localesDir := filepath.Join(".", "locales")
		watchDir(localesDir, func(event fsnotify.Event) {
			i18n.InitI18n()
			xlog.Debug("locales dir changed, prepare send hmr event", xlog.Any("event", event))
			hmrBroadcast <- "hmr"
			xlog.Debug("locales dir changed, send hmr event")
		})
	}()

	r.GET("/hmr", func(c *gin.Context) {
		xlog.Debug("receive hmr connection request")
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲（如有）

		c.Writer.Flush()

		c.SSEvent("connect", "connected")
		c.Writer.Flush()
		xlog.Debug("hmr connection established, waiting for events...")

		// 为该客户端创建一个专用通道
		clientChan := make(chan string, 1)

		// 启动一个 goroutine 监听广播通道并转发到客户端通道
		go func() {
			for {
				select {
				case msg := <-hmrBroadcast:
					clientChan <- msg
				case <-c.Request.Context().Done():
					return
				}
			}
		}()

		// 主循环接收客户端通道的消息并发送
		for {
			select {
			case msg := <-clientChan:
				xlog.Debug("receive channel event, sending SSEvent: " + msg)
				c.SSEvent("hmr", msg)
				c.Writer.Flush()
			case <-c.Request.Context().Done():
				return
			case <-time.After(30 * time.Second):
				c.SSEvent("ping", "ping")
				c.Writer.Flush()
			}
		}
	})
}

func watchDir(dir string, callback func(event fsnotify.Event)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		xlog.Error("watch dir", xlog.Any("error", err))
		return
	}
	defer watcher.Close()

	// 递归添加监听
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				xlog.Error("watch dir", xlog.Any("path", path), xlog.Any("error", err))
			}
		}
		return nil
	})

	if err != nil {
		xlog.Error("watch dir", xlog.Any("dir", dir), xlog.Any("error", err))
		return
	}

	xlog.Debug("start watch dir", xlog.Any("dir", dir))

	// 防抖动机制
	var lastEventTime time.Time
	debounceInterval := 500 * time.Millisecond

	for event := range watcher.Events {
		if event.Op == fsnotify.Chmod {
			continue
		}
		// 记录所有事件
		xlog.Debug("receive event", xlog.Any("event", event))

		// 检查是否需要触发
		now := time.Now()
		if now.Sub(lastEventTime) > debounceInterval {
			xlog.Debug("trigger callback", xlog.Any("event", event))
			callback(event)
			// 更新上次事件时间
			lastEventTime = now
		} else {
			xlog.Debug("debounce filter event", xlog.Any("event", event))
		}
	}
}
