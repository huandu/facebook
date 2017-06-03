# A Facebook Graph API SDK In Golang #

[![Build Status](https://travis-ci.org/huandu/facebook.png?branch=master)](https://travis-ci.org/huandu/facebook)

This is a Go package fully supports Facebook Graph API with file upload, batch request, FQL and multi-FQL. It can be used in Google App Engine.



### Quick start ###

Here is a sample that reads my Facebook first name by uid.

```go
package main

import (
    "fmt"
    fb "github.com/huandu/facebook"
)

func main() {
    res, _ := fb.Get("/538744468", fb.Params{
        "fields": "first_name",
        "access_token": "a-valid-access-token",
    })
    fmt.Println("here is my facebook first name:", res["first_name"])
}
```

The type of `res` is `fb.Result` (a.k.a. `map[string]interface{}`).
This type has several useful methods to decode `res` to any Go type safely.

```go
// Decode "first_name" to a Go string.
var first_name string
res.DecodeField("first_name", &first_name)
fmt.Println("alternative way to get first_name:", first_name)

// It's also possible to decode the whole result into a predefined struct.
type User struct {
    FirstName string
}

var user User
res.Decode(&user)
fmt.Println("print first_name in struct:", user.FirstName)
```

If a type implements the `json.Unmarshaler` interface, `Decode` or `DecodeField` will use it to unmarshal JSON.

```go
res := Result{
    "create_time": "2006-01-02 15:16:17Z",
}

// Type `*time.Time` implements `json.Unmarshaler`.
// res.DecodeField will use the interface to unmarshal data.
var tm time.Time
res.DecodeField("create_time", &tm)
```

### Read a graph `user` object with a valid access token ###

```go
res, err := fb.Get("/me/feed", fb.Params{
     "access_token": "a-valid-access-token",
})

if err != nil {
    // err can be an facebook API error.
    // if so, the Error struct contains error details.
    if e, ok := err.(*Error); ok {
        fmt.Printf("facebook error. [message:%v] [type:%v] [code:%v] [subcode:%v]",
            e.Message, e.Type, e.Code, e.ErrorSubcode)
        return
    }

    return
}

// read my last feed.
fmt.Println("my latest feed story is:", res.Get("data.0.story"))
```

### Read a graph `search` for page and decode slice of maps

```go
res, _ := fb.Get("/search", fb.Params{
        "access_token": "a-valid-access-token",
        "type":         "page",
        "q":            "nightlife,singapore",
    })

var items []fb.Result

err := res.DecodeField("data", &items)

if err != nil {
    fmt.Printf("An error has happened %v", err)
    return
}

for _, item := range items {
    fmt.Println(item["id"])
}
```

### Use `App` and `Session` ###

It's recommended to use `App` and `Session` in a production app. They provide more controls over all API calls. They can also make code clear and concise.

```go
// create a global App var to hold app id and secret.
var globalApp = fb.New("your-app-id", "your-app-secret")

// facebook asks for a valid redirect uri when parsing signed request.
// it's a new enforced policy starting in late 2013.
globalApp.RedirectUri = "http://your.site/canvas/url/"

// here comes a client with a facebook signed request string in query string.
// creates a new session with signed request.
session, _ := globalApp.SessionFromSignedRequest(signedRequest)

// if there is another way to get decoded access token,
// creates a session directly with the token.
session := globalApp.Session(token)

// validate access token. err is nil if token is valid.
err := session.Validate()

// use session to send api request with access token.
res, _ := session.Get("/me/feed", nil)
```

### Use `paging` field in response. ###

Some Graph API responses use a special JSON structure to provide paging information. Use `Result.Paging()` to walk through all data in such results.

```go
res, _ := session.Get("/me/home", nil)

// create a paging structure.
paging, _ := res.Paging(session)

// get current results.
results := paging.Data()

// get next page.
noMore, err := paging.Next()
results = paging.Data()
```

### Read graph api response and decode result into a struct ###

As facebook Graph API always uses lower case words as keys in API response.
This package can convert go's camel-case-style struct field name to facebook's underscore-style API key name.

For instance, to decode following JSON response...

```json
{
    "foo_bar": "player"
}
```

One can use following struct.

```go
type Data struct {
    FooBar string  // "FooBar" maps to "foo_bar" in JSON automatically in this case.
}
```

Decoding behavior can be changed per field through field tag -- just like what `encoding/json` does.

Following is a sample shows all possible field tags.

```go
// define a facebook feed object.
type FacebookFeed struct {
    Id          string `facebook:",required"`             // this field must exist in response.
                                                          // mind the "," before "required".
    Story       string
    FeedFrom    *FacebookFeedFrom `facebook:"from"`       // use customized field name "from".
    CreatedTime string `facebook:"created_time,required"` // both customized field name and "required" flag.
    Omitted     string `facebook:"-"`                     // this field is omitted when decoding.
}

type FacebookFeedFrom struct {
    Name, Id string
}

// create a feed object direct from graph api result.
var feed FacebookFeed
res, _ := session.Get("/me/feed", nil)
res.DecodeField("data.0", &feed) // read latest feed
```

### Send a batch request ###

```go
params1 := Params{
    "method": fb.GET,
    "relative_url": "me",
}
params2 := Params{
    "method": fb.GET,
    "relative_url": uint64(100002828925788),
}
results, err := fb.BatchApi(your_access_token, params1, params2)

if err != nil {
    // check error...
    return
}

// batchResult1 and batchResult2 are response for params1 and params2.
batchResult1, _ := results[0].Batch()
batchResult2, _ := results[1].Batch()

// Use parsed result.
var id string
res := batchResult1.Result
res.DecodeField("id", &id)

// Use response header.
contentType := batchResult1.Header.Get("Content-Type")
```

### Send FQL query ###

*FQL is deprecated by facebook right now.*

```go
results, _ := fb.FQL("SELECT username FROM page WHERE page_id = 20531316728")
fmt.Println(results[0]["username"]) // print "facebook"

// most FQL query requires access token. create session to hold access token.
session := &fb.Session{}
session.SetAccessToken("A-VALID-ACCESS-TOKEN")
results, _ := session.FQL("SELECT username FROM page WHERE page_id = 20531316728")
fmt.Println(results[0]["username"]) // print "facebook"
```

### Make multi-FQL ###

*FQL is deprecated by facebook right now.*

```go
res, _ := fb.MultiFQL(Params{
    "query1": "SELECT username FROM page WHERE page_id = 20531316728",
    "query2": "SELECT uid FROM user WHERE uid = 538744468",
})
var query1, query2 []Result

// get response for query1 and query2.
res.DecodeField("query1", &query1)
res.DecodeField("query2", &query2)

// most FQL query requires access token. create session to hold access token.
session := &fb.Session{}
session.SetAccessToken("A-VALID-ACCESS-TOKEN")
res, _ := session.MultiFQL(Params{
    "query1": "...",
    "query2": "...",
})

// same as the sample without access token...
```

### Use it in Google App Engine ###

Google App Engine provides `appengine/urlfetch` package as the standard http client package. The default client in `net/http` doesn't work. One must explicitly set http client in `Session` to make it work.

```go
import (
    "appengine"
    "appengine/urlfetch"
)

// suppose it's the appengine context initialized somewhere.
var context appengine.Context

// default Session object uses http.DefaultClient which is not allowed to use
// in appengine. one has to create a Session and assign it a special client.
seesion := globalApp.Session("a-access-token")
session.HttpClient = urlfetch.Client(context)

// now, session uses appengine http client now.
res, err := session.Get("/me", nil)
```

### Select Graph API version ###

See [Platform Versioning](https://developers.facebook.com/docs/apps/versions) to understand facebook versioning strategy.

```go
// this package uses default version which is controlled by facebook app setting.
// change following global variable to specific a global default version.
fb.Version = "v2.0"

// starting with graph api v2.0, it's not allowed to get user information without access token.
fb.Api("huan.du", GET, nil)

// it's possible to specify version per session.
session := &fb.Session{}
session.Version = "v2.0" // overwrite global default.
```

### Enable `appsecret_proof` ###

Facebook can verify Graph API Calls with `appsecret_proof`. It's a feature to make Graph API call more secure. See [Securing Graph API Requests](https://developers.facebook.com/docs/graph-api/securing-requests) to know more about it.

```go
globalApp := fb.New("your-app-id", "your-app-secret")

// enable "appsecret_proof" for all sessions created by this app.
globalApp.EnableAppsecretProof = true

// all calls in this session are secured.
session := globalApp.Session("a-valid-access-token")
session.Get("/me", nil)

// it's also possible to enable/disable this feature per session.
session.EnableAppsecretProof(false)
```

### Debugging API Requests ###

Facebook introduces a way to debug graph API calls. See [Debugging API Requests](https://developers.facebook.com/docs/graph-api/using-graph-api/v2.3#debugging) for details.

This package provides both package level and per session debug flag. Set `Debug` to a `DEBUG_*` constant to change debug mode globally; or use `Session#SetDebug` to change debug mode for one session.

When debug mode is turned on, use `Result#DebugInfo` to get `DebugInfo` struct from result.

```go
fb.Debug = fb.DEBUG_ALL

res, _ := fb.Get("/me", fb.Params{"access_token": "xxx"})
debugInfo := res.DebugInfo()

fmt.Println("http headers:", debugInfo.Header)
fmt.Println("facebook api version:", debugInfo.FacebookApiVersion)
```

### Work with package `golang.org/x/oauth2` ##

Package `golang.org/x/oauth2` can handle facebook OAuth2 authentication process and access token very well. This package can work with it by setting `Session#HttpClient` to OAuth2's client.

```go
import (
    "golang.org/x/oauth2"
    oauth2fb "golang.org/x/oauth2/facebook"
    fb "github.com/huandu/facebook"
)

// Get facebook access token.
conf := &oauth2.Config{
    ClientID:     "AppId",
    ClientSecret: "AppSecret",
    RedirectURL:  "CallbackURL",
    Scopes:       []string{"email"},
    Endpoint:     oauth2fb.Endpoint,
}
token, err := conf.Exchange(oauth2.NoContext, "code")

// Create a client to manage access token life cycle.
client := conf.Client(oauth2.NoContext, token)

// Use OAuth2 client with session.
session := &fb.Session{
    Version:    "v2.4",
    HttpClient: client,
}

// Use session.
res, _ := session.Get("/me", nil)
```

## Change Log ##

See [CHANGELOG.md](CHANGELOG.md).

## Out of Scope ##

precated for years. Forget about them.

## License ##

This package is licensed under MIT license. See LICENSE for details.
