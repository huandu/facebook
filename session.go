// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2014, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Makes a facebook graph api call.
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
	return session.Api(path, GET, params)
}

// Post is a short hand of Api(path, POST, params).
func (session *Session) Post(path string, params Params) (Result, error) {
	return session.Api(path, POST, params)
}

// Delete is a short hand of Api(path, DELETE, params).
func (session *Session) Delete(path string, params Params) (Result, error) {
	return session.Api(path, DELETE, params)
}

// Put is a short hand of Api(path, PUT, params).
func (session *Session) Put(path string, params Params) (Result, error) {
	return session.Api(path, PUT, params)
}

// Makes a batch call. Each params represent a single facebook graph api call.
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

// Makes a batch facebook graph api call.
// Batch is designed for more advanced usage including uploading binary files.
//
// If session access token is set, "access_token" in batchParams will be set to the token value.
//
// Facebook document: https://developers.facebook.com/docs/graph-api/making-multiple-requests
func (session *Session) Batch(batchParams Params, params ...Params) ([]Result, error) {
	return session.graphBatch(batchParams, params...)
}

// Makes a FQL query.
// Returns a slice of Result. If there is no query result, the result is nil.
//
// Facebook document: https://developers.facebook.com/docs/technical-guides/fql#query
func (session *Session) FQL(query string) ([]Result, error) {
	res, err := session.graphFQL(Params{
		"q": query,
	})

	if err != nil {
		return nil, err
	}

	// query result is stored in "data" field.
	var data []Result
	err = res.DecodeField("data", &data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// Makes a multi FQL query.
// Returns a parsed Result. The key is the multi query key, and the value is the query result.
//
// Here is a multi-query sample.
//
//     res, _ := session.MultiFQL(Params{
//         "query1": "SELECT name FROM user WHERE uid = me()",
//         "query2": "SELECT uid1, uid2 FROM friend WHERE uid1 = me()",
//     })
//
//     // Get query results from response.
//     var query1, query2 []Result
//     res.DecodeField("query1", &query1)
//     res.DecodeField("query2", &query2)
//
// Facebook document: https://developers.facebook.com/docs/technical-guides/fql#multi
func (session *Session) MultiFQL(queries Params) (Result, error) {
	res, err := session.graphFQL(Params{
		"q": queries,
	})

	if err != nil {
		return res, err
	}

	// query result is stored in "data" field.
	var data []Result
	err = res.DecodeField("data", &data)

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, fmt.Errorf("multi-fql result is not found.")
	}

	// Multi-fql data structure is:
	//     {
	//         "data": [
	//             {
	//                 "name": "query1",
	//                 "fql_result_set": [
	//                     {...}, {...}, ...
	//                 ]
	//             },
	//             {
	//                 "name": "query2",
	//                 "fql_result_set": [
	//                     {...}, {...}, ...
	//                 ]
	//             },
	//             ...
	//         ]
	//     }
	//
	// Parse the structure to following go map.
	//     {
	//         "query1": [
	//             // Come from field "fql_result_set".
	//             {...}, {...}, ...
	//         ],
	//         "query2": [
	//             {...}, {...}, ...
	//         ],
	//         ...
	//     }
	var name string
	var apiResponse interface{}
	var ok bool
	result := Result{}

	for k, v := range data {
		err = v.DecodeField("name", &name)

		if err != nil {
			return nil, fmt.Errorf("missing required field 'name' in multi-query data.%v. %v", k, err)
		}

		apiResponse, ok = v["fql_result_set"]

		if !ok {
			return nil, fmt.Errorf("missing required field 'fql_result_set' in multi-query data.%v.", k)
		}

		result[name] = apiResponse
	}

	return result, nil
}

// Makes an arbitrary HTTP request.
// It expects server responses a facebook Graph API response.
//     request, _ := http.NewRequest("https://graph.facebook.com/538744468", "GET", nil)
//     res, err := session.Request(request)
//     fmt.Println(res["gender"])  // get "male"
func (session *Session) Request(request *http.Request) (Result, error) {
	response, err := session.sendRequest(request)

	if err != nil {
		return nil, err
	}

	return MakeResult(response)
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

// Validates Session access token.
// Returns nil if access token is valid.
func (session *Session) Validate() (err error) {
	if session.accessToken == "" {
		err = fmt.Errorf("access token is not set.")
		return
	}

	var result Result
	result, err = session.Api("/me", GET, Params{"fields": "id"})

	if err != nil {
		return
	}

	if f := result.Get("id"); f == nil {
		err = fmt.Errorf("invalid access token.")
		return
	}

	return
}

// Inspect Session access token.
// Returns JSON array containing data about the inspected token.
// See https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow/v2.2#checktoken
func (session *Session) Inspect() (result Result, err error) {
	if session.accessToken == "" {
		err = fmt.Errorf("access token is not set.")
		return
	}

	if session.app == nil {
		err = fmt.Errorf("cannot inspect access token without binding an app.")
		return
	}

	appAccessToken := session.app.AppAccessToken()

	if appAccessToken == "" {
		err = fmt.Errorf("app access token is not set.")
		return
	}

	result, err = session.Api("/debug_token", GET, Params{
		"input_token":  session.accessToken,
		"access_token": appAccessToken,
	})

	if err != nil {
		return
	}

	// facebook stores everything, including error, inside result["data"].
	// make sure that result["data"] exists and doesn't contain error.
	if _, ok := result["data"]; !ok {
		err = fmt.Errorf("facebook inspect api returns unexpected result.")
		return
	}

	var data Result
	result.DecodeField("data", &data)
	result = data
	err = result.Err()
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
		session.appsecretProof = ""
	}
}

// Check appsecret proof is enabled or not.
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

// Enable or disable appsecret proof status.
// Returns error if there is no App associasted with this Session.
func (session *Session) EnableAppsecretProof(enabled bool) error {
	if session.app == nil {
		return fmt.Errorf("cannot change appsecret proof status without an associated App.")
	}

	if session.enableAppsecretProof != enabled {
		session.enableAppsecretProof = enabled

		// reset pre-calculated proof here to give caller a way to do so in some rare case,
		// e.g. associated app's secret is changed.
		session.appsecretProof = ""
	}

	return nil
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

	// always format as json.
	params["format"] = "json"

	// overwrite method as we always use post
	params["method"] = method

	// get graph api url.
	if session.isVideoPost(path, method) {
		graphUrl = session.getUrl("graph_video", path, nil)
	} else {
		graphUrl = session.getUrl("graph", path, nil)
	}

	response, err = session.sendPostRequest(graphUrl, params)

	// cannot get response from remote server
	if err != nil {
		return
	}

	res, err = MakeResult(response)
	return
}

func (session *Session) graphBatch(batchParams Params, params ...Params) (res []Result, err error) {
	var response []byte

	if batchParams == nil {
		batchParams = Params{}
	}

	batchParams["batch"] = params

	graphUrl := session.getUrl("graph", "", nil)
	response, err = session.sendPostRequest(graphUrl, batchParams)

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

func (session *Session) graphFQL(params Params) (res Result, err error) {
	if params == nil {
		params = Params{}
	}

	session.prepareParams(params)

	// encode url.
	buf := &bytes.Buffer{}
	buf.WriteString(domainMap["graph"])
	buf.WriteString("fql?")
	_, err = params.Encode(buf)

	if err != nil {
		return nil, fmt.Errorf("cannot encode params. %v", err)
	}

	var response []byte

	// it seems facebook disallow POST to /fql. always use GET for FQL.
	response, err = session.sendGetRequest(buf.String())

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

	// facebook may return an error
	err = res.Err()
	return
}

func (session *Session) prepareParams(params Params) {
	if _, ok := params["access_token"]; !ok && session.accessToken != "" {
		params["access_token"] = session.accessToken
	}

	if session.enableAppsecretProof && session.accessToken != "" && session.app != nil {
		params["appsecret_proof"] = session.AppsecretProof()
	}
}

func (session *Session) sendGetRequest(url string) ([]byte, error) {
	var request *http.Request
	var err error

	request, err = http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	return session.sendRequest(request)
}

func (session *Session) sendPostRequest(url string, params Params) ([]byte, error) {
	session.prepareParams(params)

	buf := &bytes.Buffer{}
	mime, err := params.Encode(buf)

	if err != nil {
		return nil, fmt.Errorf("cannot encode POST params. %v", err)
	}

	var request *http.Request

	request, err = http.NewRequest("POST", url, buf)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", mime)
	return session.sendRequest(request)
}

func (session *Session) sendRequest(request *http.Request) ([]byte, error) {
	var response *http.Response
	var err error

	if session.HttpClient == nil {
		response, err = http.DefaultClient.Do(request)
	} else {
		response, err = session.HttpClient.Do(request)
	}

	if err != nil {
		return nil, fmt.Errorf("cannot reach facebook server. %v", err)
	}

	defer response.Body.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, response.Body)

	if err != nil {
		return nil, fmt.Errorf("cannot read facebook response. %v", err)
	}

	return buf.Bytes(), nil
}

func (session *Session) isVideoPost(path string, method Method) bool {
	return method == POST && regexpIsVideoPost.MatchString(path)
}

func (session *Session) getUrl(name, path string, params Params) string {
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
