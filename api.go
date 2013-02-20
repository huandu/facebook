// A facebook graph api client in go.
// https://github.com/huandu/facebook/
// 
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

// A facebook graph api client purely in Go. Simple but powerful.
//
// The library design is highly influenced by facebook official js/php sdk,
// especially the design of the method Api() and its shorthands (Get/Post/Delete).
// So, this library should look familiar to one who has experience in
// official sdk.
//
// Here is a list of common scenarios while using this library.
//
// Scenario 1: Read a user object without access token.
//     res, _ := facebook.Get("/huandu", nil)
//     fmt.Println("my facebook id is", res["id"])
//
// Scenario 2: Read a user object with a valid access token.
//     res, _ := facebook.Get("/me/feed", facebook.Params{
//          "access_token": "a-valid-access-token",
//     })
//
//     // read my last feed
//     fmt.Println("my latest feed story is:", res.Get("data.0.story"))
//
// Scenario 3: Use App and Session struct.
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
//     res, _ := session.Get("/me/feed", nil)
//
//     // validate access token. err is nil if token is valid.
//     err := session.Validate()
//
// Scenario 4: Read graph api response and decode result into a struct.
//     // define a facebook feed object.
//     type FacebookFeed struct {
//         Id string `facebook:",required"` // must exist
//         Story string
//         From *FacebookFeedFrom
//         CreatedTime string `facebook:"created_time"` // use customized field name
//     }
//
//     type FacebookFeedFrom struct {
//         Name, Id string
//     }
//
//     // create a feed object direct from graph api result.
//     var feed FacebookFeed
//     res, _ := session.Get("/me/feed", nil)
//     res.DecodeField("data.0", &feed) // then you can use feed.
//
// Scenario 5: Batch graph api request.
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
// For more detailed documents, see doc for every public method. I've try my best to
// provide enough information.
//
// This library doesn't implement any deprecated old RESTful API. And it won't.
package facebook

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

// Makes a batch facebook graph api call.
//
// BatchApi supports most kinds of batch calls defines in facebook batch api document,
// except uploading binary data.
//
// See https://developers.facebook.com/docs/reference/api/batch/ to learn more about Batch Requests.
func BatchApi(accessToken string, params ...Params) ([]Result, error) {
    return defaultSession.graphBatch(accessToken, params...)
}
