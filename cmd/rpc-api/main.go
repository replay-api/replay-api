// TODO: endpoint Validate (proto pequeno)
// TODO: endpoint GetUserDetails (squad +  membership, subscriptions) (proto pequeno)
// Define an RPC server that implements the generated gRPC interface.
package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	billing "github.com/replay-api/replay-api/pkg/app/proto/billing/generated/billing"
	iam "github.com/replay-api/replay-api/pkg/app/proto/iam/generated/iam"
	squad "github.com/replay-api/replay-api/pkg/app/proto/squad/generated/squad"
	"github.com/replay-api/replay-api/pkg/infra/ioc"
)

func main() {

	builder := ioc.NewContainerBuilder()

	c := builder.WithEnvFile().WithKafka().With(ioc.InjectMongoDB).WithSquadAPI().WithInboundPorts().Build()

	rpcPort := os.Getenv("GRPC_API_PORT")

	// In main, create the gRPC server and register the rpcServer.
	lis, err := net.Listen("tcp", ":"+rpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	api := NewRPCServer(c)
	grpcServer := grpc.NewServer()
	iam.RegisterRIDServiceServer(grpcServer, api)
	squad.RegisterPlayerProfileServiceServer(grpcServer, api)
	billing.RegisterSubscriptionServiceServer(grpcServer, api)
	
	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	
	log.Printf("gRPC server is listening on %v", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
