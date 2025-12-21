package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daodao97/goreact/util"
	"github.com/daodao97/xgo/xlog"
	esbuild "github.com/evanw/esbuild/pkg/api"
)

const textEncoderPolyfill = `function TextEncoder(){} TextEncoder.prototype.encode=function(string){var octets=[],length=string.length,i=0;while(i<length){var codePoint=string.codePointAt(i),c=0,bits=0;codePoint<=0x7F?(c=0,bits=0x00):codePoint<=0x7FF?(c=6,bits=0xC0):codePoint<=0xFFFF?(c=12,bits=0xE0):codePoint<=0x1FFFFF&&(c=18,bits=0xF0),octets.push(bits|(codePoint>>c)),c-=6;while(c>=0){octets.push(0x80|((codePoint>>c)&0x3F)),c-=6}i+=codePoint>=0x10000?2:1}return octets};function TextDecoder(){} TextDecoder.prototype.decode=function(octets){var string="",i=0;while(i<octets.length){var octet=octets[i],bytesNeeded=0,codePoint=0;octet<=0x7F?(bytesNeeded=0,codePoint=octet&0xFF):octet<=0xDF?(bytesNeeded=1,codePoint=octet&0x1F):octet<=0xEF?(bytesNeeded=2,codePoint=octet&0x0F):octet<=0xF4&&(bytesNeeded=3,codePoint=octet&0x07),octets.length-i-bytesNeeded>0?function(){for(var k=0;k<bytesNeeded;){octet=octets[i+k+1],codePoint=(codePoint<<6)|(octet&0x3F),k+=1}}():codePoint=0xFFFD,bytesNeeded=octets.length-i,string+=String.fromCodePoint(codePoint),i+=bytesNeeded+1}return string};`

const messageChannelPolyfill = `if(typeof MessageChannel==="undefined"){var MessageChannel=function(){this.port1={postMessage:function(msg){setTimeout(()=>{this.onmessage&&this.onmessage({data:msg})},0)},onmessage:null},this.port2={postMessage:function(msg){setTimeout(()=>{this.onmessage&&this.onmessage({data:msg})},0)},onmessage:null}}}`

const processPolyfill = `var process = {env: {NODE_ENV: "production"}};`

// formatEsbuildMessages creates readable errors with file/line/column and path mapping.
func formatEsbuildMessages(msgs []esbuild.Message, pathMap map[string]string) string {
	var sb strings.Builder

	for i, msg := range msgs {
		if i > 0 {
			sb.WriteString("\n")
		}

		filePath := ""
		if msg.Location != nil {
			filePath = msg.Location.File
			for from, to := range pathMap {
				if strings.HasPrefix(filePath, from) {
					filePath = to + strings.TrimPrefix(filePath, from)
					break
				}
			}
			fmt.Fprintf(&sb, "%s:%d:%d: %s", filePath, msg.Location.Line, msg.Location.Column, msg.Text)
		} else {
			sb.WriteString(msg.Text)
		}

		for _, note := range msg.Notes {
			sb.WriteString("\n  ")
			if note.Location != nil {
				noteFile := note.Location.File
				for from, to := range pathMap {
					if strings.HasPrefix(noteFile, from) {
						noteFile = to + strings.TrimPrefix(noteFile, from)
						break
					}
				}
				fmt.Fprintf(&sb, "note %s:%d:%d: %s", noteFile, note.Location.Line, note.Location.Column, note.Text)
			} else {
				fmt.Fprintf(&sb, "note: %s", note.Text)
			}
		}
	}

	return sb.String()
}

func aliasPlugin(aliases map[string]string) esbuild.Plugin {
	return esbuild.Plugin{
		Name: "alias-resolver",
		Setup: func(build esbuild.PluginBuild) {
			for alias, path := range aliases {
				build.OnResolve(esbuild.OnResolveOptions{Filter: "^" + alias + "/"}, func(args esbuild.OnResolveArgs) (esbuild.OnResolveResult, error) {
					// 将 @/ 替换为实际路径
					newPath := strings.TrimPrefix(args.Path, alias+"/")
					basePath := path + "/" + newPath

					// 尝试不同的扩展名
					extensions := []string{"", ".tsx", ".ts", ".jsx", ".js"}
					for _, ext := range extensions {
						fullPath := basePath + ext
						if _, err := os.Stat(fullPath); err == nil {
							return esbuild.OnResolveResult{Path: fullPath, External: false}, nil
						}
					}

					// 如果找不到文件，尝试作为目录查找 index 文件
					for _, ext := range extensions[1:] { // 跳过空扩展名
						fullPath := basePath + "/index" + ext
						if _, err := os.Stat(fullPath); err == nil {
							return esbuild.OnResolveResult{Path: fullPath, External: false}, nil
						}
					}

					fmt.Printf("路径解析失败: %s -> %s\n", args.Path, basePath)
					// 返回原始路径，让 esbuild 继续尝试解析
					return esbuild.OnResolveResult{Path: basePath, External: false}, nil
				})
			}
		},
	}
}

