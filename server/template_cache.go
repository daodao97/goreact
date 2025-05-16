package server

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/daodao97/xgo/xlog"
)

// 缓存文件信息
type cacheFileInfo struct {
	Path       string    // 文件路径
	Size       int64     // 文件大小
	LastAccess time.Time // 最后访问时间
}

// TemplateCache 是模板渲染缓存的管理器
type TemplateCache struct {
	// 缓存目录
	CacheDir string
	// 最大缓存容量（字节）
	MaxCacheSize int64
	// 最大缓存文件数量
	MaxCacheFiles int
	// 缓存清理阈值（当达到最大容量的一定比例时触发清理）
	CleanThreshold float64
	// 缓存清理量（清理掉最旧的一定比例的缓存）
	CleanRatio float64

	// 当前缓存大小
	currentCacheSize int64
	// 缓存文件信息（文件名 -> 访问信息）
	cacheFiles     map[string]*cacheFileInfo
	cacheFilesLock sync.RWMutex
}

// 默认配置
var (
	defaultCacheDir             = filepath.Join(os.TempDir(), "go-react-ssr-cache")
	defaultMaxCacheSize   int64 = 100 * 1024 * 1024 // 100MB
	defaultMaxCacheFiles  int   = 1000
	defaultCleanThreshold       = 0.9 // 90%
	defaultCleanRatio           = 0.2 // 20%
)

// SetCacheDir 设置缓存目录
func SetCacheDir(dir string) {
	defaultCacheDir = dir
}

// SetMaxCacheSize 设置最大缓存大小
func SetMaxCacheSize(size int64) {
	defaultMaxCacheSize = size
}

// SetMaxCacheFiles 设置最大缓存文件数
func SetMaxCacheFiles(count int) {
	defaultMaxCacheFiles = count
}

// SetCleanThreshold 设置清理阈值
func SetCleanThreshold(threshold float64) {
	defaultCleanThreshold = threshold
}

// SetCleanRatio 设置清理比例
func SetCleanRatio(ratio float64) {
	defaultCleanRatio = ratio
}

// NewTemplateCache 创建一个新的模板缓存管理器
func NewTemplateCache() *TemplateCache {
	cache := &TemplateCache{
		CacheDir:       defaultCacheDir,
		MaxCacheSize:   defaultMaxCacheSize,
		MaxCacheFiles:  defaultMaxCacheFiles,
		CleanThreshold: defaultCleanThreshold,
		CleanRatio:     defaultCleanRatio,
		cacheFiles:     make(map[string]*cacheFileInfo),
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(cache.CacheDir, 0755); err != nil {
		xlog.Error("Failed to create cache directory", xlog.Any("error", err))
	}

	// 初始化缓存状态
	cache.initCacheStatus()

	return cache
}

// SetCacheConfig 设置缓存配置
func (c *TemplateCache) SetCacheConfig(maxSize int64, maxFiles int, threshold float64, ratio float64) {
	c.cacheFilesLock.Lock()
	defer c.cacheFilesLock.Unlock()

	c.MaxCacheSize = maxSize
	c.MaxCacheFiles = maxFiles
	c.CleanThreshold = threshold
	c.CleanRatio = ratio

	// 检查是否需要立即清理
	needClean := c.currentCacheSize >= int64(float64(c.MaxCacheSize)*c.CleanThreshold) ||
		len(c.cacheFiles) >= c.MaxCacheFiles

	if needClean {
		go c.cleanCache()
	}
}

// 初始化缓存状态
func (c *TemplateCache) initCacheStatus() {
	// 扫描缓存目录，统计当前缓存使用情况
	entries, err := os.ReadDir(c.CacheDir)
	if err != nil {
		xlog.Error("Failed to read cache directory", xlog.Any("error", err))
		return
	}

	c.cacheFilesLock.Lock()
	defer c.cacheFilesLock.Unlock()

	c.currentCacheSize = 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(c.CacheDir, entry.Name())
		fileSize := info.Size()
		c.currentCacheSize += fileSize

		c.cacheFiles[entry.Name()] = &cacheFileInfo{
			Path:       filePath,
			Size:       fileSize,
			LastAccess: info.ModTime(),
		}
	}

	xlog.Info("Cache initialized",
		xlog.Int("filesCount", len(c.cacheFiles)),
		xlog.Int64("totalSize", c.currentCacheSize),
		xlog.Int64("maxSize", c.MaxCacheSize),
		xlog.Int("maxFiles", c.MaxCacheFiles))
}

