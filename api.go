// A facebook graph api client in go.
// https://github.com/huandu/facebook/
// 
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

// A facebook graph api client in go. Simple but powerful.
// You can just use the Api() for most work.
//
// Sample 1: Read a user object without access token.
//     res, _ := facebook.Api("/huandu", facebook.GET, nil)
//     fmt.Println("my facebook id is", res["id"])
//
// Sample 2: Read a user object with a valid access token.
//     res, _ := facebook.Api("/me/feed", facebook.GET, facebook.Params{
//          "access_token": "a-valid-access-token",
//     })
//
//     // read my last feed
//     fmt.Println("my latest feed story is:", res.Get("data.0.story"))
//
// Sample 3: Use App and Session struct.
//     // create a global App var to hold your app id and secret.
//     var globalApp = facebook.New("your-app-id", "your-app-secret")
//
//     // here comes a client with a facebook signed request string in query string.
//     // creates a new session with signed request.
//     session, _ := globalApp.SessionFromSignedRequest(signedRequest)
//
//     // or, you just get a valid access token in other way.
//     // creates a session directly.
//     seesion := globalApp.Session(token)
//
//     // use session to send api request with your access token.
//     res, _ := session.Api("/me/feed", facebook.GET, nil)
//
//     // validate access token. err is nil if token is valid.
//     _, err := session.User()
//
// Sample 4: Read graph api response.
//     // define a facebook feed object.
//     type FacebookFeed struct {
//         Id, Story string
//         From *FacebookFeedFrom
//         CreatedTime string
//     }
//
//     type FacebookFeedFrom struct {
//         Name, Id string
//     }
//
//     // create a feed object direct from graph api result.
//     var feed FacebookFeed
//     res, _ := session.Api("/me/feed", facebook.GET, nil)
//     res.DecodeField("data.0", &feed) // then you can use feed.
//
// This library doesn't implement deprecated old-RESTFUL apis.
// I won't write code for them unless someone asks me to do so.
//
// This library doesn't include any HTTP integration.
// I will do it later.
package facebook

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
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
func (app *App) AppaccessToken() string {
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
func (app *App) ParseCode(code string) (token string, err error) {
    if code == "" {
        err = fmt.Errorf("code is empty")
        return
    }

    var response []byte
    session := &Session{}
    urlStr := getUrl("graph", "/oauth/access_token", nil)

    response, err = session.oauthRequest(urlStr, Params{
        "client_id":     app.AppId,
        "client_secret": app.AppSecret,
        "redirect_uri":  "",
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

    if token == "" {
        err = fmt.Errorf("facebook returns an empty token.")
        return
    }

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

    token, err = app.ParseCode(token)

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

// Makes a facebook graph api call.
//
// It's a wrapper of Session.Api(). Only works for graph api that doesn't require
// app id, app secret and access token. Can be called in multiple goroutines.
//
// If app id, app secret or access token is required in graph api, caller should use
// New() to create a new facebook session through App instead.
func Api(path string, method Method, params Params) (Result, error) {
    return defaultSession.Api(path, method, params)
}

// Makes a batch facebook graph api call.
//
// It's a wrapper of Session.Api(). Only works for graph api that doesn't require
// app id, app secret and access token. Can be called in multiple goroutines.
//
// If app id, app secret or access token is required in graph api, caller should use
// New() to create a new facebook session through App instead.
func BatchApi(accessToken string, params ...Params) ([]Result, error) {
    return defaultSession.graphBatch(accessToken, params...)
}

// Makes a facebook graph api call.
//
// Returns facebook graph api call result.
// If facebook returns error in response, returns error details in res and set err.
func (session *Session) Api(path string, method Method, params Params) (Result, error) {
    res, err := session.graph(path, method, params)

    if res != nil {
        return res, err
    }

    return nil, err
}

// Makes a batch call. Each params represent a single facebook graph api call.
// See https://developers.facebook.com/docs/reference/api/batch/ for batch call api details.
//
// Returns an array of batch call result on success.
func (session *Session) BatchApi(params ...Params) ([]Result, error) {
    return session.graphBatch(session.accessToken, params...)
}

// Gets current user id from access token.
//
// Returns error if access token is not set or invalid.
//
// It's a standard way to validate a facebook access token.
func (session *Session) User() (id string, err error) {
    if session.id != "" {
        id = session.id
        return
    }

    if session.accessToken == "" {
        err = fmt.Errorf("access token is not set.")
        return
    }

    var result Result
    result, err = session.Api("/me", GET, Params{"fields": "id"})

    if err != nil {
        return
    }

    err = result.DecodeField("id", &id)

    if err != nil {
        return
    }

    return
}

// Gets current access token.
func (session *Session) AccessToken() string {
    return session.accessToken
}

// Sets a new access token.
func (session *Session) SetAccessToken(token string) {
    if token != session.accessToken {
        session.id = ""
        session.accessToken = token
    }
}

// Gets associated App.
func (session *Session) App() *App {
    return session.app
}

func (session *Session) graph(path string, method Method, params Params) (res Result, err error) {
    var graphUrl string
    var response []byte

    if params == nil {
        params = Params{}
    }

    // overwrite method as we always use post
    params["method"] = method

    if session.isVideoPost(path, method) {
        graphUrl = getUrl("graph_video", path, nil)
    } else {
        graphUrl = getUrl("graph", path, nil)
    }

    response, err = session.oauthRequest(graphUrl, params)

    // cannot get response from remote server
    if err != nil {
        return
    }

    err = json.Unmarshal(response, &res)

    if err != nil {
        res = nil
        err = fmt.Errorf("cannot format facebook response. %v", err)
        return
    }

    // facebook returns an error
    if _, ok := res["error"]; ok {
        err = fmt.Errorf("facebook returns an error")
    }

    return
}

func (session *Session) graphBatch(accessToken string, params ...Params) (res []Result, err error) {
    var batchParams = Params{"access_token": accessToken}
    var batchJson []byte
    var response []byte

    // encode all params to a json array.
    batchJson, err = json.Marshal(params)

    if err != nil {
        return
    }

    batchParams["batch"] = string(batchJson)

    graphUrl := getUrl("graph", "", nil)
    response, err = session.oauthRequest(graphUrl, batchParams)

    if err != nil {
        return
    }

    err = json.Unmarshal(response, &res)

    if err != nil {
        res = nil
        err = fmt.Errorf("cannot format facebook batch response. %v", err)
        return
    }

    return
}

func (session *Session) oauthRequest(url string, params Params) ([]byte, error) {
    if _, ok := params["access_token"]; !ok && session.accessToken != "" {
        params["access_token"] = session.accessToken
    }

    return session.makeRequest(url, params)
}

func (session *Session) makeRequest(url string, params Params) ([]byte, error) {
    buf := &bytes.Buffer{}
    buf.WriteString(params.Encode())
    response, err := http.Post(url, "application/x-www-form-urlencoded", buf)

    if err != nil {
        return nil, fmt.Errorf("cannot reach facebook server. %v", err)
    }

    defer response.Body.Close()

    if response.StatusCode >= 300 {
        return nil, fmt.Errorf("facebook server response an HTTP error. code: %v, body: %s",
            response.StatusCode, string(buf.Bytes()))
    }

    buf = &bytes.Buffer{}
    _, err = io.Copy(buf, response.Body)

    if err != nil {
        return nil, fmt.Errorf("cannot read facebook response. %v", err)
    }

    return buf.Bytes(), nil
}

func (session *Session) isVideoPost(path string, method Method) bool {
    return method == POST && regexpIsVideoPost.MatchString(path)
}

func getUrl(name, path string, params Params) string {
    offset := 0

    if path != "" && path[0] == '/' {
        offset = 1
    }

    buf := &bytes.Buffer{}
    buf.WriteString(domainMap[name])
    buf.WriteString(string(path[offset:]))

    if params != nil {
        buf.WriteRune('?')
        buf.WriteString(params.Encode())
    }

    return buf.String()
}

func decodeBase64URLEncodingString(data string) ([]byte, error) {
    buf := bytes.NewBufferString(data)

    // go's URLEncoding implementation requires base64 padding.
    if m := len(data) % 4; m != 0 {
        buf.WriteString(strings.Repeat("=", 4-m))
    }

    reader := base64.NewDecoder(base64.URLEncoding, buf)
    output := &bytes.Buffer{}
    _, err := io.Copy(output, reader)

    if err != nil {
        return nil, err
    }

    return output.Bytes(), nil
}
