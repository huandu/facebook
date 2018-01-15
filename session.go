// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2015, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

// Graph API debug mode values.
const (
	DEBUG_OFF DebugMode = "" // turn off debug.

	DEBUG_ALL     DebugMode = "all"
	DEBUG_INFO    DebugMode = "info"
	DEBUG_WARNING DebugMode = "warning"
)

var (
	// Maps aliases to Facebook domains.
	// Copied from Facebook PHP SDK.
	domainMap = map[string]string{
		"api":         "https://api.facebook.com/",
		"api_video":   "https://api-video.facebook.com/",
		"api_read":    "https://api-read.facebook.com/",
		"graph":       "https://graph.facebook.com/",
		"graph_video": "https://graph-video.facebook.com/",
		"www":         "https://www.facebook.com/",
	}

	// checks whether it's a video post.
	regexpIsVideoPost = regexp.MustCompile(`\/videos$`)
)

// Session holds a facebook session with an access token.
// Session should be created by App.Session or App.SessionFromSignedRequest.
type Session struct {
	HttpClient *http.Client
	Version    string // facebook versioning.

	accessToken string // facebook access token. can be empty.
	app         *App
	id          string

	enableAppsecretProof bool   // add "appsecret_proof" parameter in every facebook API call.
	appsecretProof       string // pre-calculated "appsecret_proof" value.

	debug DebugMode // using facebook debugging api in every request.

	context context.Context // Session context.
}

// Api makes a facebook graph api call.
//
// If session access token is set, "access_token" in params will be set to the token value.
//
// Returns facebook graph api call result.
// If facebook returns error in response, returns error details in res and set err.
func (session *Session) Api(path string, method Method, params Params) (Result, error) {
	return session.graph(path, method, params)
}

// Get is a short hand of Api(path, GET, params).
func (session *Session) Get(path string, params Params) (Result, error) {
	return session.Api(path, http.MethodGet, params)
}

// Post is a short hand of Api(path, POST, params).
func (session *Session) Post(path string, params Params) (Result, error) {
	return session.Api(path, http.MethodPost, params)
}

// Delete is a short hand of Api(path, DELETE, params).
func (session *Session) Delete(path string, params Params) (Result, error) {
	return session.Api(path, http.MethodDelete, params)
}

// Put is a short hand of Api(path, PUT, params).
func (session *Session) Put(path string, params Params) (Result, error) {
	return session.Api(path, http.MethodPut, params)
}

// BatchApi makes a batch call. Each params represent a single facebook graph api call.
//
// BatchApi supports most kinds of batch calls defines in facebook batch api document,
// except uploading binary data. Use Batch to upload binary data.
//
// If session access token is set, the token will be used in batch api call.
//
// Returns an array of batch call result on success.
//
// Facebook document: https://developers.facebook.com/docs/graph-api/making-multiple-requests
func (session *Session) BatchApi(params ...Params) ([]Result, error) {
	return session.Batch(nil, params...)
}

// Batch makes a batch facebook graph api call.
// Batch is designed for more advanced usage including uploading binary files.
//
// If session access token is set, "access_token" in batchParams will be set to the token value.
//
// Facebook document: https://developers.facebook.com/docs/graph-api/making-multiple-requests
func (session *Session) Batch(batchParams Params, params ...Params) ([]Result, error) {
	return session.graphBatch(batchParams, params...)
}

// Request makes an arbitrary HTTP request.
// It expects server responses a facebook Graph API response.
//     request, _ := http.NewRequest("https://graph.facebook.com/538744468", "GET", nil)
//     res, err := session.Request(request)
//     fmt.Println(res["gender"])  // get "male"
func (session *Session) Request(request *http.Request) (res Result, err error) {
	var response *http.Response
	var data []byte

	response, data, err = session.sendRequest(request)

	if err != nil {
		return
	}

	res, err = MakeResult(data)
	session.addDebugInfo(res, response)
	session.addUsageInfo(res, response)

	if res != nil {
		err = res.Err()
	}

	return
}

// User gets current user id from access token.
//
// Returns error if access token is not set or invalid.
//
// It's a standard way to validate a facebook access token.
func (session *Session) User() (id string, err error) {
	if session.id != "" {
		id = session.id
		return
	}

	client := &http.Client{}
	if session.accessToken == "" && session.HttpClient == client {
		err = fmt.Errorf("access token is not set")
		return
	}

	var result Result
	result, err = session.Api("/me", http.MethodGet, Params{"fields": "id"})

	if err != nil {
		return
	}

	err = result.DecodeField("id", &id)

	if err != nil {
		return
	}

	return
}

