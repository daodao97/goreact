package server

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/daodao97/xgo/xlog"
)

// 客户端入口模板
const clientTemplateFormat = `import { %s } from "@/pages/%s";
import { renderPage } from "@/core/lib/PageWrapper";

renderPage({Component: %s});
`

// 服务端入口模板
const serverTemplateFormat = `import { %s } from "@/pages/%s";
import { createServerRenderer } from "@/core/lib/ServerRender";

globalThis.Render = createServerRenderer({ Component: %s });
`

var frontendDir = "./frontend"
var tmpFrontendDir string
var clientEntry string
var serverEntry string
var pagesDir string
var pwd string
var buildDir = "./build"
var buildServerDir = "./build/server"

func init() {
	pwd, _ = os.Getwd()
	frontendDir = filepath.Join(pwd, "frontend")

	projectName := filepath.Base(pwd)

	tmpFrontendDir = filepath.Join(os.TempDir(), projectName+"-frontend")
	clientEntry = filepath.Join(tmpFrontendDir, "app")
	serverEntry = filepath.Join(tmpFrontendDir, "server")
	pagesDir = filepath.Join(tmpFrontendDir, "pages")
}

func BuildJS() error {
	return BuildJSWithForce(false)
}

func BuildJSWithForce(force bool) error {
	var frontendDirChanged, isPackageChanged bool
	var clearDirCache, clearFileCache func()
	var err error

	if !force {
		frontendDirChanged, clearDirCache, err = isDirChanged(frontendDir)
		if err != nil {
			return err
		}

		isPackageChanged, clearFileCache, err = isFileChanged("package.json", "package-lock.json")
		if err != nil {
			return err
		}

		if !frontendDirChanged && !isPackageChanged {
			xlog.Debug("frontend dir is not changed and package is not changed, skip build")
			return nil
		}
	} else {
		xlog.Debug("force build requested")
		// 在强制构建时，我们仍然需要获取清理函数，但跳过变更检查
		_, clearDirCache, err = isDirChanged(frontendDir)
		if err != nil {
			return err
		}
		_, clearFileCache, err = isFileChanged("package.json", "package-lock.json")
		if err != nil {
			return err
		}
		frontendDirChanged = true
		isPackageChanged = true
	}

	xlog.Debug("build js start")
	os.RemoveAll(buildDir)
	xlog.Debug("remove build dir", xlog.String("dir", buildDir))

	if isPackageChanged {
		cmd := exec.Command("npm", "install")
		cmd.Dir = "./"
		xlog.Debug("install dependencies", xlog.String("cmd", cmd.String()))
		_, err := cmd.CombinedOutput()
		if err != nil {
			xlog.Debug("install dependencies", xlog.String("err", err.Error()))
			log.Fatal(err)
		}
	}

	os.RemoveAll(tmpFrontendDir)
	xlog.Debug("remove tmpFrontendDir dir", xlog.String("dir", tmpFrontendDir))
	xlog.Debug("copy frontend dir", xlog.String("dir", tmpFrontendDir), xlog.String("to", frontendDir))
	os.CopyFS(tmpFrontendDir, os.DirFS(frontendDir))

	BuildCSS()

	if err := buildJS(); err != nil {
		xlog.Error("Build error", xlog.Err(err))
		// clear build cache
		os.RemoveAll(tmpFrontendDir)
		// clear hash cache files to allow rebuild
		if clearDirCache != nil {
			clearDirCache()
		}
		if clearFileCache != nil {
			clearFileCache()
		}
		return err
	}

	// 构建成功，更新缓存
	if frontendDirChanged {
		currentHash, err := calculateDirHash(frontendDir)
		if err == nil {
			cacheFile := getCacheFilePath(frontendDir)
			if err := writeCachedHash(cacheFile, currentHash); err != nil {
				xlog.Debug("更新目录缓存失败", xlog.String("error", err.Error()))
			} else {
				xlog.Debug("更新目录缓存成功", xlog.String("hash", currentHash[:8]))
			}
		}
	}

	if isPackageChanged {
		currentHash, err := calculateFilesHash("package.json", "package-lock.json")
		if err == nil {
			cacheFile := getFilesCacheFilePath("package.json", "package-lock.json")
			if err := writeCachedHash(cacheFile, currentHash); err != nil {
				xlog.Debug("更新文件缓存失败", xlog.String("error", err.Error()))
			} else {
				xlog.Debug("更新文件缓存成功", xlog.String("hash", currentHash[:8]))
			}
		}
	}

	return nil
}

func BuildCSS() error {
	xlog.Debug("build css start")
	cmd := exec.Command("npx", "@tailwindcss/cli", "-i", filepath.Join(frontendDir, "css/tailwind-input.css"), "-o", filepath.Join(tmpFrontendDir, "css/tailwind.css"), "--postcss")
	cmd.Dir = "./"
	xlog.Debug("build css", xlog.String("cmd", cmd.String()))
	output, err := cmd.CombinedOutput()
	if err != nil {
		xlog.Debug("build css", xlog.String("err", err.Error()), xlog.String("output", string(output)))
	} else {
		xlog.Debug("build css", xlog.String("output", string(output)))
	}
	return err
}

