// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2014, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"io"
	"net/http"
)

// Holds facebook application information.
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
}

// An interface to send http request.
// This interface is designed to be compatible with type `*http.Client`.
type HttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
	Get(url string) (resp *http.Response, err error)
	Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error)
}

// Holds a facebook session with an access token.
// Session should be created by App.Session or App.SessionFromSignedRequest.
type Session struct {
	HttpClient HttpClient
	Version    string // facebook versioning.

	accessToken string // facebook access token. can be empty.
	app         *App
	id          string

	enableAppsecretProof bool   // add "appsecret_proof" parameter in every facebook API call.
	appsecretProof       string // pre-calculated "appsecret_proof" value.
}

// API HTTP method.
// Can be GET, POST or DELETE.
type Method string

// API params.
//
// For general uses, just use Params as a ordinary map.
//
// For advanced uses, use MakeParams to create Params from any struct.
type Params map[string]interface{}

// Facebook API call result.
type Result map[string]interface{}

// Represents facebook API call result with paging information.
type PagingResult struct {
	session  *Session
	paging   pagingData
	previous string
	next     string
}

// Facebook API error.
type Error struct {
	Message      string
	Type         string
	Code         int
	ErrorSubcode int // subcode for authentication related errors.
}

// Binary data.
type binaryData struct {
	Filename string    // filename used in multipart form writer.
	Source   io.Reader // file data source.
}

// Binary file.
type binaryFile struct {
	Filename string // filename used in multipart form writer.
	Path     string // path to file. must be readable.
}
