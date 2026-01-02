package business

import (
	"errors"
	"time"

	"github.com/riyadennis/identity-server/app/proto/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func IdentityClient(url string) (identity.IdentityClient, error) {
	if url == "" {
		return nil, errors.New("empty url for identity gRPC client")
	}
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Create gRPC connection to identity server
	conn, err := grpc.NewClient(url, opts...)

	if err != nil {
		return nil, err
	}

	return identity.NewIdentityClient(conn), nil
}
