package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daodao97/goreact/i18n"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xutil"
	"github.com/gin-gonic/gin"

	"github.com/fsnotify/fsnotify"
)

var clientIDCounter int64

// HMR广播器，支持多客户端和事件节流
type HMRBroadcaster struct {
	clients          map[string]chan string
	mutex            sync.RWMutex
	throttle         *time.Timer
	lastEvent        time.Time
	throttleDuration time.Duration
}

func NewHMRBroadcaster() *HMRBroadcaster {
	broadcaster := &HMRBroadcaster{
		clients:          make(map[string]chan string),
		throttleDuration: 300 * time.Millisecond, // 300ms 节流
	}
	xlog.Debug("HMR broadcaster created")
	return broadcaster
}

// 注册客户端
func (h *HMRBroadcaster) RegisterClient(clientID string) chan string {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	clientChan := make(chan string, 5)
	h.clients[clientID] = clientChan
	xlog.Debug("HMR client registered", xlog.String("clientID", clientID), xlog.Int("totalClients", len(h.clients)))
	return clientChan
}

// 注销客户端
func (h *HMRBroadcaster) UnregisterClient(clientID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clientChan, exists := h.clients[clientID]; exists {
		close(clientChan)
		delete(h.clients, clientID)
		xlog.Debug("HMR client unregistered", xlog.String("clientID", clientID), xlog.Int("totalClients", len(h.clients)))
	}
}

// 广播事件到所有客户端，带节流功能
func (h *HMRBroadcaster) Broadcast(event string) {
	h.mutex.Lock()
	now := time.Now()

	// 如果距离上次事件时间太短，启动或重置定时器
	if now.Sub(h.lastEvent) < h.throttleDuration {
		if h.throttle != nil {
			h.throttle.Stop()
		}
		h.throttle = time.AfterFunc(h.throttleDuration, func() {
			h.doBroadcast(event)
		})
		h.mutex.Unlock()
		xlog.Debug("HMR event throttled", xlog.String("event", event))
		return
	}

	// 立即执行广播
	h.lastEvent = now
	h.mutex.Unlock()
	h.doBroadcast(event)
}

// 执行实际的广播
func (h *HMRBroadcaster) doBroadcast(event string) {
	h.mutex.RLock()
	clientCount := len(h.clients)
	h.mutex.RUnlock()

	if clientCount == 0 {
		xlog.Debug("No HMR clients to broadcast to")
		return
	}

	xlog.Debug("Broadcasting HMR event", xlog.String("event", event), xlog.Int("clients", clientCount))

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for clientID, clientChan := range h.clients {
		select {
		case clientChan <- event:
			xlog.Debug("HMR event sent to client", xlog.String("clientID", clientID))
		default:
			xlog.Warn("HMR client channel full, skipping", xlog.String("clientID", clientID))
		}
	}
}

