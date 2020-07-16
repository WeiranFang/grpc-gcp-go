package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"

	monitoredres "google.golang.org/genproto/googleapis/api/monitoredres"
	loggingpb "google.golang.org/genproto/googleapis/logging/v2"
)

var (
	scope  = "https://www.googleapis.com/auth/cloud-platform"
	target = "logging.googleapis.com:443"
	comp   = flag.Bool("comp", false, "enable gzip grpc compression")
)

func main() {
	flag.Parse()
	keyFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	perRPC, err := oauth.NewServiceAccountFromFile(keyFile, scope)
	if err != nil {
		fmt.Errorf("got error when getting credentials, err: ", err)
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		grpc.WithPerRPCCredentials(perRPC),
	}
	if *comp {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	}
	conn, err := grpc.Dial(
		target,
		opts...,
	)
	if err != nil {
		fmt.Println("dial err: ", err)
		os.Exit(1)
	}
	client := loggingpb.NewLoggingServiceV2Client(conn)
	entry := loggingpb.LogEntry{
		Resource: &monitoredres.MonitoredResource{
			Type: "global",
		},
		LogName: "projects/grpc-gcp/logs/porkpie",
		Payload: &loggingpb.LogEntry_TextPayload{TextPayload: "xxxxxxxxxxxxxxxxxxxx"},
	}
	req := loggingpb.WriteLogEntriesRequest{
		Entries: []*loggingpb.LogEntry{&entry},
	}
	fmt.Printf("----- req: %+v\n", &req)
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "grpc-encoding", "gzip")
	resp, err := client.WriteLogEntries(
		ctx,
		&req,
	)
	if err != nil {
		fmt.Println("----- RPC error: ", err)
		os.Exit(1)
	}
	fmt.Printf("----- response: %+v\n", *resp)
}
