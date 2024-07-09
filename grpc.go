package main

import (
	"fmt"
	pb "streakai/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initGRPCConnection() error {
	var err error
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server at localhost:50051: %v", err)
	}
	grpcClient = pb.NewStreakAiServiceClient(conn)
	return nil
}
