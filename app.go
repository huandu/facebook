// A facebook graph api client in go.
// https://github.com/runnart/facebook/
//
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/runnart/facebook/blob/master/LICENSE

package facebook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// App holds facebook application information.
type App struct {
	// Facebook app id
	AppId string

	// Facebook app secret
	AppSecret string

	// Facebook app redirect URI in the app's configuration.
	RedirectUri string

	// Enable appsecret proof in every API call to facebook.
	// Facebook document: https://developers.facebook.com/docs/graph-api/securing-requests
	EnableAppsecretProof bool

	// The session to send request when parsing tokens or code.
	// If it's not set, default session will be used.
	session *Session
}

// New creates a new App and sets app id and secret.
func New(appID, appSecret, redirectURI string) *App {
	return &App{
		AppId:       appID,
		AppSecret:   appSecret,
		RedirectUri: redirectURI,
		session:     defaultSession,
	}
}

// AppAccessToken gets application access token, useful for gathering public information about users and applications.
func (app *App) AppAccessToken() string {
	return app.AppId + "|" + app.AppSecret
}

// SetSession is used to overwrite the default session used by the app
func (app *App) SetSession(s *Session) {
	app.session = s
}

// ParseSignedRequest parses signed request.
func (app *App) ParseSignedRequest(signedRequest string) (res Result, err error) {
	strs := strings.SplitN(signedRequest, ".", 2)

	if len(strs) != 2 {
		err = fmt.Errorf("facebook: invalid signed request format")
		return
	}

	sig, e1 := base64.RawURLEncoding.DecodeString(strs[0])

	if e1 != nil {
		err = fmt.Errorf("facebook: fail to decode signed request sig with error %w", e1)
		return
	}

	payload, e2 := base64.RawURLEncoding.DecodeString(strs[1])

	if e2 != nil {
		err = fmt.Errorf("facebook: fail to decode signed request payload with error is %w", e2)
		return
	}

	err = json.Unmarshal(payload, &res)

	if err != nil {
		err = fmt.Errorf("facebook: signed request payload is not a valid json string with error %w", err)
		return
	}

	var hashMethod string
	err = res.DecodeField("algorithm", &hashMethod)

	if err != nil {
		err = fmt.Errorf("facebook: signed request payload doesn't contains a valid 'algorithm' field")
		return
	}

	hashMethod = strings.ToUpper(hashMethod)

	if hashMethod != "HMAC-SHA256" {
		err = fmt.Errorf("facebook: signed request payload uses an unknown HMAC method; expect 'HMAC-SHA256' but actual is '%v'", hashMethod)
		return
	}

	hash := hmac.New(sha256.New, []byte(app.AppSecret))
	hash.Write([]byte(strs[1])) // note: here uses the payload base64 string, not decoded bytes
	expectedSig := hash.Sum(nil)

	if !hmac.Equal(sig, expectedSig) {
		err = fmt.Errorf("facebook: bad signed request signiture")
		return
	}

	return
}

// ParseCode redeems code for a valid access token.
// It's a shorthand call to ParseCodeInfo(code, "").
//
// In facebook PHP SDK, there is a CSRF state to avoid attack.
// That state is not checked in this library.
// Caller is responsible to store and check state if possible.
func (app *App) ParseCode(code string) (token string, err error) {
	token, _, _, err = app.ParseCodeInfo(code, "")
	return
}

// ParseCodeInfo redeems code for access token and returns extra information.
// The machineId is optional.
//
// See https://developers.facebook.com/docs/facebook-login/access-tokens#extending
func (app *App) ParseCodeInfo(code, machineID string) (token string, expires int, newMachineID string, err error) {
	if code == "" {
		err = fmt.Errorf("facebook: code is empty")
		return
	}

	var res Result
	res, err = app.session.sendOauthRequest("/oauth/access_token", Params{
		"client_id":     app.AppId,
		"redirect_uri":  app.RedirectUri,
		"client_secret": app.AppSecret,
		"code":          code,
	})

	if err != nil {
		err = fmt.Errorf("facebook: fail to parse facebook response with error %w", err)
		return
	}

	err = res.DecodeField("access_token", &token)

	if err != nil {
		return
	}

	expiresKey := "expires_in"

	if _, ok := res["expires"]; ok {
		expiresKey = "expires"
	}

	if _, ok := res[expiresKey]; ok {
		err = res.DecodeField(expiresKey, &expires)

		if err != nil {
			return
		}
	}

	if _, ok := res["machine_id"]; ok {
		err = res.DecodeField("machine_id", &newMachineID)
	}

	return
}

// ExchangeToken exchanges a short-lived access token to a long-lived access token.
// Return new access token and its expires time.
func (app *App) ExchangeToken(accessToken string) (token string, expires int, err error) {
	if accessToken == "" {
		err = fmt.Errorf("short lived accessToken is empty")
		return
	}

	var res Result
	res, err = app.session.sendOauthRequest("/oauth/access_token", Params{
		"grant_type":        "fb_exchange_token",
		"client_id":         app.AppId,
		"client_secret":     app.AppSecret,
		"fb_exchange_token": accessToken,
	})

	if err != nil {
		err = fmt.Errorf("fail to parse facebook response with error %w", err)
		return
	}

	err = res.DecodeField("access_token", &token)

	if err != nil {
		return
	}

	expiresKey := "expires_in"

	if _, ok := res["expires"]; ok {
		expiresKey = "expires"
	}

	if _, ok := res[expiresKey]; ok {
		err = res.DecodeField(expiresKey, &expires)
	}

	return
}

// GetCode gets code from a long lived access token.
// Return the code retrieved from facebook.
func (app *App) GetCode(accessToken string) (code string, err error) {
	if accessToken == "" {
		err = fmt.Errorf("facebook: long lived accessToken is empty")
		return
	}

	var res Result
	res, err = app.session.sendOauthRequest("/oauth/client_code", Params{
		"client_id":     app.AppId,
		"client_secret": app.AppSecret,
		"redirect_uri":  app.RedirectUri,
		"access_token":  accessToken,
	})

	if err != nil {
		err = fmt.Errorf("facebook: fail to get code from facebook with error %w", err)
		return
	}

	err = res.DecodeField("code", &code)
	return
}

// Session creates a session based on current App setting.
func (app *App) Session(accessToken string) *Session {
	return &Session{
		accessToken:          accessToken,
		app:                  app,
		enableAppsecretProof: app.EnableAppsecretProof,
	}
}

// SessionFromSignedRequest creates a session from a signed request.
// If signed request contains a code, it will automatically use this code
// to exchange a valid access token.
func (app *App) SessionFromSignedRequest(signedRequest string) (session *Session, err error) {
	var res Result

	res, err = app.ParseSignedRequest(signedRequest)

	if err != nil {
		return
	}

	var id, token string

	res.DecodeField("user_id", &id) // it's ok without user id.
	err = res.DecodeField("oauth_token", &token)

	if err == nil {
		session = &Session{
			accessToken:          token,
			app:                  app,
			id:                   id,
			enableAppsecretProof: app.EnableAppsecretProof,
		}
		return
	}

	// cannot get "oauth_token"? try to get "code".
	err = res.DecodeField("code", &token)

	if err != nil {
		// no code? no way to continue.
		err = fmt.Errorf("facebook: cannot find 'oauth_token' and 'code'; unable to continue")
		return
	}

	token, err = app.ParseCode(token)

	if err != nil {
		return
	}

	session = &Session{
		accessToken:          token,
		app:                  app,
		id:                   id,
		enableAppsecretProof: app.EnableAppsecretProof,
	}
	return
}
