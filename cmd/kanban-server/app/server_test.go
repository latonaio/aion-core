package app

const (
	port = 50051
	host = "localhost"
)

// func StartServerAndClient(ctx context.Context, t *testing.T) (StatusKanbanClient, *grpc.ClientConn) {
// 	g, ctx := errgroup.WithContext(ctx)
//
// 	g.Go(func() error {
// 		err := NewServer(port)
// 		return err
// 	})
//
// 	address := host + ":" + strconv.Itoa(port)
// 	client, err := grpc.DialContext(ctx, address, grpc.WithInsecure())
// 	if err != nil {
// 		t.Fatalf("did not connect: %v", err)
// 	}
// 	c := pb.NewStatusKanbanClient(client)
//
// 	return c, client
// }
//
