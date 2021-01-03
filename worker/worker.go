package main

import (
	"context"
	"crypto/tls"

	"github.com/spf13/pflag"
	"go.od2.network/hive/pkg/auth"
	"go.od2.network/hive/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	logConfig := zap.NewDevelopmentConfig()
	logConfig.DisableStacktrace = true
	logConfig.DisableCaller = true
	log, err := logConfig.Build()
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	// TODO remove token
	token := pflag.String("token", "HCyp15tCAAhXAZlV1WiHSy7ADxenGX_CXRq9dvnAzdT30", "Worker auth token")
	client, err := grpc.Dial(
		"worker.hive.od2.network:443",
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithPerRPCCredentials(&auth.WorkerCredentials{Token: *token}),
	)
	if err != nil {
		log.Fatal("Failed to connect to worker API", zap.Error(err))
	}
	assignments := types.NewAssignmentsClient(client)

	ctx := context.Background()
	_, err = assignments.CloseAssignmentsStream(ctx, &types.CloseAssignmentsStreamRequest{
		StreamId: 3,
	})
	if err != nil {
		log.Error("Failed to close stream", zap.Error(err))
	}
}
