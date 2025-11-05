package conf

import (
	"strings"

	"github.com/daodao97/goreact/i18n"
)

type Conf struct {
	AppID              string      `json:"app_id" yaml:"app_id"`
	GoogleAdsTxt       string      `yaml:"google_ads_txt"`
	GoogleAdsJS        string      `yaml:"google_ads_js"`
	GoogleAnalytics    string      `yaml:"google_analytics"`
	JwtSecret          string      `json:"jwt_secret" yaml:"jwt_secret"`
	GitTag             string      `json:"git_tag" yaml:"git_tag"`
	MicrosoftClarityId string      `yaml:"microsoft_clarity_id"`
	Website            Website     `json:"website" yaml:"website"`
	EmailFilter        EmailFilter `json:"email_blacklist" yaml:"email_blacklist"`
}

type EmailFilter struct {
	Mode     string   `json:"mode" yaml:"mode"`
	Suffixes []string `json:"suffixes" yaml:"suffixes"`
	Keywords []string `json:"keywords" yaml:"keywords"`
	Exact    []string `json:"exact" yaml:"exact"`
}

var conf *Conf

func SetConf(c *Conf) {
	c.Website.Lang = i18n.DefaultLanguage
	c.Website.SupportLang = i18n.SupportedLanguages
	c.Website.LangMap = i18n.LangMap
	mode := strings.ToLower(strings.TrimSpace(c.EmailFilter.Mode))
	if mode == "" {
		mode = "blacklist"
	}
	c.EmailFilter.Mode = mode
	conf = c
}

func Get() *Conf {
	return conf
}
