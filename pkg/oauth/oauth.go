package oauth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
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
	return "TUBERRTOKEN"
}

func accessTokenCookieKey() string {
	return "TUBERATOKEN"
}

func accessTokenExpirationCookieKey() string {
	return "TUBERATOKENEXP"
}

type oauthCtxKey string

var accessTokenCtxKey oauthCtxKey = "accessToken"
var refreshTokenCtxKey oauthCtxKey = "refreshToken"
var accessTokenExpirationCtxKey oauthCtxKey = "accessTokenExpiration"
var expirationTimeFormat = time.RFC3339

func GetAccessToken(ctx context.Context) (string, error) {
	accessToken, ok := ctx.Value(accessTokenCtxKey).(string)
	if !ok || accessToken == "" {
		return "", fmt.Errorf("no access token found, try /login")
	}

	return accessToken, nil
}

func (a *Authenticator) TrySetHeaderAuthContext(request *http.Request) (*http.Request, bool) {
	accessTokenHeaderValue := request.Header.Get(AccessTokenHeaderKey())
	if accessTokenHeaderValue == "" {
		return request, false
	}
	request = request.WithContext(context.WithValue(request.Context(), accessTokenCtxKey, accessTokenHeaderValue))
	return request, true
}

// TrySetCookieAuthContext - gqlgen gives us the context in the resolver, but does not expose any way to alter it midflight
// specifically, the generated handler functions hand off a context to the resolver, and then hand the same context to the marshaller.
// so we're doing everything before it gets there.
// this means that even requests that don't NEED an access token will have it refreshed when expired.
// Given this is only every 30 minutes per user, it's not the worst.
// But given we don't plan to be harsh on view access, it's unavoidable major overkill.
func (a *Authenticator) TrySetCookieAuthContext(w http.ResponseWriter, r *http.Request, sc *securecookie.SecureCookie) (http.ResponseWriter, *http.Request, bool, error) {
	var refreshToken string
	var accessToken string
	var accessTokenExpiration string
	for _, cookie := range r.Cookies() {
		if cookie.Name == refreshTokenCookieKey() && cookie.Value != "" {
			err := sc.Decode(refreshTokenCookieKey(), cookie.Value, &refreshToken)
			if err != nil {
				return w, r, false, err
			}
			continue
		}
		if cookie.Name == accessTokenCookieKey() && cookie.Value != "" {
			err := sc.Decode(refreshTokenCookieKey(), cookie.Value, &accessToken)
			if err != nil {
				return w, r, false, err
			}
			continue
		}
		if cookie.Name == accessTokenExpirationCookieKey() && cookie.Value != "" {
			accessTokenExpiration = cookie.Value
		}
	}

	if refreshToken == "" {
		return w, r, false, nil
	}

	var refreshed bool
	expiration, expParseErr := time.Parse(expirationTimeFormat, accessTokenExpiration)
	// 1 min of space just in case it's about to expire and WOULD before a can-i check further in the request
	// this also implies we'll never run an entire deploy impersonating, using a web access token, unless we were to manually refresh it beforehand
	if expParseErr != nil || expiration.Before(time.Now().Add(time.Minute)) {
		token, refreshErr := a.oauthConfig.TokenSource(r.Context(), &oauth2.Token{RefreshToken: refreshToken}).Token()
		if refreshErr != nil {
			return w, r, false, fmt.Errorf("refresh expired cookie error: %v", refreshErr)
		}
		if token.AccessToken == "" {
			return w, r, false, fmt.Errorf("cookie refresh token reissue returned blank access token")
		}
		accessToken = token.AccessToken
		accessTokenExpiration = token.Expiry.Format(expirationTimeFormat)
		refreshed = true
	}

	r = r.WithContext(context.WithValue(r.Context(), refreshTokenCtxKey, refreshToken))
	r = r.WithContext(context.WithValue(r.Context(), accessTokenCtxKey, accessToken))
	r = r.WithContext(context.WithValue(r.Context(), accessTokenExpirationCtxKey, accessTokenExpiration))

	if refreshed {
		encodedRefresh, err := sc.Encode(refreshTokenCookieKey(), refreshToken)
		if err != nil {
			return w, r, false, fmt.Errorf("encode refresh token cookie error: %v", err)
		}

		encodedAccess, err := sc.Encode(refreshTokenCookieKey(), accessToken)
		if err != nil {
			return w, r, false, fmt.Errorf("encode access token cookie error: %v", err)
		}

		cookies := []*http.Cookie{
			{Name: refreshTokenCookieKey(), Value: encodedRefresh, HttpOnly: true, Secure: true, Path: "/"},
			{Name: accessTokenCookieKey(), Value: encodedAccess, HttpOnly: true, Secure: true, Path: "/"},
			{Name: accessTokenExpirationCookieKey(), Value: accessTokenExpiration, HttpOnly: true, Secure: true, Path: "/"},
		}
		for _, cookie := range cookies {
			http.SetCookie(w, cookie)
		}
	}

	return w, r, true, nil
}

func (a *Authenticator) GetTokenCookiesFromAuthToken(ctx context.Context, authorizationToken string, sc *securecookie.SecureCookie) ([]*http.Cookie, error) {
	token, err := a.oauthConfig.Exchange(ctx, authorizationToken, oauth2.AccessTypeOffline)
	if err != nil {
		return nil, err
	}
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token blank on auth code exchange")
	}

	encodedRefresh, err := sc.Encode(refreshTokenCookieKey(), token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("encode refresh token cookie error: %v", err)
	}

	encodedAccess, err := sc.Encode(refreshTokenCookieKey(), token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("encode access token cookie error: %v", err)
	}

	return []*http.Cookie{
		{Name: refreshTokenCookieKey(), Value: encodedRefresh, HttpOnly: true, Secure: true, Path: "/"},
		{Name: accessTokenCookieKey(), Value: encodedAccess, HttpOnly: true, Secure: true, Path: "/"},
		{Name: accessTokenExpirationCookieKey(), Value: token.Expiry.String(), HttpOnly: true, Secure: true, Path: "/"},
	}, nil
}

func (a *Authenticator) RefreshTokenConsentUrl() string {
	return a.oauthConfig.AuthCodeURL(a.oauthStateKey, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}
