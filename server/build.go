package server

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daodao97/xgo/xlog"
)

const (
	// 文件权限常量
	DefaultFileMode = 0644

	// 客户端入口模板
	clientTemplateFormat = `import { %s } from "@/pages/%s";
import { renderPage } from "@/core/lib/PageWrapper";

renderPage({Component: %s});
`

	// 服务端入口模板
	serverTemplateFormat = `import { %s } from "@/pages/%s";
import { createServerRenderer } from "@/core/lib/ServerRender";

globalThis.Render = createServerRenderer({ Component: %s });
`
)

// BuildConfig 构建配置结构体
type BuildConfig struct {
	FrontendDir    string
	TmpFrontendDir string
	ClientEntry    string
	ServerEntry    string
	PagesDir       string
	BuildDir       string
	BuildServerDir string
}

// CacheManager 缓存管理结构体
type CacheManager struct {
	config *BuildConfig
}

var globalConfig *BuildConfig

func init() {
	globalConfig = initBuildConfig()
}

// initBuildConfig 初始化构建配置
func initBuildConfig() *BuildConfig {
	pwd, _ := os.Getwd()
	projectName := filepath.Base(pwd)

	tmpFrontendDir := filepath.Join(os.TempDir(), projectName+"-frontend")

	return &BuildConfig{
		FrontendDir:    filepath.Join(pwd, "frontend"),
		TmpFrontendDir: tmpFrontendDir,
		ClientEntry:    filepath.Join(tmpFrontendDir, "app"),
		ServerEntry:    filepath.Join(tmpFrontendDir, "server"),
		PagesDir:       filepath.Join(tmpFrontendDir, "pages"),
		BuildDir:       filepath.Join(pwd, "build"),
		BuildServerDir: filepath.Join(pwd, "build/server"),
	}
}

// BuildJS 构建 JavaScript 文件
func BuildJS() error {
	return BuildJSWithForce(false)
}

// BuildJSWithForce 构建 JavaScript 文件（可强制构建）
func BuildJSWithForce(force bool) error {
	builder := NewJSBuilder(globalConfig)
	return builder.Build(force)
}

// JSBuilder JavaScript 构建器
type JSBuilder struct {
	config       *BuildConfig
	cacheManager *CacheManager
}

// NewJSBuilder 创建新的 JavaScript 构建器
func NewJSBuilder(config *BuildConfig) *JSBuilder {
	return &JSBuilder{
		config:       config,
		cacheManager: &CacheManager{config: config},
	}
}

// Build 执行构建过程
func (b *JSBuilder) Build(force bool) error {
	// 检查缓存
	shouldBuild, clearCaches, err := b.checkShouldBuild(force)
	if err != nil {
		return err
	}

	if !shouldBuild {
		xlog.Debug("no changes detected, skipping build")
		return nil
	}

	// 执行构建
	if err := b.executeBuild(); err != nil {
		b.cleanupOnError(clearCaches)
		return err
	}

	// 更新缓存
	return b.updateCaches()
}

// checkShouldBuild 检查是否需要构建
func (b *JSBuilder) checkShouldBuild(force bool) (bool, []func(), error) {
	var clearCaches []func()

	if force {
		xlog.Debug("force build requested")
		// 获取清理函数但跳过检查
		_, clearDirCache, _ := isDirChanged(b.config.FrontendDir)
		_, clearFileCache, _ := isFileChanged("package.json", "package-lock.json")
		clearCaches = append(clearCaches, clearDirCache, clearFileCache)
		return true, clearCaches, nil
	}

	// 检查前端目录变化
	frontendChanged, clearDirCache, err := isDirChanged(b.config.FrontendDir)
	if err != nil {
		return false, nil, err
	}
	clearCaches = append(clearCaches, clearDirCache)

	// 检查包文件变化
	packageChanged, clearFileCache, err := isFileChanged("package.json", "package-lock.json")
	if err != nil {
		return false, nil, err
	}
	clearCaches = append(clearCaches, clearFileCache)

	return frontendChanged || packageChanged, clearCaches, nil
}

