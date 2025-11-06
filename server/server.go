package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	rest "github.com/riyadennis/ingestion-service/handlers"
	"github.com/sirupsen/logrus"
)

const timeOut = 5 * time.Second

var (
	errEmptyPort           = errors.New("port number empty")
	errPortNotAValidNumber = errors.New("port number is not a valid number")
	errPortReserved        = errors.New("port is a reserved number")
	errPortBeyondRange     = errors.New("port is beyond the allowed range")
)

// Server have all the setup needed to run and shut down a http server
type Server struct {
	restServer  http.Server
	ServerError chan error
	ShutDown    chan os.Signal
}

// NewServer creates a server instance with error and shutdown channels initialized
func NewServer(restPort string) *Server {
	errChan := make(chan error, 2)
	shutdown := make(chan os.Signal, 1)

	err := validatePort(restPort)
	if err != nil {
		errChan <- err
	}
	return &Server{
		restServer: http.Server{
			Addr:         ":" + restPort,
			ReadTimeout:  timeOut,
			WriteTimeout: timeOut,
		},
		ServerError: errChan,
		ShutDown:    shutdown,
	}
}

// Run registers routes and starts a webserver
// and waits to receive from shutdown and error channels
func (s *Server) Run(logger *logrus.Logger, client *minio.Client, bucketName string) error {
	s.restServer.Handler = rest.LoadRESTEndpoints(logger, client, bucketName)
	// Start the service
	go func() {
		logger.Printf("server running on port %s", s.restServer.Addr)
		s.ServerError <- s.restServer.ListenAndServe()
	}()

	select {
	case err := <-s.ServerError:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-s.ShutDown:
		logger.Printf("main: %v: Start shutdown", sig)
		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), timeOut)
		defer cancel()

		err := s.restServer.Shutdown(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func validatePort(port string) error {
	if port == "" {
		return errEmptyPort
	}

	addr, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return errPortNotAValidNumber
	}

	if addr < 1024 {
		return errPortReserved
	}

	if addr > 65535 {
		return errPortBeyondRange
	}

	return nil
}