func buildJS() error {
	// 确保目录存在
	err := ensureDirectories(clientEntry, serverEntry)
	if err != nil {
		return err
	}

	xlog.Debug("BuildJS: generate entry files")
	// 生成入口文件
	err = generateEntryFiles()
	if err != nil {
		return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	xlog.Debug("BuildJS: build client components")
	err = BuildClientComponents(clientEntry, buildDir, map[string]string{
		"@": filepath.Join(currentDir, "frontend"),
		"#": filepath.Join(currentDir, "node_modules", "goreact", "ui"),
	})
	if err != nil {
		return err
	}

	xlog.Debug("BuildJS: build server components")
	_, err = BuildServerComponents(serverEntry, buildServerDir, map[string]string{
		"@": filepath.Join(currentDir, "frontend"),
		"#": filepath.Join(currentDir, "node_modules", "goreact", "ui"),
	})
	if err != nil {
		return err
	}

	// copy frontend/public to build/public
	xlog.Debug("BuildJS: copy frontend/public to build/public")
	err = copyDir(filepath.Join(currentDir, "frontend/public"), buildDir)
	if err != nil {
		return err
	}

	xlog.Debug("BuildJS: build done")
	return nil
}

func copyDir(src string, dest string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// 组合目标路径
		destPath := filepath.Join(dest, relPath)

		if d.IsDir() {
			// 创建目标目录
			return os.MkdirAll(destPath, 0755)
		} else {
			// 确保目标目录存在
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return err
			}

			// 复制文件
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			return os.WriteFile(destPath, data, 0644)
		}
	})
}

// 确保目录存在
func ensureDirectories(dirs ...string) error {
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
			}
		}
	}
	return nil
}

// 生成入口文件
func generateEntryFiles() error {
	// 扫描组件目录
	componentFiles, err := getComponentFiles(pagesDir)
	if err != nil {
		return err
	}

	// 为每个组件生成入口文件
	for _, file := range componentFiles {
		baseName := filepath.Base(file)
		componentName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		// 生成客户端入口
		clientContent := fmt.Sprintf(clientTemplateFormat, componentName, componentName, componentName)
		clientPath := filepath.Join(clientEntry, baseName)
		err := os.WriteFile(clientPath, []byte(clientContent), 0644)
		if err != nil {
			return fmt.Errorf("写入客户端入口 %s 失败: %w", clientPath, err)
		}

		// 生成服务端入口
		serverContent := fmt.Sprintf(serverTemplateFormat, componentName, componentName, componentName)
		serverPath := filepath.Join(serverEntry, baseName)
		err = os.WriteFile(serverPath, []byte(serverContent), 0644)
		if err != nil {
			return fmt.Errorf("写入服务端入口 %s 失败: %w", serverPath, err)
		}
	}

	return nil
}

