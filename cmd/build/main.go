package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/daodao97/goreact/server"
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

var appEntry = "./frontend/app"
var serverEntry = "./frontend/server"

var buildDir = "./build"
var buildServerDir = "./build/server"

var componentDir = "./frontend/pages"

func main() {
	// 确保目录存在
	err := ensureDirectories(appEntry, serverEntry)
	if err != nil {
		log.Fatal(err)
	}

	// 生成入口文件
	err = generateEntryFiles()
	if err != nil {
		log.Fatal(err)
	}

	// 执行原有的构建过程，添加别名配置
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = server.BuildClientComponents(appEntry, buildDir, map[string]string{
		"@": filepath.Join(currentDir, "frontend"),
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = server.BuildServerComponents(serverEntry, buildServerDir, map[string]string{
		"@": filepath.Join(currentDir, "frontend"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// copy frontend/public to build/public
	err = copyDir(filepath.Join(currentDir, "frontend/public"), buildDir)
	if err != nil {
		log.Fatal(err)
	}
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
	componentFiles, err := getComponentFiles(componentDir)
	if err != nil {
		return err
	}

	// 为每个组件生成入口文件
	for _, file := range componentFiles {
		baseName := filepath.Base(file)
		componentName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		// 生成客户端入口
		clientContent := fmt.Sprintf(clientTemplateFormat, componentName, componentName, componentName)
		clientPath := filepath.Join(appEntry, baseName)
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
