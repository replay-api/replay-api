// TODO: endpoint Validate (proto pequeno)
// TODO: endpoint GetUserDetails (squad +  membership, subscriptions) (proto pequeno)
// Define an RPC server that implements the generated gRPC interface.
package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	billing "github.com/replay-api/replay-api/cmd/rpc-api/proto/billing/generated/billing"
	iam "github.com/replay-api/replay-api/cmd/rpc-api/proto/iam/generated/iam"
	squad "github.com/replay-api/replay-api/cmd/rpc-api/proto/squad/generated/squad"
)

func main() {
	// In main, create the gRPC server and register the rpcServer.
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	iam.RegisterRIDServiceServer(grpcServer, &rpcServer{})
	squad.RegisterPlayerProfileServiceServer(grpcServer, &rpcServer{})
	billing.RegisterSubscriptionServiceServer(grpcServer, &rpcServer{})
	log.Printf("gRPC server is listening on %v", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
