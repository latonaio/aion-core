package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "aionctl",
	Short: "aion command line tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aionctl")
	},
}

func init() {
	cobra.OnInitialize()
	RootCmd.AddCommand(
		applyCmd(),
	)
}