// Validate validates Session access token.
// Returns nil if access token is valid.
func (session *Session) Validate() (err error) {
	client := &http.Client{}
	if session.accessToken == "" && session.HttpClient == client {
		err = fmt.Errorf("access token is not set")
		return
	}

	var result Result
	result, err = session.Api("/me", http.MethodGet, Params{"fields": "id"})

	if err != nil {
		return
	}

	if f := result.Get("id"); f == nil {
		err = fmt.Errorf("invalid access token")
		return
	}

	return
}

// Inspect Session access token.
// Returns JSON array containing data about the inspected token.
// See https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow/v2.2#checktoken
func (session *Session) Inspect() (result Result, err error) {
	client := &http.Client{}
	if session.accessToken == "" && session.HttpClient == client {
		err = fmt.Errorf("access token is not set")
		return
	}

	if session.app == nil {
		err = fmt.Errorf("cannot inspect access token without binding an app")
		return
	}

	appAccessToken := session.app.AppAccessToken()

	if appAccessToken == "" {
		err = fmt.Errorf("app access token is not set")
		return
	}

	result, err = session.Api("/debug_token", http.MethodGet, Params{
		"input_token":  session.accessToken,
		"access_token": appAccessToken,
	})

	if err != nil {
		return
	}

	// facebook stores everything, including error, inside result["data"].
	// make sure that result["data"] exists and doesn't contain error.
	if _, ok := result["data"]; !ok {
		err = fmt.Errorf("facebook inspect api returns unexpected result")
		return
	}

	var data Result
	result.DecodeField("data", &data)
	result = data
	err = result.Err()
	return
}

// AccessToken gets current access token.
func (session *Session) AccessToken() string {
	return session.accessToken
}

// Sets a new access token.
func (session *Session) SetAccessToken(token string) {
	if token != session.accessToken {
		session.id = ""
		session.accessToken = token
		session.appsecretProof = ""
	}
}

// AppsecretProof checks appsecret proof is enabled or not.
func (session *Session) AppsecretProof() string {
	if !session.enableAppsecretProof {
		return ""
	}

	if session.accessToken == "" || session.app == nil {
		return ""
	}

	if session.appsecretProof == "" {
		hash := hmac.New(sha256.New, []byte(session.app.AppSecret))
		hash.Write([]byte(session.accessToken))
		session.appsecretProof = hex.EncodeToString(hash.Sum(nil))
	}

	return session.appsecretProof
}

// EnableAppsecretProof enables or disable appsecret proof status.
// Returns error if there is no App associated with this Session.
func (session *Session) EnableAppsecretProof(enabled bool) error {
	if session.app == nil {
		return fmt.Errorf("cannot change appsecret proof status without an associated App")
	}

	if session.enableAppsecretProof != enabled {
		session.enableAppsecretProof = enabled

		// reset pre-calculated proof here to give caller a way to do so in some rare case,
		// e.g. associated app's secret is changed.
		session.appsecretProof = ""
	}

	return nil
}

// App gets associated App.
func (session *Session) App() *App {
	return session.app
}

// Debug returns current debug mode.
func (session *Session) Debug() DebugMode {
	if session.debug != DEBUG_OFF {
		return session.debug
	}

	return Debug
}

// SetDebug updates per session debug mode and returns old mode.
// If per session debug mode is DEBUG_OFF, session will use global
// Debug mode.
func (session *Session) SetDebug(debug DebugMode) DebugMode {
	old := session.debug
	session.debug = debug
	return old
}

func (session *Session) graph(path string, method Method, params Params) (res Result, err error) {
	var graphURL string

	if params == nil {
		params = Params{}
	}

	// always format as json.
	params["format"] = "json"

	// overwrite method as we always use post
	params["method"] = method

	// get graph api url.
	if session.isVideoPost(path, method) {
		graphURL = session.getURL("graph_video", path, nil)
	} else {
		graphURL = session.getURL("graph", path, nil)
	}

	var response *http.Response
	response, err = session.sendPostRequest(graphURL, params, &res)

	if response != nil {
		session.addDebugInfo(res, response)
		session.addUsageInfo(res, response)
	}

	if res != nil {
		err = res.Err()
	}

	return
}

func (session *Session) graphBatch(batchParams Params, params ...Params) ([]Result, error) {
	if batchParams == nil {
		batchParams = Params{}
	}

	batchParams["batch"] = params

	var res []Result
	graphUrl := session.getURL("graph", "", nil)
	_, err := session.sendPostRequest(graphUrl, batchParams, &res)
	return res, err
}

func (session *Session) prepareParams(params Params) {
	if _, ok := params["access_token"]; !ok && session.accessToken != "" {
		params["access_token"] = session.accessToken
	}

	if session.enableAppsecretProof && session.accessToken != "" && session.app != nil {
		params["appsecret_proof"] = session.AppsecretProof()
	}

	debug := session.Debug()

	if debug != DEBUG_OFF {
		params["debug"] = debug
	}
}

