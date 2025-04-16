package conf

import (
	"github.com/daodao97/goreact/i18n"
)

type Conf struct {
	AppID              string  `json:"app_id" yaml:"app_id"`
	GoogleAdsTxt       string  `yaml:"google_ads_txt"`
	GoogleAdsJS        string  `yaml:"google_ads_js"`
	GoogleAnalytics    string  `yaml:"google_analytics"`
	JwtSecret          string  `json:"jwt_secret" yaml:"jwt_secret"`
	GitTag             string  `json:"git_tag" yaml:"git_tag"`
	MicrosoftClarityId string  `yaml:"microsoft_clarity_id"`
	Website            Website `json:"website" yaml:"website"`
}

var conf *Conf

func SetConf(c *Conf) {
	c.Website.Lang = i18n.DefaultLanguage
	c.Website.SupportLang = i18n.SupportedLanguages
	c.Website.LangMap = i18n.LangMap
	conf = c
}

func Get() *Conf {
	return conf
}
