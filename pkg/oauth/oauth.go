package oauth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Authenticator struct {
	oauthConfig   *oauth2.Config
	oauthStateKey string
}

func NewAuthenticator(oauthRedirectUrl string, oauthClientSecret string, oauthClientID string, oauthStateKey string) *Authenticator {
	config := &oauth2.Config{
		RedirectURL:  oauthRedirectUrl,
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		Scopes:       []string{"openid", "email", "https://www.googleapis.com/auth/cloud-platform"},
		Endpoint:     google.Endpoint,
	}
	return &Authenticator{
		oauthConfig:   config,
		oauthStateKey: oauthStateKey,
	}
}

func AccessTokenHeaderKey() string {
	return "Tuber-Token"
}

func refreshTokenCookieKey() string {
	return "TUBERTOKEN"
}

func accessTokenCtxKey() string {
	return "accessToken"
}

func refreshTokenCtxKey() string {
	return "refreshToken"
}

func (a *Authenticator) GetAccessToken(ctx context.Context) (string, error) {
	accessToken, ok := ctx.Value(accessTokenCtxKey()).(string)
	if ok && accessToken != "" {
		return accessToken, nil
	}

	refreshToken, ok := ctx.Value(refreshTokenCtxKey()).(string)
	if !ok || refreshToken == "" {
		fmt.Println("context refresh token" + refreshToken)
		return "", fmt.Errorf("no token found on request")
	}

	token, err := a.oauthConfig.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		fmt.Println(token)
		fmt.Println(err)
		return "", err
	}
	if token.AccessToken == "" {
		fmt.Println(token)
		return "", fmt.Errorf("cookie refresh token reissue returned blank access token")
	}

	return token.AccessToken, nil
}

func (a *Authenticator) TrySetAccessTokenContext(request *http.Request) (*http.Request, bool) {
	accessTokenHeaderValue := request.Header.Get(AccessTokenHeaderKey())
	if accessTokenHeaderValue != "" {
		return request, false
	}
	request = request.WithContext(context.WithValue(request.Context(), accessTokenCtxKey(), accessTokenHeaderValue))
	return request, true
}

func (a *Authenticator) RefreshTokenCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{Name: refreshTokenCookieKey(), Value: refreshToken, HttpOnly: true, Secure: true, Path: "/"}
}

func (a *Authenticator) TrySetRefreshTokenContext(request *http.Request) (*http.Request, bool) {
	for _, cookie := range request.Cookies() {
		if cookie.Name == refreshTokenCookieKey() && cookie.Value != "" {
			fmt.Println("refresh token cookie found: " + cookie.Value)
			return request.WithContext(context.WithValue(request.Context(), refreshTokenCtxKey(), cookie.Value)), true
		}
	}
	return request, false
}

func (a *Authenticator) GetRefreshTokenFromAuthToken(ctx context.Context, authorizationToken string) (string, error) {
	token, err := a.oauthConfig.Exchange(ctx, authorizationToken, oauth2.AccessTypeOffline)
	if err != nil {
		return "", err
	}
	if token.RefreshToken == "" {
		return "", fmt.Errorf("refresh token blank on auth code exchange")
	}
	return token.RefreshToken, nil
}

func (a *Authenticator) RefreshTokenConsentUrl() string {
	return a.oauthConfig.AuthCodeURL(a.oauthStateKey, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}