func (session *Session) sendGetRequest(uri string, res interface{}) (*http.Response, error) {
	request, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		return nil, err
	}

	response, data, err := session.sendRequest(request)

	if err != nil {
		return response, err
	}

	err = makeResult(data, res)
	return response, err
}

func (session *Session) sendPostRequest(uri string, params Params, res interface{}) (*http.Response, error) {
	session.prepareParams(params)

	buf := &bytes.Buffer{}
	mime, err := params.Encode(buf)

	if err != nil {
		return nil, fmt.Errorf("cannot encode POST params. %v", err)
	}

	var request *http.Request
	request, err = http.NewRequest(http.MethodPost, uri, buf)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", mime)
	response, data, err := session.sendRequest(request)

	if err != nil {
		return response, err
	}

	err = makeResult(data, res)
	return response, err
}

func (session *Session) sendOauthRequest(uri string, params Params) (Result, error) {
	urlStr := session.getURL("graph", uri, nil)
	buf := &bytes.Buffer{}
	mime, err := params.Encode(buf)

	if err != nil {
		return nil, fmt.Errorf("cannot encode POST params. %v", err)
	}

	var request *http.Request

	request, err = http.NewRequest(http.MethodPost, urlStr, buf)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", mime)
	_, data, err := session.sendRequest(request)

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty response from facebook")
	}

	// facebook may return a query string.
	if 'a' <= data[0] && data[0] <= 'z' {
		query, err := url.ParseQuery(string(data))

		if err != nil {
			return nil, err
		}

		// convert a query to Result.
		res := Result{}

		for k := range query {
			// json.Number is an alias of string and can be decoded as a string or number.
			// therefore, it's safe to convert all query values to this type for all purpose.
			res[k] = json.Number(query.Get(k))
		}

		return res, nil
	}

	res, err := MakeResult(data)
	return res, err
}

func (session *Session) sendRequest(request *http.Request) (response *http.Response, data []byte, err error) {
	if session.context != nil {
		request = request.WithContext(session.context)
	}

	response, err = session.HttpClient.Do(request)
	if err != nil {
		err = fmt.Errorf("cannot reach facebook server. %v", err)
		return
	}

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, response.Body)
	response.Body.Close()

	if err != nil {
		err = fmt.Errorf("cannot read facebook response. %v", err)
	}

	data = buf.Bytes()
	return
}

func (session *Session) isVideoPost(path string, method Method) bool {
	return method == http.MethodPost && regexpIsVideoPost.MatchString(path)
}

func (session *Session) getURL(name, path string, params Params) string {
	offset := 0

	if path != "" && path[0] == '/' {
		offset = 1
	}

	buf := &bytes.Buffer{}
	buf.WriteString(domainMap[name])

	// facebook versioning.
	if session.Version == "" {
		if Version != "" {
			buf.WriteString(Version)
			buf.WriteRune('/')
		}
	} else {
		buf.WriteString(session.Version)
		buf.WriteRune('/')
	}

	buf.WriteString(string(path[offset:]))

	if params != nil {
		buf.WriteRune('?')
		params.Encode(buf)
	}

	return buf.String()
}

func (session *Session) addDebugInfo(res Result, response *http.Response) Result {
	if session.Debug() == DEBUG_OFF || res == nil || response == nil {
		return res
	}

	debugInfo := make(map[string]interface{})

	// save debug information in result directly.
	res.DecodeField(debugInfoKey, &debugInfo)
	debugInfo[debugProtoKey] = response.Proto
	debugInfo[debugHeaderKey] = response.Header

	res[debugInfoKey] = debugInfo
	return res
}

func (session *Session) addUsageInfo(res Result, response *http.Response) Result {
	if res == nil || response == nil {
		return res
	}

	var usageInfo UsageInfo

	if usage := response.Header.Get("X-App-Usage"); usage != "" {
		json.Unmarshal([]byte(usage), &usageInfo.App)
	}

	if usage := response.Header.Get("X-Page-Usage"); usage != "" {
		json.Unmarshal([]byte(usage), &usageInfo.Page)
	}

	res[usageInfoKey] = &usageInfo
	return res
}

// Context returns the session's context.
// To change the context, use `Session#WithContext`.
//
// The returned context is always non-nil; it defaults to the background context.
// For outgoing Facebook API requests, the context controls timeout/deadline and cancelation.
func (session *Session) Context() context.Context {
	if session.context != nil {
		return session.context
	}

	return context.Background()
}

// WithContext returns a shallow copy of session with its context changed to ctx.
// The provided ctx must be non-nil.
func (session *Session) WithContext(ctx context.Context) *Session {
	s := *session
	s.context = ctx
	return &s
}
