package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/riyadennis/ingestion-service/business"
	"github.com/sirupsen/logrus"
)

const (
	// UploadEndpoint is to upload files
	UploadEndpoint = "/upload"

	// LivenessEndPoint is for kubernetes to check when to restart the container
	LivenessEndPoint = "/liveness"

	// ReadinessEndPoint is for kubernetes to check when the container is read to accept traffic
	ReadinessEndPoint = "/readiness"
)

// LoadRESTEndpoints adds REST endpoints to the router
func LoadRESTEndpoints(logger *logrus.Logger, client business.Storage, bucketName string) http.Handler {
	r := chi.NewRouter()
	// wrap already initialised logger to Chi logger
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger}))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Set a timeout value on the request context (ctx) that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc: func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value isn't ignored by any of the major browsers
	}))
	r.Get(LivenessEndPoint, Liveness)
	r.Get(ReadinessEndPoint, Ready)

	h := NewUploader(logger, business.NewBucketUpload(client, bucketName))
	r.Post(UploadEndpoint, h.Upload)

	return r
}
