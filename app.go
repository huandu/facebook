// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Creates a new App and sets app id and secret.
func New(appId, appSecret string) *App {
	return &App{
		AppId:     appId,
		AppSecret: appSecret,
	}
}

// Gets application access token, useful for gathering public information about users and applications.
func (app *App) AppAccessToken() string {
	return app.AppId + "|" + app.AppSecret
}

// Parses signed request.
func (app *App) ParseSignedRequest(signedRequest string) (res Result, err error) {
	strs := strings.SplitN(signedRequest, ".", 2)

	if len(strs) != 2 {
		err = fmt.Errorf("invalid signed request format.")
		return
	}

	sig, e1 := decodeBase64URLEncodingString(strs[0])

	if e1 != nil {
		err = fmt.Errorf("cannot decode signed request sig. error is %v.", e1)
		return
	}

	payload, e2 := decodeBase64URLEncodingString(strs[1])

	if e2 != nil {
		err = fmt.Errorf("cannot decode signed request payload. error is %v.", e2)
		return
	}

	err = json.Unmarshal(payload, &res)

	if err != nil {
		err = fmt.Errorf("signed request payload is not a valid json string. error is %v.", err)
		return
	}

	var hashMethod string
	err = res.DecodeField("algorithm", &hashMethod)

	if err != nil {
		err = fmt.Errorf("signed request payload doesn't contains a valid 'algorithm' field.")
		return
	}

	hashMethod = strings.ToUpper(hashMethod)

	if hashMethod != "HMAC-SHA256" {
		err = fmt.Errorf("signed request payload uses an unknown HMAC method. expect 'HMAC-SHA256'. actual '%v'.", hashMethod)
		return
	}

	hash := hmac.New(sha256.New, []byte(app.AppSecret))
	hash.Write([]byte(strs[1])) // note: here uses the payload base64 string, not decoded bytes
	expectedSig := hash.Sum(nil)

	if bytes.Compare(sig, expectedSig) != 0 {
		err = fmt.Errorf("bad signed request signiture.")
		return
	}

	return
}

// Parses facebook code to a valid access token.
//
// In facebook PHP SDK, there is a CSRF state to avoid attack.
// That state is not checked in this library.
// Caller is responsible to store and check state if possible.
//
// Returns a valid access token exchanged from a code.
func (app *App) ParseCode(code string, redirectUri string) (token string, err error) {
	if code == "" {
		err = fmt.Errorf("code is empty")
		return
	}

	res := &Result{}
	var response []byte
	session := &Session{}
	urlStr := getUrl("graph", "/oauth/access_token", nil)

	response, err = session.oauthRequest(urlStr, Params{
		"client_id":     app.AppId,
		"client_secret": app.AppSecret,
		"redirect_uri":  redirectUri,
		"code":          code,
	})

	if err != nil {
		return
	}

	var values url.Values
	values, err = url.ParseQuery(string(response))

	if err != nil {
		err = fmt.Errorf("cannot parse facebook response. error is %v.", err)
		return
	}

	token = values.Get("access_token")

	if token != "" {
		return
	}

	err = json.Unmarshal(response, res)

	if err != nil {
		err = fmt.Errorf("facebook returns an empty token.")
		return
	}
	err = res.Err()

	return
}

// Creates a session based on current App setting.
func (app *App) Session(accessToken string) *Session {
	return &Session{
		accessToken: accessToken,
		app:         app,
	}
}

// Creates a session from a signed request.
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
			accessToken: token,
			app:         app,
			id:          id,
		}
		return
	}

	// cannot get "oauth_token"? try to get "code".
	err = res.DecodeField("code", &token)

	if err != nil {
		// no code? no way to continue.
		err = fmt.Errorf("cannot find 'oauth_token' and 'code'. no way to continue.")
		return
	}

	token, err = app.ParseCode(token, "")

	if err != nil {
		return
	}

	session = &Session{
		accessToken: token,
		app:         app,
		id:          id,
	}
	return
}