// executeBuild 执行构建步骤
func (b *JSBuilder) executeBuild() error {
	xlog.Debug("build js start")

	// 清理构建目录
	if err := b.cleanBuildDirs(); err != nil {
		return err
	}

	// 安装依赖
	if err := b.installDependencies(); err != nil {
		return err
	}

	// 准备前端文件
	if err := b.prepareFrontendFiles(); err != nil {
		return err
	}

	// 构建 CSS
	if err := b.buildCSS(); err != nil {
		return err
	}

	// 构建 JS
	return b.buildJS()
}

// cleanBuildDirs 清理构建目录
func (b *JSBuilder) cleanBuildDirs() error {
	os.RemoveAll(b.config.BuildDir)
	os.RemoveAll(b.config.TmpFrontendDir)
	return nil
}

// installDependencies 安装依赖
func (b *JSBuilder) installDependencies() error {
	packageChanged, _, err := isFileChanged("package.json", "package-lock.json")
	if err != nil {
		return err
	}

	if !packageChanged {
		return nil
	}

	cmd := exec.Command("npm", "install")
	cmd.Dir = "./"
	xlog.Debug("install dependencies", xlog.String("cmd", cmd.String()))

	output, err := cmd.CombinedOutput()
	if err != nil {
		xlog.Error("install dependencies failed",
			xlog.String("err", err.Error()),
			xlog.String("output", string(output)))
		return fmt.Errorf("npm install failed: %w", err)
	}

	return nil
}

// prepareFrontendFiles 准备前端文件
func (b *JSBuilder) prepareFrontendFiles() error {
	return os.CopyFS(b.config.TmpFrontendDir, os.DirFS(b.config.FrontendDir))
}

// buildCSS 构建 CSS
func (b *JSBuilder) buildCSS() error {
	inputCSS := filepath.Join(b.config.FrontendDir, "css/tailwind-input.css")
	outputCSS := filepath.Join(b.config.TmpFrontendDir, "css/tailwind.css")
	return BuildCSS(inputCSS, outputCSS)
}

// buildJS 构建 JavaScript
func (b *JSBuilder) buildJS() error {
	// 确保目录存在
	err := ensureDirectories(b.config.ClientEntry, b.config.ServerEntry)
	if err != nil {
		return err
	}

	xlog.Debug("BuildJS: generate entry files")
	// 生成入口文件
	err = generateEntryFiles(b.config.PagesDir, b.config.ClientEntry, b.config.ServerEntry)
	if err != nil {
		return err
	}

	aliases := map[string]string{
		"@": b.config.FrontendDir,
	}

	err = BuildClientComponents(b.config.ClientEntry, b.config.BuildDir, aliases, b.config.TmpFrontendDir)
	if err != nil {
		return err
	}

	_, err = BuildServerComponents(b.config.ServerEntry, b.config.BuildServerDir, aliases)
	if err != nil {
		return err
	}

	// copy frontend/public to build/public
	err = copyDir(filepath.Join(b.config.FrontendDir, "public"), b.config.BuildDir)
	if err != nil {
		return err
	}

	xlog.Debug("BuildJS: build done")
	return nil
}

// cleanupOnError 错误时清理
func (b *JSBuilder) cleanupOnError(clearCaches []func()) {
	xlog.Error("Build failed, cleaning up")
	os.RemoveAll(b.config.TmpFrontendDir)

	for _, clearCache := range clearCaches {
		if clearCache != nil {
			clearCache()
		}
	}
}

// updateCaches 更新缓存
func (b *JSBuilder) updateCaches() error {
	return b.cacheManager.UpdateAllCaches()
}

// UpdateAllCaches 更新所有缓存
func (cm *CacheManager) UpdateAllCaches() error {
	// 更新目录缓存
	if err := cm.updateDirCache(); err != nil {
		xlog.Debug("更新目录缓存失败", xlog.String("error", err.Error()))
	}

	// 更新文件缓存
	if err := cm.updateFileCache(); err != nil {
		xlog.Debug("更新文件缓存失败", xlog.String("error", err.Error()))
	}

	return nil
}

// updateDirCache 更新目录缓存
func (cm *CacheManager) updateDirCache() error {
	currentHash, err := calculateDirHash(cm.config.FrontendDir)
	if err != nil {
		return err
	}

	cacheFile := getCacheFilePath(cm.config.FrontendDir)
	if err := writeCachedHash(cacheFile, currentHash); err != nil {
		return err
	}

	xlog.Debug("更新目录缓存成功", xlog.String("hash", currentHash[:8]))
	return nil
}

