package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/latonaio/aion-core/pkg/log"
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
	fmt.Printf("apply request: %s \n", reply.Message)
	return nil
}

func Apply(address string, aion *pb.AionSetting) error {
	conn, err := grpc.DialContext(
		context.Background(),
		address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)

	if err != nil {
		return err
	}
	return request(pb.NewProjectClient(conn), aion)
}

func Status(address string) (string, error) {
	conn, err := grpc.DialContext(
		context.Background(),
		address,
		grpc.WithInsecure(),
	)
	if err != nil {
		return "", err
	}
	log.Debugf("grpc con status: %v", conn.GetState().String())

	clt := pb.NewProjectClient(conn)
	statuses, err := clt.Status(context.Background(), &pb.Empty{})
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	byts, err := json.Marshal(statuses.Status)
	if err != nil {
		return "", err
	}
	err = json.Indent(&buf, byts, "", "  ")
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
