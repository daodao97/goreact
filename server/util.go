package server

import (
	"embed"
	"encoding/json"
	"html/template"
	"strings"

	"github.com/daodao97/goreact/conf"
	"github.com/daodao97/goreact/model"
	"github.com/daodao97/xgo/xapp"
)

//go:embed templates
var Templates embed.FS

var functions template.FuncMap = template.FuncMap{
	"convertToJson": convertToJson,
}

func convertToJson(a any) string {
	s, _ := json.Marshal(a)
	return string(s)
}

type GeneralPayload struct {
	Translations           any
	Payload                any
	Template               string
	TemplateID             string
	ServerURL              string
	Component              string
	InnerHtmlContent       template.HTML
	Lang                   string
	Website                *conf.Website
	UserInfo               any
	GoogleAdsTxt           string
	GoogleAdsJS            string
	GoogleAnalytics        string
	MicrosoftClarityId     string
	Head                   *model.Head
	Version                string
	IsDev                  bool
	CloudflareTurnstileKey string
}

func extendPayload(
	data any,
	name string,
	component string,
	htmlContent template.HTML,
) *GeneralPayload {
	templateID := strings.ReplaceAll(name, "/", "-")
	templateID = strings.ReplaceAll(templateID, ".html", "")

	cloudflareTurnstile := conf.Get().Website.AuthConfig.CloudflareTurnstile
	cloudflareTurnstileKey := ""
	if cloudflareTurnstile != nil {
		cloudflareTurnstileKey = cloudflareTurnstile.SiteKey
	}

	return &GeneralPayload{
		Payload:                data,
		Template:               name,
		TemplateID:             templateID,
		Component:              component,
		InnerHtmlContent:       htmlContent,
		IsDev:                  xapp.IsDev(),
		CloudflareTurnstileKey: cloudflareTurnstileKey,
	}
}
