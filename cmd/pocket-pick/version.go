package main

import (
	"github.com/spf13/cobra"
	pocket "github.com/whitekid/pocket-pick/pkg"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			pocket.Version()
		},
	})
}