// 获取组件文件列表
func getComponentFiles(componentsDir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(componentsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 只处理 .jsx 和 .tsx 文件
		if !d.IsDir() && (strings.HasSuffix(path, ".jsx") || strings.HasSuffix(path, ".tsx")) {
			// 只保留文件名，不包含路径
			relPath, err := filepath.Rel(componentsDir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("扫描组件目录失败: %w", err)
	}

	return files, nil
}

func isDirChanged(dir1 string) (bool, func(), error) {
	// 计算目录的hash值
	currentHash, err := calculateDirHash(dir1)
	if err != nil {
		return false, nil, fmt.Errorf("计算目录hash失败: %w", err)
	}

	// 获取缓存文件路径
	cacheFile := getCacheFilePath(dir1)

	// 创建清理函数
	clearCache := func() {
		if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
			xlog.Debug("清除目录缓存文件失败", xlog.String("file", cacheFile), xlog.String("error", err.Error()))
		} else {
			xlog.Debug("清除目录缓存文件", xlog.String("file", cacheFile))
		}
	}

	// 读取之前缓存的hash
	cachedHash, err := readCachedHash(cacheFile)
	if err != nil {
		// 第一次运行或缓存文件不存在，认为有变更
		xlog.Debug("无法读取缓存hash，认为目录有变更", xlog.String("error", err.Error()))
		// 注意：这里不立即写入缓存，等构建成功后再写入
		return true, clearCache, nil
	}

	// 对比hash值
	hasChanged := currentHash != cachedHash
	if hasChanged {
		xlog.Debug("目录有变更", xlog.String("dir", dir1), xlog.String("oldHash", cachedHash[:8]), xlog.String("newHash", currentHash[:8]))
	} else {
		xlog.Debug("目录无变更", xlog.String("dir", dir1), xlog.String("hash", currentHash[:8]))
	}

	return hasChanged, clearCache, nil
}

// 计算目录的hash值
func calculateDirHash(dir string) (string, error) {
	hasher := sha256.New()

	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过某些不需要检查的目录和文件
		if d.IsDir() {
			name := d.Name()
			if slices.Contains([]string{"node_modules", ".git", "dist", "build"}, name) {
				return filepath.SkipDir
			}
			return nil
		}

		// 只检查特定类型的文件
		ext := filepath.Ext(path)
		if slices.Contains([]string{".tsx", ".ts", ".jsx", ".js", ".css", ".scss", ".json", ".html"}, ext) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// 排序确保hash的一致性
	sort.Strings(files)

	for _, file := range files {
		// 获取文件信息
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		// 将文件路径、大小、修改时间写入hasher
		relPath, _ := filepath.Rel(dir, file)
		hasher.Write([]byte(relPath))
		hasher.Write([]byte(fmt.Sprintf("%d", info.Size())))
		hasher.Write([]byte(info.ModTime().Format(time.RFC3339Nano)))

		// 对于小文件，直接读取内容计算hash
		if info.Size() < 1024*1024 { // 小于1MB的文件
			content, err := os.ReadFile(file)
			if err == nil {
				hasher.Write(content)
			}
		}
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// 获取缓存文件路径
func getCacheFilePath(dir string) string {
	// 使用目录路径生成唯一的缓存文件名
	dirHash := sha256.Sum256([]byte(dir))
	fileName := fmt.Sprintf("goreact-dir-cache-%x.txt", dirHash[:8])
	return filepath.Join(os.TempDir(), fileName)
}

// 读取缓存的hash
func readCachedHash(cacheFile string) (string, error) {
	content, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

// 写入缓存的hash
func writeCachedHash(cacheFile string, hash string) error {
	return os.WriteFile(cacheFile, []byte(hash), 0644)
}

func isFileChanged(filePath ...string) (bool, func(), error) {
	if len(filePath) == 0 {
		return false, nil, nil
	}

	// 计算所有文件的联合hash值
	currentHash, err := calculateFilesHash(filePath...)
	if err != nil {
		return false, nil, fmt.Errorf("计算文件hash失败: %w", err)
	}

	// 获取缓存文件路径
	cacheFile := getFilesCacheFilePath(filePath...)

	// 创建清理函数
	clearCache := func() {
		if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
			xlog.Debug("清除文件缓存文件失败", xlog.String("file", cacheFile), xlog.String("error", err.Error()))
		} else {
			xlog.Debug("清除文件缓存文件", xlog.String("file", cacheFile))
		}
	}

	// 读取之前缓存的hash
	cachedHash, err := readCachedHash(cacheFile)
	if err != nil {
		// 第一次运行或缓存文件不存在，认为有变更
		xlog.Debug("无法读取缓存hash，认为文件有变更", xlog.String("error", err.Error()))
		// 注意：这里不立即写入缓存，等构建成功后再写入
		return true, clearCache, nil
	}

	// 对比hash值
	hasChanged := currentHash != cachedHash
	if hasChanged {
		xlog.Debug("文件有变更", xlog.Any("files", filePath), xlog.String("oldHash", cachedHash[:8]), xlog.String("newHash", currentHash[:8]))
	} else {
		xlog.Debug("文件无变更", xlog.Any("files", filePath), xlog.String("hash", currentHash[:8]))
	}

	return hasChanged, clearCache, nil
}

// 计算多个文件的联合hash值
func calculateFilesHash(filePaths ...string) (string, error) {
	hasher := sha256.New()

	// 对文件路径进行排序，确保hash的一致性
	sortedPaths := make([]string, len(filePaths))
	copy(sortedPaths, filePaths)
	sort.Strings(sortedPaths)

	for _, filePath := range sortedPaths {
		// 检查文件是否存在
		_, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				// 文件不存在，将路径和特殊标记写入hasher
				hasher.Write([]byte(filePath))
				hasher.Write([]byte("NOT_EXIST"))
				continue
			}
			return "", fmt.Errorf("获取文件信息失败 %s: %w", filePath, err)
		}

		// 将文件路径写入hasher
		hasher.Write([]byte(filePath))

		// 读取文件内容计算hash（只基于内容，不包含修改时间和大小）
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取文件内容失败 %s: %w", filePath, err)
		}
		hasher.Write(content)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// 获取文件列表的缓存文件路径
func getFilesCacheFilePath(filePaths ...string) string {
	// 将所有文件路径合并生成唯一的缓存文件名
	allPaths := strings.Join(filePaths, "|")
	pathHash := sha256.Sum256([]byte(allPaths))
	fileName := fmt.Sprintf("goreact-files-cache-%x.txt", pathHash[:8])
	return filepath.Join(os.TempDir(), fileName)
}
