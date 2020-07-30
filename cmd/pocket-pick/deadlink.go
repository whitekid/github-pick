package main

import (
	"github.com/spf13/cobra"
	pocket "github.com/whitekid/pocket-pick/pkg"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:  "check-dead-link",
		Long: "check dead link",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pocket.CheckDeadLink()
		},
	})
}
