package app

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:   "aionctl",
		Short: "aion command line tool",
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("\nBasic Commands\n  apply\n  status\n")
		},
	}
	HostName   = "localhost"
	PortNumber = "31110"
)

func init() {
	cobra.OnInitialize()
	RootCmd.PersistentFlags().StringVarP(&HostName, "host", "H", "localhost", "master broker server address")
	RootCmd.PersistentFlags().StringVarP(&PortNumber, "port", "P", "31110", "master broker server port")
	RootCmd.AddCommand(
		applyCmd(),
		statusCmd(),
	)
}
