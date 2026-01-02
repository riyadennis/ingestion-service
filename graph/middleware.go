package graph

import (
	"context"
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
			// Extract authorization token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Expected format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
			token := parts[1]

			// Create gRPC metadata with the token
			ctx := r.Context()
			md := metadata.New(map[string]string{
				"Authorization": "Bearer " + token,
			})
			grpcCtx := metadata.NewOutgoingContext(ctx, md)

			// Call identity service to validate token and get user info
			userResp, err := client.Me(grpcCtx, &identity.UserRequest{})
			if err != nil {
				logger.Errorf("failed to validate token: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Store user info in context
			ctx = context.WithValue(ctx, UserIDContextKey, userResp.ID)

			// Continue to next handler with user context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
