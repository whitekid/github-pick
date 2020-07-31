package config

import "github.com/spf13/viper"

const (
	// KeyBind binding address
	KeyBind = "bind_addr"
)

func BindAddr() string { return viper.GetString(KeyBind) }
