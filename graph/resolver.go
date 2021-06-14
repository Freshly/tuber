package graph

import (
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/events"
	"go.uber.org/zap"
)

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db          *core.DB
	logger      *zap.Logger
	credentials []byte
	projectName string
	processor   *events.Processor
}

func NewResolver(db *core.DB, logger *zap.Logger, processor *events.Processor, credentials []byte, projectName string) *Resolver {
	return &Resolver{
		db:          db,
		logger:      logger,
		credentials: credentials,
		projectName: projectName,
		processor:   processor,
	}
}
