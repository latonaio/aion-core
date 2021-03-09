package app

import (
	"fmt"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/services"

	"github.com/spf13/cobra"
)

func applyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply project file",
		Args:  cobra.RangeArgs(1, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			aion, err := config.LoadConfigFromFile(args[0])
			if err != nil {
				return err
			}
			if err := services.Apply("localhost:30644", aion); err != nil {
				fmt.Printf("failed to send project.yaml to service-broker")
			}
			if err := services.Apply("localhost:30655", aion); err != nil {
				fmt.Printf("failed to send project.yaml to service-broker")
			}
			return nil
		},
	}
	return cmd
}
