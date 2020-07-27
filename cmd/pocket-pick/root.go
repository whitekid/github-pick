package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/whitekid/pocket-pick/pkg"
)

var rootCmd = &cobra.Command{
	Use: "pocket-pick",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pocket.New().Serve(context.TODO(), args...)
	},
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetEnvPrefix("pp")
	viper.AutomaticEnv()
}
