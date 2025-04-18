package util

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func GetFiles(root, extension string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Check if the file has a .jsx extension
		if !d.IsDir() && filepath.Ext(path) == extension {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// 定义大文件阈值（字节）
const largeFileSizeThreshold = 1024 * 1024 // 1MB

// 全局文件内容缓存
type cacheEntry struct {
	data        map[string]string    // 文件内容
	timestamp   time.Time            // 缓存时间
	fileModTime map[string]time.Time // 文件修改时间
}

var (
	fileCache      = make(map[string]cacheEntry)
	fileCacheMutex sync.RWMutex
	maxCacheSize   = 200 // 最大缓存条目数量
)

// 清理老旧缓存
func cleanCache() {
	if len(fileCache) <= maxCacheSize {
		return
	}

	// 如果缓存超过限制，移除最旧的条目
	var entries []struct {
		key       string
		timestamp time.Time
	}

	for k, v := range fileCache {
		entries = append(entries, struct {
			key       string
			timestamp time.Time
		}{k, v.timestamp})
	}

	// 按时间排序（冒泡排序，简单实现）
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].timestamp.After(entries[j].timestamp) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// 移除最旧的条目直到达到限制
	removeCount := len(entries) - maxCacheSize
	for i := 0; i < removeCount; i++ {
		delete(fileCache, entries[i].key)
	}
}

// 检查文件是否已修改
func hasFileChanged(path string, cachedModTime time.Time) (bool, time.Time) {
	info, err := os.Stat(path)
	if err != nil {
		// 如果无法获取文件信息，保守起见认为文件已修改
		return true, time.Time{}
	}

	newModTime := info.ModTime()
	// 如果文件修改时间比缓存的修改时间更新，则文件已修改
	return newModTime.After(cachedModTime), newModTime
}

func GetFileContent(dir string, suffix string) (map[string]string, error) {
	// 生成缓存键
	cacheKey := dir + ":" + suffix

	// 先尝试从缓存获取
	fileCacheMutex.RLock()
	if entry, exists := fileCache[cacheKey]; exists {
		// 检查文件是否被修改
		filesChanged := false

		// 遍历目录，先检查缓存的文件是否有修改
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && filepath.Ext(path) == suffix {
				componentName := filepath.Base(path)
				if cachedModTime, ok := entry.fileModTime[componentName]; ok {
					changed, _ := hasFileChanged(path, cachedModTime)
					if changed {
						filesChanged = true
						return filepath.SkipDir // 如果找到任何修改，停止遍历
					}
				} else {
					// 缓存中没有这个文件的修改时间信息，说明是新文件
					filesChanged = true
					return filepath.SkipDir
				}
			}
			return nil
		})

		// 如果没有错误且文件没有变化，直接返回缓存内容
		if err == nil && !filesChanged {
			result := entry.data
			fileCacheMutex.RUnlock()
			return result, nil
		}
	}
	fileCacheMutex.RUnlock()

	// 缓存未命中或文件已修改，重新加载文件内容
	fileContents := make(map[string]string)
	fileModTimes := make(map[string]time.Time)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == suffix {
			// 使用完整文件名作为键
			componentName := filepath.Base(path)

			// 获取文件信息
			info, err := os.Stat(path)
			if err != nil {
				return err
			}

			// 记录文件修改时间
			fileModTimes[componentName] = info.ModTime()

			// 对大文件特殊处理
			if info.Size() > largeFileSizeThreshold {
				// 对于大文件，只读取开头部分或者分块处理
				// 这里简化处理，实际中可能需要更复杂的逻辑
				content, err := os.ReadFile(path) // 实际项目中可能需要替换为分块读取
				if err != nil {
					return err
				}
				fileContents[componentName] = string(content)
			} else {
				// 普通文件正常读取
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				fileContents[componentName] = string(content)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 更新缓存
	fileCacheMutex.Lock()
	fileCache[cacheKey] = cacheEntry{
		data:        fileContents,
		timestamp:   time.Now(),
		fileModTime: fileModTimes,
	}
	// 清理老旧缓存
	cleanCache()
	fileCacheMutex.Unlock()

	return fileContents, nil
}
