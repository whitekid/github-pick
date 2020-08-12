package config

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/whitekid/go-utils/log"
)

const (
	keyBind        = "bind_addr"
	keyRootURL     = "root-url"
	keyConsumerKey = "consumer-key"
	keyAccessToken = "access-token"
)

var configs = map[string][]struct {
	key          string
	short        string
	defaultValue interface{}
	description  string
}{
	"pocket-pick": {
		{keyBind, "B", "127.0.0.1:8000", "bind address"},
		{keyRootURL, "r", "http://127.0.0.0:8000", "root url"},
		{keyConsumerKey, "k", "", "getpocket consumer key"},
		{keyAccessToken, "a", "", "getpocket access token"},
	},
}

func BindAddr() string    { return viper.GetString(keyBind) }
func RootURL() string     { return viper.GetString(keyRootURL) }
func ConsumerKey() string { return viper.GetString(keyConsumerKey) }
func AccessToken() string { return viper.GetString(keyAccessToken) }

func init() {
	// InitDefaults initialize config
	for use := range configs {
		for _, config := range configs[use] {
			if config.defaultValue != nil {
				viper.SetDefault(config.key, config.defaultValue)
			}
		}
	}
}

func InitConfig() {
	viper.SetEnvPrefix("pp")
	viper.AutomaticEnv()
}

// InitFlagSet cobra.Command와 연결
func InitFlagSet(use string, fs *pflag.FlagSet) {
	for _, config := range configs[use] {
		switch v := config.defaultValue.(type) {
		case string:
			fs.StringP(config.key, config.short, v, config.description)
		case time.Duration:
			fs.DurationP(config.key, config.short, v, config.description)
		case []byte:
			fs.BytesHexP(config.key, config.short, v, config.description)
		default:
			log.Errorf("unsupported type %T", config.defaultValue)
		}
		viper.BindPFlag(config.key, fs.Lookup(config.key))
	}
}
