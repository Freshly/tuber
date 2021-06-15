package iap

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/freshly/tuber/pkg/config"
	"github.com/freshly/tuber/pkg/iap/internal"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var refreshTokenFile string

func init() {
	dir, err := config.Dir()
	if err != nil {
		panic(err)
	}

	refreshTokenFile = path.Join(dir, "refresh_token")
}

func RefreshTokenExists() bool {
	_, err := os.Stat(refreshTokenFile)
	return !os.IsNotExist(err)
}

func newOAuthConfig() (*oauth2.Config, error) {
	cnf, err := config.Load()
	if err != nil {
		return nil, err
	}

	if cnf.Auth == nil {
		cnf.Auth = &config.Auth{}
	}

	return &oauth2.Config{
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		ClientID:     cnf.Auth.OAuthClientID,
		ClientSecret: cnf.Auth.OAuthSecret,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}, nil
}

func CreateRefreshToken() error {
	c, err := newOAuthConfig()
	if err != nil {
		return err
	}

	fmt.Println("Go to the following URL and paste the code back here, ok?")
	fmt.Println(c.AuthCodeURL("", oauth2.AccessTypeOffline))
	fmt.Println("")
	fmt.Print("Auth Code: ")

	var code string
	fmt.Scanln(&code)

	token, err := c.Exchange(context.Background(), code)
	if err != nil {
		return err
	}

	return os.WriteFile(refreshTokenFile, []byte(token.RefreshToken), 0600)
}

func CreateIDToken(IAPClientID string) (string, error) {
	refreshToken, err := os.ReadFile(refreshTokenFile)
	if err != nil {
		return "", err
	}

	c, err := newOAuthConfig()
	if err != nil {
		return "", err
	}

	v := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {string(refreshToken)},
		"audience":      {IAPClientID},
	}

	token, err := internal.RetrieveToken(context.Background(), c.ClientID, c.ClientSecret, c.Endpoint.TokenURL, v, internal.AuthStyle(c.Endpoint.AuthStyle))
	if err != nil {
		return "", err
	}

	vals, ok := token.Raw.(map[string]interface{})
	if !ok {
		return "", errors.New("could not assert raw token values")
	}

	return vals["id_token"].(string), nil
}
