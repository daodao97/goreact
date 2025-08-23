package server

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

var frontendDir = ""
var tmpFrontendDir string
var clientEntry string
var serverEntry string
var pagesDir string
var buildDir = ""
var buildServerDir = ""

func init() {
	pwd, _ := os.Getwd()

	projectName := filepath.Base(pwd)
	frontendDir = filepath.Join(pwd, "frontend")
	tmpFrontendDir = filepath.Join(os.TempDir(), projectName+"-frontend")
	clientEntry = filepath.Join(tmpFrontendDir, "app")
	serverEntry = filepath.Join(tmpFrontendDir, "server")
	pagesDir = filepath.Join(tmpFrontendDir, "pages")
	buildDir = filepath.Join(pwd, "build")
	buildServerDir = filepath.Join(pwd, "build/server")
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
	os.CopyFS(tmpFrontendDir, os.DirFS(frontendDir))

	BuildCSS(filepath.Join(frontendDir, "css/tailwind-input.css"), filepath.Join(tmpFrontendDir, "css/tailwind.css"))

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

func BuildCSS(inputCSS, outputCSS string) error {
	cmd := exec.Command("npx", "@tailwindcss/cli", "-i", inputCSS, "-o", outputCSS, "--postcss")
	cmd.Dir = "./"
	xlog.Debug("build css", xlog.String("cmd", cmd.String()))
	output, err := cmd.CombinedOutput()
	if err != nil {
		xlog.Error("build css", xlog.String("err", err.Error()), xlog.String("output", string(output)))
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
	err = generateEntryFiles(pagesDir, clientEntry, serverEntry)
	if err != nil {
		return err
	}

	aliases := map[string]string{
		"@": frontendDir,
	}

	err = BuildClientComponents(clientEntry, buildDir, aliases, tmpFrontendDir)
	if err != nil {
		return err
	}

	_, err = BuildServerComponents(serverEntry, buildServerDir, aliases)
	if err != nil {
		return err
	}

	// copy frontend/public to build/public
	err = copyDir(filepath.Join(frontendDir, "public"), buildDir)
	if err != nil {
		return err
	}

	xlog.Debug("BuildJS: build done")
	return nil
}

// 生成入口文件
func generateEntryFiles(pagesDir string, clientEntry string, serverEntry string) error {
	// 扫描组件目录
	pageFiles, err := getComponentFiles(pagesDir)
	if err != nil {
		return err
	}

	// 为每个组件生成入口文件
	for _, file := range pageFiles {
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
