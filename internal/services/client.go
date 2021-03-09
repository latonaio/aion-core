package services

import (
	"context"
	"fmt"
	"time"

	pb "bitbucket.org/latonaio/aion-core/proto/projectpb"
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

func Apply(address string, aion *pb.AionSetting) error {
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
