package gapi

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	gRPCGWUserAgent = "grpcgateway-user-agent"
	gRPCForwaredFor = "x-forwarded-for"
	gRPCUserAgent   = "user-agent"
)

type metaData struct {
	userAgent string
	clientIP  string
}

func extractMetaData(ctx context.Context) *metaData {

	mtd := &metaData{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Printf("%+v", md)

		// If coming through gateway, parse out the user agent.
		if userAgents := md.Get(gRPCGWUserAgent); len(userAgents) > 0 {
			mtd.userAgent = userAgents[0]
		}

		// If coming through GRPC client, parse out the user agent.
		if userAgents := md.Get(gRPCUserAgent); len(userAgents) > 0 {
			mtd.userAgent = userAgents[0]
		}
		// Parse out the client ip if coming through gateway.
		if clientIPs := md.Get(gRPCForwaredFor); len(clientIPs) > 0 {
			mtd.clientIP = clientIPs[0]
		}
	}

	// The client ip is available through peer package if coming through gRPC client.
	if peer, ok := peer.FromContext(ctx); ok {
		addr := peer.LocalAddr.String()
		mtd.clientIP = addr
	}

	return mtd
}
