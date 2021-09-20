package iap

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/freshly/tuber/pkg/config"
	"github.com/freshly/tuber/pkg/iap/internal"
	"github.com/goccy/go-yaml"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func RefreshTokenExists(audience string) (bool, error) {
	path, err := RefreshTokenPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

func dir() (string, error) {
	basePath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(basePath, "tuber"), nil
}

func RefreshTokenPath() (string, error) {
	directory, err := dir()
	if err != nil {
		return "", fmt.Errorf("config not set up, please run 'tuber config'")
	}
	return path.Join(directory, "refresh_tokens"), nil
}

type refreshTokens struct {
	Tokens []*audienceToken `yaml:"tokens"`
}

type audienceToken struct {
	Audience     string `yaml:"audience"`
	RefreshToken string `yaml:"refreshToken"`
}

func LoadOrCreateRefreshTokens() (*refreshTokens, error) {
	path, err := RefreshTokenPath()
	if err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		_, err = os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("unable to create tuber auth: %v", err)
		}
		raw, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("created tuber auth but reading failed: %v", err)
		}
	}

	var rts refreshTokens
	err = yaml.Unmarshal(raw, &rts)
	if err != nil {
		return nil, fmt.Errorf("tuber auth invalid, please run `tuber auth`")
	}

	return &rts, nil
}

func newOAuthConfig() (*oauth2.Config, error) {
	cnf, err := config.Load()
	if err != nil {
		return nil, err
	}

	cluster, err := cnf.CurrentClusterConfig()
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		ClientID:     cluster.Auth.TuberDesktopClientID,
		ClientSecret: cluster.Auth.TuberDesktopClientSecret,
		Scopes:       []string{"openid", "email", "https://www.googleapis.com/auth/cloud-platform"},
		Endpoint:     google.Endpoint,
	}, nil
}

func CreateRefreshToken() error {
	cnf, err := config.Load()
	if err != nil {
		return err
	}

	cluster, err := cnf.CurrentClusterConfig()
	if err != nil {
		return err
	}

	if cluster.Auth.TuberDesktopClientID == "" || cluster.Auth.TuberDesktopClientSecret == "" || cluster.Auth.Audience == "" {
		return fmt.Errorf("tuber config auth section incomplete for this cluster, please run 'tuber config'")
	}

	rts, err := LoadOrCreateRefreshTokens()
	if err != nil {
		return err
	}

	path, err := RefreshTokenPath()
	if err != nil {
		return err
	}

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

	var tokens []*audienceToken
	var found bool
	for _, t := range rts.Tokens {
		if t.Audience == cluster.Auth.Audience {
			found = true
			t.RefreshToken = token.RefreshToken
		}
		tokens = append(tokens, t)
	}
	if !found {
		tokens = append(tokens, &audienceToken{
			Audience:     cluster.Auth.Audience,
			RefreshToken: token.RefreshToken,
		})
	}

	rts.Tokens = tokens

	out, err := yaml.Marshal(rts)
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, os.ModePerm)
}

type OauthTokens struct {
	IDToken      string
	RefreshToken string
	AccessToken  string
	Raw          *internal.Token
}

func CreateIDToken(IAPAudience string) (*OauthTokens, error) {
	rts, err := LoadOrCreateRefreshTokens()
	if err != nil {
		return nil, fmt.Errorf("refresh token blank for this context, please run 'tuber auth'")
	}

	var activeToken *audienceToken
	for _, t := range rts.Tokens {
		if t.Audience == IAPAudience {
			activeToken = t
		}
	}
	if activeToken == nil || activeToken.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token blank for this context, please run 'tuber auth'")
	}

	c, err := newOAuthConfig()
	if err != nil {
		return nil, err
	}

	v := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {activeToken.RefreshToken},
		"audience":      {IAPAudience},
	}

	token, err := internal.RetrieveToken(context.Background(), c.ClientID, c.ClientSecret, c.Endpoint.TokenURL, v, internal.AuthStyle(c.Endpoint.AuthStyle))
	if err != nil {
		return nil, err
	}

	vals, ok := token.Raw.(map[string]interface{})
	if !ok {
		return nil, errors.New("could not assert raw token values")
	}

	return &OauthTokens{IDToken: vals["id_token"].(string), AccessToken: vals["access_token"].(string), RefreshToken: activeToken.RefreshToken, Raw: token}, nil
}
