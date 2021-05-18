package graph

import "github.com/freshly/tuber/pkg/db"

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db *db.DB
}

func NewResolver(db *db.DB) *Resolver {
	return &Resolver{db: db}
}
