// A facebook graph api client in go.
// https://github.com/huandu/facebook/
// 
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

// A Facebook Graph API library purely in Go. Simple but powerful.
//
// Library design is highly influenced by facebook official php/js SDK.
// So, it should look familiar to one who has experience in official sdk.
//
// To get start, here is a list of common scenarios for you.
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
//     // facebook asks for a valid redirect uri when parsing signed request.
//     // it's a new enforced policy starting in late 2013.
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
//     res.DecodeField("data.0", &feed) // read latest feed
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
// See https://developers.facebook.com/docs/reference/api/batch/ to learn more about Batch Requests.
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
// See https://developers.facebook.com/docs/reference/api/batch/ to learn more about Batch Requests.
func Batch(batchParams Params, params ...Params) ([]Result, error) {
    return defaultSession.Batch(batchParams, params...)
}
