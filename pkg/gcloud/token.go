package gcloud

import (
	"context"

	"golang.org/x/oauth2/google"
)

// GetAccessToken gets a gcloud key from credentials json
func GetAccessToken(credentials []byte) (accessToken string, err error) {
	config, err := google.JWTConfigFromJSON(credentials,
		"https://www.googleapis.com/auth/devstorage.read_only",
	)

	if err != nil {
		return
	}

	ctx := context.Background()
	token, err := config.TokenSource(ctx).Token()
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
