package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pocket "github.com/whitekid/pocket-pick/pkg"
)

var rootCmd = &cobra.Command{
	Use: "pocket-pick",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pocket.New().Serve(context.TODO(), args...)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	fs := rootCmd.Flags()

	fs.StringP(pocket.KeyBind, "B", "127.0.0.1:8000", "bind address")
	viper.BindPFlag(pocket.KeyBind, fs.Lookup(pocket.KeyBind))

}

func initConfig() {
	viper.SetEnvPrefix("pp")
	viper.AutomaticEnv()
}
