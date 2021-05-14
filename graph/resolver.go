package graph

import (
	"github.com/freshly/tuber/pkg/core"
)

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db *core.Data
}

func NewResolver(db *core.Data) *Resolver {
	return &Resolver{db: db}
}
