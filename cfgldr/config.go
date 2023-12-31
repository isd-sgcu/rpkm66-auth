package cfgldr

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Dbnum    int    `mapstructure:"dbnum"`
}

type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSL      string `mapstructure:"ssl"`
}

type Service struct {
	Backend string `mapstructure:"backend"`
}

type App struct {
	Port            int    `mapstructure:"port"`
	Debug           bool   `mapstructure:"debug"`
	Secret          string `mapstructure:"secret"`
	MaxRestrictYear int    `mapstructure:"max_restrict_year"`
}

type ChulaSSO struct {
	Host         string `mapstructure:"host"`
	DeeAppID     string `mapstructure:"app-id"`
	DeeAppSecret string `mapstructure:"app-secret"`
}

type Jwt struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn int32  `mapstructure:"expires_in"`
	Issuer    string `mapstructure:"issuer"`
}

type Oauth struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectUri  string `mapstructure:"redirect_uri"`
}

type Config struct {
	Redis    Redis    `mapstructure:"redis"`
	Oauth    Oauth    `mapstructure:"google-oauth"`
	Database Database `mapstructure:"database"`
	App      App      `mapstructure:"app"`
	ChulaSSO ChulaSSO `mapstructure:"chula-sso"`
	Jwt      Jwt      `mapstructure:"jwt"`
	Service  Service  `mapstructure:"service"`
}

func LoadConfig() (config *Config, err error) {
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "error occurs while reading the config")
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.Wrap(err, "error occurs while unmarshal the config")
	}

	return
}

func LoadOauthConfig(oauth Oauth) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     oauth.ClientID,
		ClientSecret: oauth.ClientSecret,
		RedirectURL:  oauth.RedirectUri,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	}
}