func setupDev(r *gin.Engine) {
	BuildJS()
	// 创建 HMR 广播器
	hmrBroadcaster := NewHMRBroadcaster()

	// 监听 frontend 目录, 有变动就重新构建
	xutil.Go(context.Background(), func() {
		frontendDir := filepath.Join(".", "frontend")
		xlog.Debug("HMR init: start watch frontend dir", xlog.String("dir", frontendDir))
		watchDir(frontendDir, func(event fsnotify.Event) {
			BuildJS()
			xlog.Debug("frontend dir changed, broadcasting hmr event", xlog.Any("event", event))
			hmrBroadcaster.Broadcast("hmr")
		})
	})

	// 监听 locale 目录, 有变动就重新构建
	xutil.Go(context.Background(), func() {
		localesDir := filepath.Join(".", "locales")
		xlog.Debug("HMR init: start watch locales dir", xlog.String("dir", localesDir))
		watchDir(localesDir, func(event fsnotify.Event) {
			i18n.InitI18n()
			xlog.Debug("locales dir changed, broadcasting hmr event", xlog.Any("event", event))
			hmrBroadcaster.Broadcast("hmr")
		})
	})

	xutil.Go(context.Background(), func() {
		packageJson := filepath.Join(".", "package.json")
		xlog.Debug("HMR init: start watch package.json", xlog.String("file", packageJson))
		watchFileContentChange([]string{packageJson}, func(changedFiles []string) {
			BuildJS()
			xlog.Debug("package.json changed, broadcasting hmr event", xlog.Any("changedFiles", changedFiles))
			hmrBroadcaster.Broadcast("hmr")
		})
	})

	r.GET("/hmr", func(c *gin.Context) {
		xlog.Debug("receive hmr connection request")
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲（如有）
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

		c.Writer.Flush()

		// 生成客户端ID
		clientID := fmt.Sprintf("hmr-client-%d", atomic.AddInt64(&clientIDCounter, 1))

		// 注册客户端到广播器
		clientChan := hmrBroadcaster.RegisterClient(clientID)
		defer hmrBroadcaster.UnregisterClient(clientID)

		// 发送连接确认
		c.SSEvent("connect", "connected")
		c.Writer.Flush()
		xlog.Debug("hmr connection established, waiting for events...", xlog.String("clientID", clientID))

		// 主循环接收客户端通道的消息并发送
		for {
			select {
			case msg, ok := <-clientChan:
				if !ok {
					xlog.Debug("client channel closed", xlog.String("clientID", clientID))
					return
				}
				xlog.Debug("receive channel event, sending SSEvent", xlog.String("msg", msg), xlog.String("clientID", clientID))
				c.SSEvent("hmr", msg)
				c.Writer.Flush()
			case <-c.Request.Context().Done():
				xlog.Debug("client connection closed", xlog.String("clientID", clientID))
				return
			case <-time.After(30 * time.Second):
				xlog.Debug("sending heartbeat", xlog.String("clientID", clientID))
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

// 检查事件是否与目标文件相关
func isFileEvent(eventPath, targetPath string) bool {
	// 清理路径
	eventPath = filepath.Clean(eventPath)
	targetPath = filepath.Clean(targetPath)

	// 直接匹配
	if eventPath == targetPath {
		return true
	}

	// 检查是否是目标文件的父目录事件（用于监听不存在的文件）
	targetDir := filepath.Dir(targetPath)
	targetFileName := filepath.Base(targetPath)
	eventDir := filepath.Dir(eventPath)
	eventFileName := filepath.Base(eventPath)

	return eventDir == targetDir && eventFileName == targetFileName
}

// 智能文件监控：只在文件内容真正变化时触发回调
func watchFileContentChange(filePaths []string, callback func(changedFiles []string)) {
	if len(filePaths) == 0 {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		xlog.Error("create smart file watcher", xlog.Any("error", err))
		return
	}
	defer watcher.Close()

	// 跟踪已添加到监听的目录，避免重复添加
	watchedDirs := make(map[string]bool)

	// 为每个文件添加监听
	for _, filePath := range filePaths {
		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			xlog.Debug("file does not exist, watching parent directory", xlog.String("file", filePath))
			// 如果文件不存在，监听其父目录
			parentDir := filepath.Dir(filePath)
			if !watchedDirs[parentDir] {
				err = watcher.Add(parentDir)
				if err != nil {
					xlog.Error("watch file parent dir", xlog.String("dir", parentDir), xlog.Any("error", err))
					continue
				}
				watchedDirs[parentDir] = true
			}
		} else {
			// 文件存在，监听其父目录（更稳定的监听方式）
			parentDir := filepath.Dir(filePath)
			if !watchedDirs[parentDir] {
				err = watcher.Add(parentDir)
				if err != nil {
					xlog.Error("watch file parent dir", xlog.String("dir", parentDir), xlog.Any("error", err))
					continue
				}
				watchedDirs[parentDir] = true
			}
		}
	}

	xlog.Debug("start smart watch files", xlog.Any("files", filePaths), xlog.Any("dirs", getKeys(watchedDirs)))

	// 防抖动机制
	var lastEventTime time.Time
	debounceInterval := 1000 * time.Millisecond // 稍长的防抖时间，因为要进行内容检查

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				xlog.Debug("smart watcher events channel closed")
				return
			}

			// 跳过权限变更事件
			if event.Op == fsnotify.Chmod {
				continue
			}

			// 检查事件是否与任何目标文件相关
			relatedFiles := getRelatedFiles(event.Name, filePaths)
			if len(relatedFiles) == 0 {
				continue
			}

			xlog.Debug("receive related file event", xlog.Any("event", event), xlog.Any("relatedFiles", relatedFiles))

			// 防抖动检查
			now := time.Now()
			if now.Sub(lastEventTime) > debounceInterval {
				// 检查文件内容是否真正变化
				changed, _, err := isFileChanged(relatedFiles...)
				if err != nil {
					xlog.Error("check file content change", xlog.Any("files", relatedFiles), xlog.Any("error", err))
					continue
				}

				if changed {
					xlog.Debug("files content actually changed, trigger callback", xlog.Any("files", relatedFiles))
					callback(relatedFiles)
				} else {
					xlog.Debug("files content not changed, skip callback", xlog.Any("files", relatedFiles))
				}
				lastEventTime = now
			} else {
				xlog.Debug("debounce filter smart file event", xlog.Any("event", event))
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				xlog.Debug("smart watcher errors channel closed")
				return
			}
			xlog.Error("smart file watcher error", xlog.Any("error", err))
		}
	}
}

// 获取与事件相关的文件列表
func getRelatedFiles(eventPath string, targetPaths []string) []string {
	var related []string
	for _, targetPath := range targetPaths {
		if isFileEvent(eventPath, targetPath) {
			related = append(related, targetPath)
		}
	}
	return related
}

// 获取map的所有key
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// 监控多个文件，支持按组分别处理
func watchFileGroups(fileGroups map[string][]string, callbacks map[string]func(changedFiles []string)) {
	for groupName, files := range fileGroups {
		if callback, exists := callbacks[groupName]; exists {
			xutil.Go(context.Background(), func() {
				xlog.Debug("start watching file group", xlog.String("group", groupName), xlog.Any("files", files))
				watchFileContentChange(files, func(changedFiles []string) {
					xlog.Debug("file group changed", xlog.String("group", groupName), xlog.Any("changedFiles", changedFiles))
					callback(changedFiles)
				})
			})
		}
	}
}

// 使用智能文件监控的示例：
//
// func setupDevWithSmartWatch(r *gin.Engine) {
// 	BuildJS()
// 	hmrBroadcast := make(chan string, 10)
//
// 	// 定义文件组和对应的处理函数
// 	fileGroups := map[string][]string{
// 		"package": {"package.json", "package-lock.json"},
// 		"config":  {"tsconfig.json", "tailwind.config.js"},
// 		"env":     {".env", ".env.local"},
// 	}
//
// 	callbacks := map[string]func(changedFiles []string){
// 		"package": func(changedFiles []string) {
// 			xlog.Debug("package files changed, rebuilding...", xlog.Any("files", changedFiles))
// 			BuildJS()
// 			hmrBroadcast <- "hmr"
// 		},
// 		"config": func(changedFiles []string) {
// 			xlog.Debug("config files changed, rebuilding...", xlog.Any("files", changedFiles))
// 			BuildJS()
// 			hmrBroadcast <- "hmr"
// 		},
// 		"env": func(changedFiles []string) {
// 			xlog.Debug("environment files changed", xlog.Any("files", changedFiles))
// 			hmrBroadcast <- "hmr"
// 		},
// 	}
//
// 	// 启动智能文件组监控
// 	watchFileGroups(fileGroups, callbacks)
//
// 	// 继续原有的目录监控...
// }