func BuildClientComponents(jsFolder, jsOutput string, aliases map[string]string, tmpFrontendDir string, frontendDir string) error {
	xlog.Debug(fmt.Sprintf("Building client Javascript, jsFolder %s => jsOutput %s", jsFolder, jsOutput))

	filesJSX, err := util.GetFiles(jsFolder, ".jsx")
	if err != nil {
		return err
	}

	filesTSX, err := util.GetFiles(jsFolder, ".tsx")
	if err != nil {
		return err
	}

	allFiles := append(filesJSX, filesTSX...)
	allFiles = append(allFiles, tmpFrontendDir+"/app.js")

	pwd, _ := os.Getwd()

	builds := esbuild.Build(esbuild.BuildOptions{
		EntryPoints:    allFiles,
		Bundle:         true,
		Write:          true,
		Splitting:      true,
		AllowOverwrite: true,
		AssetNames:     "[name]-[hash]",
		Outdir:         jsOutput,
		Format:         esbuild.FormatESModule,
		Platform:       esbuild.PlatformBrowser,
		Target:         esbuild.ESNext,
		Loader: map[string]esbuild.Loader{
			".jsx":  esbuild.LoaderJSX,
			".tsx":  esbuild.LoaderTSX,
			".scss": esbuild.LoaderLocalCSS,
		},
		Plugins:       []esbuild.Plugin{aliasPlugin(aliases)},
		NodePaths:     []string{filepath.Join(pwd, "node_modules")},
		AbsWorkingDir: pwd,
	})

	if len(builds.Errors) > 0 {
		return fmt.Errorf("esbuild errors:\n%s", formatEsbuildMessages(builds.Errors, map[string]string{
			tmpFrontendDir: frontendDir,
		}))
	}

	return nil
}

func BuildServerComponents(jsFolder, jsOutput string, aliases map[string]string, tmpFrontendDir string, frontendDir string) (map[string]string, error) {
	result := map[string]string{}

	filesJSX, err := util.GetFiles(jsFolder, ".jsx")
	if err != nil {
		return result, err
	}

	filesTSX, err := util.GetFiles(jsFolder, ".tsx")
	if err != nil {
		return result, err
	}

	// filesJS, err := GetFiles(jsFolder, ".js")
	// if err != nil {
	// 	return result, err
	// }

	allFiles := append(filesJSX, filesTSX...)

	pwd, _ := os.Getwd()

	builds := esbuild.Build(esbuild.BuildOptions{
		EntryPoints: allFiles,
		Bundle:      true,
		Write:       true,
		Outdir:      jsOutput,
		Format:      esbuild.FormatESModule,
		Platform:    esbuild.PlatformBrowser,
		Target:      esbuild.ESNext,
		Banner: map[string]string{
			"js": processPolyfill + messageChannelPolyfill + textEncoderPolyfill,
		},
		Loader: map[string]esbuild.Loader{
			".jsx":  esbuild.LoaderJSX,
			".tsx":  esbuild.LoaderTSX,
			".scss": esbuild.LoaderLocalCSS,
		},
		Plugins:       []esbuild.Plugin{aliasPlugin(aliases)},
		NodePaths:     []string{filepath.Join(pwd, "node_modules")},
		AbsWorkingDir: pwd,
	})

	if len(builds.Errors) > 0 {
		return result, fmt.Errorf("esbuild errors:\n%s", formatEsbuildMessages(builds.Errors, map[string]string{
			tmpFrontendDir: frontendDir,
		}))
	}

	for _, file := range builds.OutputFiles {
		if strings.Contains(file.Path, jsOutput) {
			paths := strings.Split(file.Path, jsOutput)
			path := ""
			if len(paths) >= 2 {
				path = strings.Join(paths[1:], "")
			}
			result[path] = string(file.Contents)
		}
	}

	return result, nil
}
