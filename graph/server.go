package graph

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/riyadennis/identity-server/app/proto/identity"
	"github.com/riyadennis/ingestion-service/business"
	"github.com/riyadennis/ingestion-service/graph/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ErrFailedToStartListener means that the listener couldn't be started
var ErrFailedToStartListener = errors.New("failed to start listener")

// HTTPServer encapsulates two http server operations  that we need to execute in the service
// it is mainly helpful for testing, by creating mocks for http calls.
type HTTPServer interface {
	Shutdown(ctx context.Context) error
	Serve(l net.Listener) error
}
type Server struct {
	Server   HTTPServer
	Logger   *logrus.Logger
	ShutDown chan os.Signal
}

func NewServer(logger *logrus.Logger, bu *business.BucketUpload, port, identityURL string) *Server {
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	// Create gRPC connection to identity server
	// Use 127.0.0.1 instead of localhost to force IPv4
	conn, err := grpc.NewClient(identityURL, opts...)

	if err != nil {
		logger.Fatalf("failed to create gRPC client: %v", err)
	}
	identityClient := identity.NewIdentityClient(conn)
	logger.Infof("gRPC client created for identity server at %s", identityURL)

	resolver := NewResolver(logger, bu)
	srv := handler.New(generated.NewExecutableSchema(
		generated.Config{
			Resolvers: resolver,
		},
	))
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.Use(extension.Introspection{})

	addr := fmt.Sprintf(":%s", port)
	server := &Server{
		Server: &http.Server{
			Addr:    addr,
			Handler: newRouter(srv, identityClient, logger),
		},
		Logger:   logger,
		ShutDown: make(chan os.Signal, 1),
	}
	return server
}

func (s *Server) Start(port string) error {
	s.Logger.Info("starting service", "port", port)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		s.Logger.Errorf("failed to start http listener: %v", err)
		return ErrFailedToStartListener
	}
	var sErr error
	go func() {
		s.Logger.Info("service finished starting and is now ready to accept requests")

		// start http listener
		sErr = s.Server.Serve(listener)
		if sErr != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.Logger.Errorf("failed to start http server: %v", sErr)
				return
			}
		}
	}()

	return sErr
}

func newRouter(srv *handler.Server, client identity.IdentityClient, logger *logrus.Logger) http.Handler {
	chiRouter := chi.NewRouter()

	chiRouter.Use(middleware.RequestID)
	chiRouter.Use(middleware.Recoverer)
	chiRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))
	chiRouter.Use(NeedsAuthMiddleWare(client, logger))
	chiRouter.Handle("/", playground.Handler("GraphQL playground", "/graphql"))

	chiRouter.Handle("/graphql", srv)
	return chiRouter
}
