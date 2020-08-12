package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/spf13/cobra"
	pocket "github.com/whitekid/pocket-pick/pkg"
	"github.com/whitekid/pocket-pick/pkg/config"
)

var rootCmd = &cobra.Command{
	Use: "pocket-pick",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pocket.New().Serve(context.TODO(), args...)
	},
}

func init() {
	rand.Seed(time.Now().UnixNano())
	cobra.OnInitialize(config.InitConfig)

	config.InitFlagSet(rootCmd.Use, rootCmd.Flags())
}
