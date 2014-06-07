// A facebook graph api client in go.
// https://github.com/huandu/facebook/
// 
// Copyright 2012-2014, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

// This is a Go library fully supports Facebook Graph API with file upload,
// batch request and FQL. It also supports Graph API 2.0 using the same set
// of methods.
//
// Library design is highly influenced by facebook official PHP/JS SDK.
// If you have used PHP/JS SDK before, it should look quite familiar.
//
// Here is a list of common scenarios to help you to get started.
//
// Scenario 1: Read a graph `user` object without access token.
//     res, _ := facebook.Get("/huandu", nil)
//     fmt.Println("my facebook id is", res["id"])
//
// Scenario 2: Read a graph `user` object with a valid access token.
//     res, err := facebook.Get("/me/feed", facebook.Params{
//          "access_token": "a-valid-access-token",
//     })
//
//     if err != nil {
//         // err can be an facebook API error.
//         // if so, the Error struct contains error details.
//         if e, ok := err.(*Error); ok {
//             fmt.Logf("facebook error. [message:%v] [type:%v] [code:%v] [subcode:%v]",
//                 fbErr.Message, fbErr.Type, fbErr.Code, fbErr.ErrorSubcode)
//             return
//         }
//     }
//
//     // read my last feed.
//     fmt.Println("my latest feed story is:", res.Get("data.0.story"))
//
// Scenario 3: Use App and Session struct. It's recommended to use them
// in a production app.
//     // create a global App var to hold your app id and secret.
//     var globalApp = facebook.New("your-app-id", "your-app-secret")
//
//     // facebook asks for a valid redirect uri when parsing signed request.
//     // it's a new enforced policy starting in late 2013.
//     // it can be omitted in a mobile app server.
//     globalApp.RedirectUri = "http://your-site-canvas-url/"
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
//     res, _ := session.Get("/me/feed", nil)
//
//     // validate access token. err is nil if token is valid.
//     err := session.Validate()
//
// Scenario 4: Read graph api response and decode result into a struct.
// 
// As facebook Graph API always uses lower case words as keys in API response.
// this library can convert go's camel-case-style struct field name to underscore-style
// API key name.
//
// For instance, given we have following JSON response from facebook.
//     {
//         "foo_bar": "player"
//     }
//
// We can use following struct to decode it.
//     type Data struct {
//         FooBar string  // "FooBar" => "foo_bar"
//     }
// 
// Like `encoding/json` package, struct can have tag definitions, which is compatible with
// the JSON package.
//
// Following is a full sample wrap up everything about struct decoding.
//     // define a facebook feed object.
//     type FacebookFeed struct {
//         Id string `facebook:",required"`         // must exist
//         Story string
//         From *FacebookFeedFrom `facebook:"from"` // use customized field name
//         CreatedTime string
//     }
//
//     type FacebookFeedFrom struct {
//         Name, Id string
//     }
//
//     // create a feed object direct from graph api result.
//     var feed FacebookFeed
//     res, _ := session.Get("/me/feed", nil)
//     res.DecodeField("data.0", &feed) // read latest feed
//
// Scenario 5: Send a batch request.
//     params1 := Params{
//         "method": facebook.GET,
//         "relative_url": "huandu",
//     }
//     params2 := Params{
//         "method": facebook.GET,
//         "relative_url": uint64(100002828925788),
//     }
//     res, err := facebook.BatchApi(your_access_token, params1, params2)
//     // res is a []Result. if err is nil, res[0] and res[1] are response to
//     // params1 and params2 respectively.
//
// Scenario 6: Send FQL query.
//     results, _ := FQL("SELECT username FROM page WHERE page_id = 20531316728")
//     fmt.Println(results[0]["username"]) // print "facebook"
//
//     // most FQL query requires access token. create session to hold access token.
//     session := &Session{}
//     session.SetAccessToken("A-VALID-ACCESS-TOKEN")
//     results, _ := session.FQL("SELECT username FROM page WHERE page_id = 20531316728")
//     fmt.Println(results[0]["username"]) // print "facebook"
//
// Scenario 7: Make multi-FQL.
//     res, _ := MultiFQL(Params{
//         "query1": "SELECT username FROM page WHERE page_id = 20531316728",
//         "query2": "SELECT uid FROM user WHERE uid = 538744468",
//     })
//     var query1, query2 []Result
//
//     // get response for query1 and query2.
//     res.DecodeField("query1", &query1)
//     res.DecodeField("query2", &query2)
//
//     // most FQL query requires access token. create session to hold access token.
//     session := &Session{}
//     session.SetAccessToken("A-VALID-ACCESS-TOKEN")
//     res, _ := session.MultiFQL(Params{
//         "query1": "...",
//         "query2": "...",
//     })
//     // same as the sample without access token...
//
// Scenario 8: Use it in Google App Engine with `appengine/urlfetch` package.
//     import (
//         "appengine"
//         "appengine/urlfetch"
//     )
//
//     // suppose it's the appengine context initialized somewhere.
//     var context appengine.Context
//
//     // default Session object uses http.DefaultClient which is not supported
//     // by appengine. we have to create a Session and assign it a special client.
//     seesion := globalApp.Session("a-access-token")
//     session.HttpClient = urlfetch.Client(context)
//
//     // now, session uses appengine http client now.
//     res, err := session.Get("/me", nil)
//
// Scenario 9: Select Graph API version. See https://developers.facebook.com/docs/apps/versions .
//     // this library uses default version by default.
//     // change following global variable to specific a global default version.
//     Version = "v2.0"
//
//     // now you will get an error as v2.0 api doesn't allow you to do so.
//     Api("huan.du", GET, nil)
//
//     // you can also specify version per session.
//     session := &Session{}
//     session.Version = "v2.0" // overwrite global default.
//
// I've try my best to add enough information in every public method and type.
// If you still have any question or suggestion, feel free to create an issue
// or send pull request to me. Thank you.
//
// This library doesn't implement any deprecated old RESTful API. And it won't.
package facebook

