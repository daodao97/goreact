package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/daodao97/goreact/util"
	esbuild "github.com/evanw/esbuild/pkg/api"
)

/*

使用 esbuild 打包 前端代码文件

*/

const textEncoderPolyfill = `function TextEncoder(){} TextEncoder.prototype.encode=function(string){var octets=[],length=string.length,i=0;while(i<length){var codePoint=string.codePointAt(i),c=0,bits=0;codePoint<=0x7F?(c=0,bits=0x00):codePoint<=0x7FF?(c=6,bits=0xC0):codePoint<=0xFFFF?(c=12,bits=0xE0):codePoint<=0x1FFFFF&&(c=18,bits=0xF0),octets.push(bits|(codePoint>>c)),c-=6;while(c>=0){octets.push(0x80|((codePoint>>c)&0x3F)),c-=6}i+=codePoint>=0x10000?2:1}return octets};function TextDecoder(){} TextDecoder.prototype.decode=function(octets){var string="",i=0;while(i<octets.length){var octet=octets[i],bytesNeeded=0,codePoint=0;octet<=0x7F?(bytesNeeded=0,codePoint=octet&0xFF):octet<=0xDF?(bytesNeeded=1,codePoint=octet&0x1F):octet<=0xEF?(bytesNeeded=2,codePoint=octet&0x0F):octet<=0xF4&&(bytesNeeded=3,codePoint=octet&0x07),octets.length-i-bytesNeeded>0?function(){for(var k=0;k<bytesNeeded;){octet=octets[i+k+1],codePoint=(codePoint<<6)|(octet&0x3F),k+=1}}():codePoint=0xFFFD,bytesNeeded=octets.length-i,string+=String.fromCodePoint(codePoint),i+=bytesNeeded+1}return string};`

const messageChannelPolyfill = `if(typeof MessageChannel==="undefined"){var MessageChannel=function(){this.port1={postMessage:function(msg){setTimeout(()=>{this.onmessage&&this.onmessage({data:msg})},0)},onmessage:null},this.port2={postMessage:function(msg){setTimeout(()=>{this.onmessage&&this.onmessage({data:msg})},0)},onmessage:null}}}`

const processPolyfill = `var process = {env: {NODE_ENV: "production"}};`

func aliasPlugin(aliases map[string]string) esbuild.Plugin {
	return esbuild.Plugin{
		Name: "alias-resolver",
		Setup: func(build esbuild.PluginBuild) {
			build.OnResolve(esbuild.OnResolveOptions{Filter: "^@/"}, func(args esbuild.OnResolveArgs) (esbuild.OnResolveResult, error) {
				// 将 @/ 替换为实际路径
				newPath := strings.TrimPrefix(args.Path, "@/")
				basePath := aliases["@"] + "/" + newPath

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
			build.OnResolve(esbuild.OnResolveOptions{Filter: "^#core/"}, func(args esbuild.OnResolveArgs) (esbuild.OnResolveResult, error) {
				// 将 #core/ 替换为实际路径 ./core/ui/
				newPath := strings.TrimPrefix(args.Path, "#core/")
				basePath := "./core/ui/" + newPath // 硬编码目标路径

				// 尝试不同的扩展名
				extensions := []string{".tsx", ".ts", ".jsx", ".js"} // 优先查找带扩展名的文件
				for _, ext := range extensions {
					fullPath := basePath + ext
					if _, err := os.Stat(fullPath); err == nil {
						return esbuild.OnResolveResult{Path: fullPath, External: false}, nil
					}
				}
				// 尝试不带扩展名的文件（可能是目录）
				if _, err := os.Stat(basePath); err == nil {
					// 检查是否是目录，如果是目录，则尝试 index 文件
					isDir, _ := os.Stat(basePath)
					if isDir.IsDir() {
						for _, ext := range extensions {
							fullPath := basePath + "/index" + ext
							if _, err := os.Stat(fullPath); err == nil {
								return esbuild.OnResolveResult{Path: fullPath, External: false}, nil
							}
						}
					} else {
						// 如果是文件，直接返回
						return esbuild.OnResolveResult{Path: basePath, External: false}, nil
					}
				}

				fmt.Printf("核心库路径解析失败: %s -> %s\n", args.Path, basePath)
				// 返回原始路径，让 esbuild 继续尝试解析或其他插件处理
				return esbuild.OnResolveResult{Path: args.Path, External: false}, nil // 返回原始路径以便其他解析器处理
			})
		},
	}
}

func BuildClientComponents(jsFolder, jsOutput string, aliases map[string]string) error {
	fmt.Println("Building client Javascript")

	filesJSX, err := util.GetFiles(jsFolder, ".jsx")
	if err != nil {
		return err
	}

	filesTSX, err := util.GetFiles(jsFolder, ".tsx")
	if err != nil {
		return err
	}

	allFiles := append(filesJSX, filesTSX...)
	allFiles = append(allFiles, "./frontend/app.js")

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
		Plugins: []esbuild.Plugin{aliasPlugin(aliases)},
	})

	if len(builds.Errors) > 0 {
		return fmt.Errorf("error on esbuild: %v", builds.Errors)
	}

	for _, file := range builds.OutputFiles {
		fmt.Println("Created file:", file.Path)
	}

	return nil
}

func BuildServerComponents(jsFolder, jsOutput string, aliases map[string]string) (map[string]string, error) {
	fmt.Println("Building server Javascript")
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
		Plugins: []esbuild.Plugin{aliasPlugin(aliases)},
	})

	if len(builds.Errors) > 0 {
		return result, fmt.Errorf("error on esbuild: %v", builds.Errors)
	}

	for _, file := range builds.OutputFiles {
		if strings.Contains(file.Path, jsOutput) {
			paths := strings.Split(file.Path, jsOutput)
			path := ""
			if len(paths) >= 2 {
				path = strings.Join(paths[1:], "")
			}
			result[path] = string(file.Contents)
			fmt.Println("Server file built in:", path)

		}
	}

	return result, nil
}
