package i18n

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

var SupportedLanguages []string

var LangMap = map[string]string{}

// 默认语言
const DefaultLanguage = "en"

var translate = map[string][]byte{}

func InitI18n() error {
	// 读取locales目录
	langDirs, err := os.ReadDir("locales")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// 清空 SupportedLanguages 数组
	SupportedLanguages = []string{}

	// 遍历语言目录
	for _, langDir := range langDirs {
		if !langDir.IsDir() {
			continue
		}

		// 语言代码就是目录名
		lang := langDir.Name()
		SupportedLanguages = append(SupportedLanguages, lang)

		// 合并的JSON数据

		mergedData := make(map[string]map[string]any)
		mergedData[lang] = make(map[string]any)
		// 读取该语言目录下的所有JSON文件
		langPath := filepath.Join("locales", lang)
		files, err := os.ReadDir(langPath)
		if err != nil {
			return err
		}

		// 遍历语言目录下的所有JSON文件
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
				// 读取JSON文件内容
				jsonFilePath := filepath.Join(langPath, file.Name())
				jsonFile, err := os.Open(jsonFilePath)
				if err != nil {
					return err
				}

				jsonData, err := io.ReadAll(jsonFile)
				jsonFile.Close()
				if err != nil {
					return err
				}

				// 解析JSON数据
				var pageData map[string]any
				if err := json.Unmarshal(jsonData, &pageData); err != nil {
					return err
				}

				mergedData[lang][strings.TrimSuffix(file.Name(), ".json")] = pageData

			}
		}

		for k, v := range mergedData {
			mergedDataBytes, err := json.Marshal(v)
			if err != nil {
				return err
			}
			translate[k] = mergedDataBytes
		}

		// 设置语言显示名称
		langName := gjson.GetBytes(translate[lang], "root.lang").String()
		if langName != "" {
			LangMap[lang] = langName
		} else {
			LangMap[lang] = lang // 如果没有定义lang字段，使用语言代码作为显示名称
		}

	}

	return nil
}

func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		lang := ""

		if len(path) >= 3 {
			prefix := path[1:3] // 获取第一个斜杠后的两个字符
			for _, l := range SupportedLanguages {
				if prefix == l {
					lang = l
					break
				}
			}
		}

		if lang == "" {
			lang = DefaultLanguage // 使用默认语言
		}

		c.Set("lang", lang)

		c.Set("request", c.Request)

		c.Next()
	}
}

func GetLang(ctx *gin.Context) string {
	lang := ctx.GetString("lang")

	if lang == "" {
		lang = DefaultLanguage
	}

	return lang
}

func GetWithDefault(ctx *gin.Context, path string, defaultValue string) string {
	lang := GetLang(ctx)

	val := gjson.GetBytes(translate[lang], path).String()
	if val == "" {
		return defaultValue
	}

	return val
}

func Get(ctx *gin.Context, path string) string {
	return GetWithDefault(ctx, path, "")
}

func GetTranslations(ctx *gin.Context) map[string]any {
	lang := GetLang(ctx)

	_data := translate[lang]

	var translations map[string]any

	err := json.Unmarshal(_data, &translations)
	if err != nil {
		return nil
	}

	return translations
}
