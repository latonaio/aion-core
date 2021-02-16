package app

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/latonaio/aion-core/config"
	pb "bitbucket.org/latonaio/aion-core/proto/projectpb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func request(client pb.ProjectClient, aion *pb.AionSetting) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Minute,
	)
	defer cancel()
	reply, err := client.Apply(ctx, aion)
	if err != nil {
		return err
	}
	fmt.Printf("request: %s", reply.Message)
	return nil
}

func apply(address string, aion *pb.AionSetting) error {
	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return err
	}
	client := pb.NewProjectClient(conn)
	return request(client, aion)
}

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
			return apply("localhost:30644", aion)
		},
	}
	return cmd
}
