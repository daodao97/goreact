package conf

type Website struct {
	Lang         string
	SupportLang  []string
	LangMap      map[string]string
	AuthProvider []AuthProvider `yaml:"auth_provider"`
	AuthConfig   AuthConfig     `yaml:"auth_config"`
}

type AuthConfig struct {
	CloudflareTurnstile *CloudflareTurnstile `yaml:"cloudflare_turnstile"`
}

type CloudflareTurnstile struct {
	SiteKey   string `yaml:"site_key"`
	SecretKey string `yaml:"secret_key"`
}

type AuthProviderType string

const (
	AuthProviderTypeGoogle AuthProviderType = "google"
	AuthProviderTypeGithub AuthProviderType = "github"
)

type AuthProvider struct {
	Provider     AuthProviderType `yaml:"provider"`
	ClientID     string           `yaml:"client_id"`
	ClientSecret string           `yaml:"client_secret"`
	RedirectURI  string           `yaml:"redirect_uri"`
}

type Header struct {
	Logo  string `json:"Logo,omitempty" yaml:"logo,omitempty"`
	Title string `json:"Title,omitempty" yaml:"title,omitempty"`
	Nav   []Link `json:"Nav,omitempty" yaml:"nav,omitempty"`
}

type Link struct {
	Text    string `yaml:"text"`
	URL     string `yaml:"url"`
	IsLogin bool   `yaml:"is_login"`
}

type LinkGroup struct {
	Title string `yaml:"title"`
	Links []Link `yaml:"links"`
}

type Footer struct {
	Logo      string      `yaml:"logo"`
	Title     string      `yaml:"title"`
	Desc      string      `yaml:"desc"`
	Social    []Social    `yaml:"social"`
	Links     []LinkGroup `yaml:"links"`
	Copyright string      `yaml:"copyright"`
	Policy    []Link      `yaml:"policy"`
}

type Social struct {
	Icon  string `yaml:"icon"`
	URL   string `yaml:"url"`
	Title string `yaml:"title"`
}
