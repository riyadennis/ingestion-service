package graph

import (
	"github.com/sirupsen/logrus"

	"github.com/riyadennis/ingestion-service/business"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Uploader *business.BucketUpload
	Logger   *logrus.Logger
}

func NewResolver(logger *logrus.Logger, bu *business.BucketUpload) *Resolver {
	return &Resolver{
		Uploader: bu,
		Logger:   logger,
	}
}
