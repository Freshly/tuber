package graph

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/oauth"
	"go.uber.org/zap"
)

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db                *core.DB
	logger            *zap.Logger
	credentials       []byte
	projectName       string
	processor         *events.Processor
	clusterName       string
	clusterRegion     string
	reviewAppsEnabled bool
}

func NewResolver(db *core.DB, logger *zap.Logger, processor *events.Processor, credentials []byte, projectName string, clusterName string, clusterRegion string, reviewAppsEnabled bool) *Resolver {
	return &Resolver{
		db:                db,
		logger:            logger,
		credentials:       credentials,
		projectName:       projectName,
		processor:         processor,
		clusterName:       clusterName,
		clusterRegion:     clusterRegion,
		reviewAppsEnabled: reviewAppsEnabled,
	}
}

func canUpdateDeployments(ctx context.Context, appName string) error {
	return authCheck(ctx, appName, "update", "deployments")
}

func canDeleteDeployments(ctx context.Context, appName string) error {
	return authCheck(ctx, appName, "delete", "deployments")
}

func canGetDeployments(ctx context.Context, appName string) error {
	return authCheck(ctx, appName, "get", "deployments")
}

func canCreateDeployments(ctx context.Context, appName string) error {
	return authCheck(ctx, appName, "create", "deployments")
}

func canGetSecret(ctx context.Context, appName string, secretName string) error {
	return authCheck(ctx, appName, "get", "secret/"+secretName)
}

func canUpdateSecret(ctx context.Context, appName string, secretName string) error {
	return authCheck(ctx, appName, "update", "secret/"+secretName)
}

func canCreateApps(ctx context.Context) error {
	return authCheckAllNamespaces(ctx, "create", "deployments")
}

func canViewAllApps(ctx context.Context) error {
	return authCheckAllNamespaces(ctx, "view", "deployments")
}

func authCheckAllNamespaces(ctx context.Context, verb string, subject string) error {
	token, err := oauth.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving authorization params")
	}
	authorized, err := k8s.CanIAllNamespaces(verb, subject, "--token="+token)
	if err != nil {
		return fmt.Errorf("error determining authorization status")
	}
	if !authorized {
		return fmt.Errorf("unauthorized to perform this action")
	}
	return nil
}

func authCheck(ctx context.Context, appName string, verb string, subject string) error {
	token, err := oauth.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving authorization params")
	}
	authorized, err := k8s.CanI(appName, verb, subject, "--token="+token)
	if err != nil {
		return fmt.Errorf("error determining authorization status")
	}
	if !authorized {
		return fmt.Errorf("unauthorized to perform this action")
	}
	return nil
}
