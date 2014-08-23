# A Facebook Graph API SDK In Golang #

[![Build Status](https://travis-ci.org/huandu/facebook.png?branch=master)](https://travis-ci.org/huandu/facebook)

This is a Go library fully supports Facebook Graph API (both 1.0 and 2.0) with file upload, batch request, FQL and multi-FQL. It can be used in Google App Engine.

See [full document](http://godoc.org/github.com/huandu/facebook) for details.

## Usage ##

### Quick Tutorial ###

Here is a sample to read my Facebook username by uid.

```go
package main

import (
    "fmt"
    fb "github.com/huandu/facebook"
)

func main() {
    res, _ := fb.Get("/538744468", fb.Params{
        "fields": "username",
    })
    fmt.Println("here is my facebook username:", res["username"])
}
```

Type of `res["username"]` is `interface{}`. This library provides several helpful methods to decode fields to any Go type or even a custom Go struct.

```go
// Decode "username" to a go string.
var username string
res.DecodeField("username", &username)
fmt.Println("alternative way to get username:", username)

// It's also possible to decode the whole result into a predefined struct.
type User struct {
    Username string
}

var user User
res.Decode(&user)
fmt.Println("print username in struct:", user.Username)
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
        fmt.Logf("facebook error. [message:%v] [type:%v] [code:%v] [subcode:%v]",
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
    fmt.Logf("An error has happened %v", err)
    return
}

for _, item := range items {
    fmt.Println(item["id"])
}
```

### Use `App` and `Session` ###

It's recommended to use `App` and `Session` in a production app. They provide more controls over all API calls. They can also make your code clear and concise.

```go
// create a global App var to hold your app id and secret.
var globalApp = fb.New("your-app-id", "your-app-secret")

// facebook asks for a valid redirect uri when parsing signed request.
// it's a new enforced policy starting in late 2013.
globalApp.RedirectUri = "http://your.site/canvas/url/"

// here comes a client with a facebook signed request string in query string.
// creates a new session with signed request.
session, _ := globalApp.SessionFromSignedRequest(signedRequest)

// if you can get a valid access token in other way.
// creates a session directly with the token.
seesion := globalApp.Session(token)

// validate access token. err is nil if token is valid.
err := session.Validate()

// use session to send api request with your access token.
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

As facebook Graph API always uses lower case words as keys in API response. This library can convert go's camel-case-style struct field name to facebook's underscore-style API key name.

For instance, given we have following JSON response from facebook.

```json
{
    "foo_bar": "player"
}
```

We can use following struct to decode it.

```go
type Data struct {
    FooBar string  // "FooBar" => "foo_bar"
}
```

Like `encoding/json` package, struct can have tag definitions. Tags can be used to mark a field as required in API response and/or map field to a specific key name.

Following is a full sample wrap up everything about struct decoding.

```go
// define a facebook feed object.
type FacebookFeed struct {
    Id string `facebook:",required"`             // this field must exist in response
    Story string
    FeedFrom *FacebookFeedFrom `facebook:"from"` // use customized field name "from"
    CreatedTime string
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
    "relative_url": "huandu",
}
params2 := Params{
    "method": fb.GET,
    "relative_url": uint64(100002828925788),
}
res, err := fb.BatchApi(your_access_token, params1, params2)

// res is a []Result. if err is nil, res[0] and res[1] are response to
// params1 and params2 respectively.
```

### Send FQL query ###

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

Google App Engine provide `appengine/urlfetch` package as standard http client package. Default client in `net/http` doesn't work. One must explicitly set http client in `Session` to make it work.

```go
import (
    "appengine"
    "appengine/urlfetch"
)

// suppose it's the appengine context initialized somewhere.
var context appengine.Context

// default Session object uses http.DefaultClient which is not allowed to use
// in appengine. we have to create a Session and assign it a special client.
seesion := globalApp.Session("a-access-token")
session.HttpClient = urlfetch.Client(context)

// now, session uses appengine http client now.
res, err := session.Get("/me", nil)
```

### Select Graph API version ###

See [Platform Versioning](https://developers.facebook.com/docs/apps/versions) to understand facebook versioning strategy.

```go
// this library uses default version which is controlled by facebook app setting.
// change following global variable to specific a global default version.
fb.Version = "v2.0"

// now you will get an error as v2.0 api doesn't allow you to do so.
fb.Api("huan.du", GET, nil)

// you can also specify version per session.
session := &fb.Session{}
session.Version = "v2.0" // overwrite global default.
```

### Enable `appsecret_proof` ###

Facebook can verify Graph API Calls with `appsecret_proof`. It's a feature to make your Graph API call more secure. See [Securing Graph API Requests](https://developers.facebook.com/docs/graph-api/securing-requests) to know more about it.

```go
globalApp := fb.New("your-app-id", "your-app-secret")

// enable "appsecret_proof" for all sessions created by this app.
globalApp.EnableAppsecretProof = true

// all your calls in this session are secured.
session := globalApp.Session("a-valid-access-token")
session.Get("/me", nil)

// it's also possible to enable/disable this feature per session.
session.EnableAppsecretProof(false)
```

### Need more samples? ###

I've try my best to add enough information in every public method and type. If you still have any question or suggestion, feel free to create an issue or send pull request to me. Thank you.

## TODO ##

1. Real-time update subscriptions.

## Get It ##

Use `go get github.com/huandu/facebook` to get and install it.

## Out of Scope ##

1. No OAuth integration. This library only provides APIs to parse/verify access token and code generated in OAuth 2.0 authentication process.
2. No old RESTful API support. Such APIs are deprecated for years. Forget about them.

## License ##

This library is licensed under MIT license. See LICENSE for details.
