package app

import (
	"fmt"

	"bitbucket.org/latonaio/aion-core/internal/services"

	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show services status from all nodes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			hostName := fmt.Sprintf("%v:%v", HostName, PortNumber)
			fmt.Printf("aion-master:%v \n", hostName)
			jsnData, err := services.Status(hostName)
			if err != nil {
				return err
			}
			fmt.Print(jsnData)
			return nil
		},
	}
	return cmd
}