// 清理过期缓存
func (c *TemplateCache) cleanCache() {
	c.cacheFilesLock.Lock()
	defer c.cacheFilesLock.Unlock()

	// 检查是否需要清理
	if c.currentCacheSize < int64(float64(c.MaxCacheSize)*c.CleanThreshold) && len(c.cacheFiles) < c.MaxCacheFiles {
		return
	}

	xlog.Info("Starting cache cleanup",
		xlog.Int("currentFiles", len(c.cacheFiles)),
		xlog.Int64("currentSize", c.currentCacheSize))

	// 将文件信息转换为切片，以便于排序
	files := make([]*cacheFileInfo, 0, len(c.cacheFiles))
	for _, info := range c.cacheFiles {
		files = append(files, info)
	}

	// 按最后访问时间排序（最旧的在前面）
	sort.Slice(files, func(i, j int) bool {
		return files[i].LastAccess.Before(files[j].LastAccess)
	})

	// 确定要删除的文件数量
	removeCount := 0
	if len(c.cacheFiles) >= c.MaxCacheFiles {
		// 如果文件数超限，至少删除超出部分+一些额外文件(清理比例)
		overFilesCount := len(c.cacheFiles) - c.MaxCacheFiles
		removeCount = overFilesCount + int(float64(len(c.cacheFiles))*c.CleanRatio)
	} else {
		// 否则根据清理比例删除
		removeCount = int(float64(len(c.cacheFiles)) * c.CleanRatio)
	}

	// 确保至少删除一个文件
	if removeCount < 1 {
		removeCount = 1
	}

	// 限制删除数量不超过文件总数
	if removeCount > len(files) {
		removeCount = len(files)
	}

	// 删除旧文件
	var freedSize int64 = 0
	for i := 0; i < removeCount; i++ {
		fileInfo := files[i]
		// 从文件系统删除
		if err := os.Remove(fileInfo.Path); err != nil {
			xlog.Warn("Failed to remove cache file", xlog.String("path", fileInfo.Path), xlog.Any("error", err))
			continue
		}

		// 找到对应的缓存键
		var keyToRemove string
		for k, v := range c.cacheFiles {
			if v.Path == fileInfo.Path {
				keyToRemove = k
				break
			}
		}

		if keyToRemove != "" {
			freedSize += fileInfo.Size
			delete(c.cacheFiles, keyToRemove)
		}
	}

	c.currentCacheSize -= freedSize
	xlog.Info("Cache cleanup completed",
		xlog.Int("removedFiles", removeCount),
		xlog.Int64("freedSize", freedSize),
		xlog.Int("remainingFiles", len(c.cacheFiles)),
		xlog.Int64("remainingSize", c.currentCacheSize))
}

// GenerateKey 生成缓存键
func (c *TemplateCache) GenerateKey(fragment string, data any) (string, error) {
	key, err := generateCacheKey(fragment, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s_%s", fragment, key), nil
}

// 获取缓存文件路径
func (c *TemplateCache) getCachePath(cacheKey string) string {
	return filepath.Join(c.CacheDir, cacheKey)
}

// Load 从缓存中加载
func (c *TemplateCache) Load(cacheKey string) (template.HTML, bool) {
	cachePath := c.getCachePath(cacheKey)

	// 检查缓存文件是否存在
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return "", false
	}

	// 读取缓存文件
	content, err := os.ReadFile(cachePath)
	if err != nil {
		xlog.Warn("Failed to read cache file", xlog.Any("error", err))
		return "", false
	}

	// 更新文件访问时间
	c.cacheFilesLock.Lock()
	if info, exists := c.cacheFiles[cacheKey]; exists {
		info.LastAccess = time.Now()
	}
	c.cacheFilesLock.Unlock()

	return template.HTML(content), true
}

// Save 保存到缓存
func (c *TemplateCache) Save(cacheKey string, html template.HTML) error {
	// 检查是否需要清理缓存
	if c.currentCacheSize >= int64(float64(c.MaxCacheSize)*c.CleanThreshold) || len(c.cacheFiles) >= c.MaxCacheFiles {
		go c.cleanCache()
	}

	cachePath := c.getCachePath(cacheKey)
	content := []byte(html)
	fileSize := int64(len(content))

	// 写入文件
	if err := os.WriteFile(cachePath, content, 0644); err != nil {
		return err
	}

	// 更新缓存状态
	c.cacheFilesLock.Lock()
	defer c.cacheFilesLock.Unlock()

	c.cacheFiles[cacheKey] = &cacheFileInfo{
		Path:       cachePath,
		Size:       fileSize,
		LastAccess: time.Now(),
	}
	c.currentCacheSize += fileSize

	return nil
}

func generateCacheKey(fragment string, data any) (string, error) {
	hasher := md5.New()
	hasher.Write([]byte(fragment))

	if err := hashAnything(hasher, data); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashAnything(hasher hash.Hash, data any) error {
	switch v := data.(type) {
	case nil:
		hasher.Write([]byte("nil"))
	case string:
		hasher.Write([]byte(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		hasher.Write([]byte(fmt.Sprintf("%v", v)))
	case []any:
		for _, item := range v {
			if err := hashAnything(hasher, item); err != nil {
				return err
			}
		}
	case map[string]any:
		// 收集并排序键
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// 按排序后的键哈希值
		for _, k := range keys {
			hasher.Write([]byte(k))
			if err := hashAnything(hasher, v[k]); err != nil {
				return err
			}
		}
	default:
		// 复杂类型还是用 JSON
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		hasher.Write(data)
	}
	return nil
}