var (
    // Default facebook api version.
    // It can be "v1.0" or "v2.0" or empty per facebook current document.
    // See https://developers.facebook.com/docs/apps/versions for details.
    Version string
)

// Makes a facebook graph api call.
//
// Method can be GET, POST, DELETE or PUT.
//
// Params represents query strings in this call.
// Keys and values in params will be encoded for URL automatically. So there is
// no need to encode keys or values in params manually. Params can be nil.
//
// If you want to get
//     https://graph.facebook.com/huandu?fields=name,username
// Api should be called as following
//     Api("/huandu", GET, Params{"fields": "name,username"})
// or in a simplified way
//     Get("/huandu", Params{"fields": "name,username"})
//
// Api is a wrapper of Session.Api(). It's designed for graph api that doesn't require
// app id, app secret and access token. It can be called in multiple goroutines.
//
// If app id, app secret or access token is required in graph api, caller should
// create a new facebook session through App instance instead.
func Api(path string, method Method, params Params) (Result, error) {
    return defaultSession.Api(path, method, params)
}

// Get is a short hand of Api(path, GET, params).
func Get(path string, params Params) (Result, error) {
    return Api(path, GET, params)
}

// Post is a short hand of Api(path, POST, params).
func Post(path string, params Params) (Result, error) {
    return Api(path, POST, params)
}

// Delete is a short hand of Api(path, DELETE, params).
func Delete(path string, params Params) (Result, error) {
    return Api(path, DELETE, params)
}

// Put is a short hand of Api(path, PUT, params).
func Put(path string, params Params) (Result, error) {
    return Api(path, PUT, params)
}

// Makes a batch facebook graph api call.
//
// BatchApi supports most kinds of batch calls defines in facebook batch api document,
// except uploading binary data. Use Batch to do so.
//
// Note: API response is stored in "body" field of a Result.
//     var res1, res2 Result
//     results, _ := BatchApi(accessToken, Params{...}, Params{...})
//
//     // Get batch request response.
//     results[0].DecodeField("body", &res1)
//     results[1].DecodeField("body", &res2)
//
// Facebook document: https://developers.facebook.com/docs/graph-api/making-multiple-requests
func BatchApi(accessToken string, params ...Params) ([]Result, error) {
    return Batch(Params{"access_token": accessToken}, params...)
}

// Makes a batch facebook graph api call.
// Batch is designed for more advanced usage including uploading binary files.
//
// An uploading files sample
//     // equivalent to following curl command (borrowed from facebook docs)
//     //     curl \
//     //         -F 'access_token=…' \
//     //         -F 'batch=[{"method":"POST","relative_url":"me/photos","body":"message=My cat photo","attached_files":"file1"},{"method":"POST","relative_url":"me/photos","body":"message=My dog photo","attached_files":"file2"},]' \
//     //         -F 'file1=@cat.gif' \
//     //         -F 'file2=@dog.jpg' \
//     //         https://graph.facebook.com
//     Batch(Params{
//         "access_token": "the-access-token",
//         "file1": File("cat.gif"),
//         "file2": File("dog.jpg"),
//     }, Params{
//         "method": "POST",
//         "relative_url": "me/photos",
//         "body": "message=My cat photo",
//         "attached_files": "file1",
//     }, Params{
//         "method": "POST",
//         "relative_url": "me/photos",
//         "body": "message=My dog photo",
//         "attached_files": "file2",
//     })
//
// Facebook document: https://developers.facebook.com/docs/graph-api/making-multiple-requests
func Batch(batchParams Params, params ...Params) ([]Result, error) {
    return defaultSession.Batch(batchParams, params...)
}

// Makes a FQL query.
// Returns a slice of Result. If there is no query result, the result is nil.
//
// FQL can only make query without "access_token". For query requiring "access_token", create
// Session and call its FQL method.
//
// Facebook document: https://developers.facebook.com/docs/technical-guides/fql#query
func FQL(query string) ([]Result, error) {
    return defaultSession.FQL(query)
}

// Makes a multi FQL query.
// Returns a parsed Result. The key is the multi query key, and the value is the query result.
//
// MultiFQL can only make query without "access_token". For query requiring "access_token", create
// Session and call its MultiFQL method.
//
// See Session.MultiFQL document for samples.
//
// Facebook document: https://developers.facebook.com/docs/technical-guides/fql#multi
func MultiFQL(queries Params) (Result, error) {
    return defaultSession.MultiFQL(queries)
}