// updateFileCache 更新文件缓存
func (cm *CacheManager) updateFileCache() error {
	currentHash, err := calculateFilesHash("package.json", "package-lock.json")
	if err != nil {
		return err
	}

	cacheFile := getFilesCacheFilePath("package.json", "package-lock.json")
	if err := writeCachedHash(cacheFile, currentHash); err != nil {
		return err
	}

	xlog.Debug("更新文件缓存成功", xlog.String("hash", currentHash[:8]))
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

// EntryFileGenerator 入口文件生成器
type EntryFileGenerator struct {
	pagesDir    string
	clientEntry string
	serverEntry string
}

// NewEntryFileGenerator 创建入口文件生成器
func NewEntryFileGenerator(pagesDir, clientEntry, serverEntry string) *EntryFileGenerator {
	return &EntryFileGenerator{
		pagesDir:    pagesDir,
		clientEntry: clientEntry,
		serverEntry: serverEntry,
	}
}

// Generate 生成入口文件
func (g *EntryFileGenerator) Generate() error {
	pageFiles, err := g.getComponentFiles()
	if err != nil {
		return err
	}

	for _, file := range pageFiles {
		if err := g.generateEntryFile(file); err != nil {
			return err
		}
	}

	return nil
}

// generateEntryFile 为单个组件生成入口文件
func (g *EntryFileGenerator) generateEntryFile(file string) error {
	baseName := filepath.Base(file)
	componentName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// 生成客户端入口
	if err := g.writeClientEntry(baseName, componentName); err != nil {
		return err
	}

	// 生成服务端入口
	return g.writeServerEntry(baseName, componentName)
}

// writeClientEntry 写入客户端入口文件
func (g *EntryFileGenerator) writeClientEntry(baseName, componentName string) error {
	content := fmt.Sprintf(clientTemplateFormat, componentName, componentName, componentName)
	clientPath := filepath.Join(g.clientEntry, baseName)

	if err := os.WriteFile(clientPath, []byte(content), DefaultFileMode); err != nil {
		return fmt.Errorf("写入客户端入口 %s 失败: %w", clientPath, err)
	}
	return nil
}

// writeServerEntry 写入服务端入口文件
func (g *EntryFileGenerator) writeServerEntry(baseName, componentName string) error {
	content := fmt.Sprintf(serverTemplateFormat, componentName, componentName, componentName)
	serverPath := filepath.Join(g.serverEntry, baseName)

	if err := os.WriteFile(serverPath, []byte(content), DefaultFileMode); err != nil {
		return fmt.Errorf("写入服务端入口 %s 失败: %w", serverPath, err)
	}
	return nil
}

// getComponentFiles 获取组件文件列表
func (g *EntryFileGenerator) getComponentFiles() ([]string, error) {
	return getComponentFiles(g.pagesDir)
}

// generateEntryFiles 生成入口文件（保持向后兼容）
func generateEntryFiles(pagesDir string, clientEntry string, serverEntry string) error {
	generator := NewEntryFileGenerator(pagesDir, clientEntry, serverEntry)
	return generator.Generate()
}

// ComponentScanner 组件扫描器
type ComponentScanner struct {
	rootDir string
}

// NewComponentScanner 创建组件扫描器
func NewComponentScanner(rootDir string) *ComponentScanner {
	return &ComponentScanner{rootDir: rootDir}
}

// Scan 扫描组件文件
func (s *ComponentScanner) Scan() ([]string, error) {
	var files []string

	err := filepath.WalkDir(s.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if s.isComponentFile(path, d) {
			relPath, err := filepath.Rel(s.rootDir, path)
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

// isComponentFile 判断是否为组件文件
func (s *ComponentScanner) isComponentFile(path string, d fs.DirEntry) bool {
	if d.IsDir() {
		return false
	}

	return strings.HasSuffix(path, ".jsx") || strings.HasSuffix(path, ".tsx")
}

// getComponentFiles 获取组件文件列表（保持向后兼容）
func getComponentFiles(componentsDir string) ([]string, error) {
	scanner := NewComponentScanner(componentsDir)
	return scanner.Scan()
}
