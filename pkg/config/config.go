package config

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/whitekid/go-utils/flags"
)

const (
	keyBind         = "bind_addr"
	keyRootURL      = "root_url"
	keyConsumerKey  = "consumer_key"
	keyAccessToken  = "access_token"
	keyCacheTimeout = "favorite_cache_timeout"
)

var configs = map[string][]flags.Flag{
	"pocket-pick": {
		{keyBind, "B", "127.0.0.1:8000", "bind address"},
		{keyRootURL, "r", "http://127.0.0.0:8000", "root url"},
		{keyConsumerKey, "k", "", "getpocket consumer key"},
		{keyAccessToken, "a", "", "getpocket access token"},
		{keyCacheTimeout, "", time.Hour, "timeout for cache favorite items"},
	},
}

func init() {
	viper.SetEnvPrefix("pp")
	viper.AutomaticEnv()

	flags.InitDefaults(configs)
}

func InitFlagSet(use string, fs *pflag.FlagSet) { flags.InitFlagSet(configs, use, fs) }

// Config access functions
func BindAddr() string                    { return viper.GetString(keyBind) }
func RootURL() string                     { return viper.GetString(keyRootURL) }
func ConsumerKey() string                 { return viper.GetString(keyConsumerKey) }
func AccessToken() string                 { return viper.GetString(keyAccessToken) }
func CacheEvictionTimeout() time.Duration { return viper.GetDuration(keyCacheTimeout) }
