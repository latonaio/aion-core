package app

import (
	"fmt"

	"bitbucket.org/latonaio/aion-core/pkg/log"

	"bitbucket.org/latonaio/aion-core/config"
	"bitbucket.org/latonaio/aion-core/internal/services"

	"github.com/spf13/cobra"
)

func applyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply services file",
		Args:  cobra.RangeArgs(1, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hostName := fmt.Sprintf("%s:%s", HostName, PortNumber)
			fmt.Printf("aion-master:%v\n", hostName)
			if len(args) == 0 {
				return nil
			}
			aion, err := config.LoadConfigFromFile(args[0])
			if err != nil {
				return err
			}
			log.Debugf("aionctl apply : %+v \n", aion)
			if err := services.Apply(hostName, aion); err != nil {
				fmt.Printf("failed to send services.yml to service-broker cause: %v \n", err)
			}
			return nil
		},
	}
	return cmd
}
