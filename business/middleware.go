package business

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/riyadennis/identity-server/app/proto/identity"
)

type GraphQLRequest struct {
	OperationName string `json:"operationName"`
	Query         string `json:"query"`
}

type contextKey string

const UserIDContextKey contextKey = "userID"

func NeedsAuthMiddleWare(client identity.IdentityClient, logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			userID, err := UsrIDFromToken(ctx, r, client)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			// Store user info in context
			ctx = context.WithValue(ctx, UserIDContextKey, userID)

			// Continue to next handler with user context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UsrIDFromToken(ctx context.Context, r *http.Request, client identity.IdentityClient) (string, error) {
	// Extract authorization token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	// Expected format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}
	token := parts[1]

	// Create gRPC metadata with the token
	md := metadata.New(map[string]string{
		"Authorization": "Bearer " + token,
	})
	grpcCtx := metadata.NewOutgoingContext(ctx, md)

	// Call identity service to validate token and get user info
	userResp, err := client.Me(grpcCtx, &identity.UserRequest{})
	if err != nil {
		return "", err
	}

	if userResp == nil {
		return "", errors.New("empty response from identity server")
	}

	return *userResp.ID, nil
}
