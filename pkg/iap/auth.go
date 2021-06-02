package iap

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/freshly/tuber/pkg/iap/internal"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var refreshTokenFile string

func mustTuberConfigDir() string {
	basePath, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(basePath, "tuber")
}

func init() {
	refreshTokenFile = path.Join(mustTuberConfigDir(), "refresh_token")
}

func RefreshTokenExists() bool {
	spew.Dump(refreshTokenFile)
	_, err := os.Stat(refreshTokenFile)
	return !os.IsNotExist(err)
}

func config() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		ClientID:     viper.GetString("oauth-client-id"),
		ClientSecret: viper.GetString("oauth-client-secret"),
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}
}

func CreateRefreshToken() error {
	c := config()

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

func CreateIDToken() (string, error) {
	refreshToken, err := os.ReadFile(refreshTokenFile)
	if err != nil {
		return "", err
	}

	c := config()
	v := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {string(refreshToken)},
		"audience":      {viper.GetString("iap-client-id")},
	}

	token, err := internal.RetrieveToken(context.Background(), c.ClientID, c.ClientSecret, c.Endpoint.TokenURL, v, internal.AuthStyle(c.Endpoint.AuthStyle))
	if err != nil {
		return "", err
	}

	spew.Dump(token)

	vals, ok := token.Raw.(map[string]interface{})
	if !ok {
		return "", errors.New("could not assert raw token values")
	}

	return vals["id_token"].(string), nil
}