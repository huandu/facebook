# A Facebook Graph API SDK In Golang

[![Build Status](https://github.com/huandu/facebook/workflows/Go/badge.svg)](https://github.com/huandu/facebook/actions)
[![GoDoc](https://godoc.org/github.com/huandu/facebook?status.svg)](https://pkg.go.dev/github.com/huandu/facebook/v2)

This is a Go package that fully supports the [Facebook Graph API](https://developers.facebook.com/docs/graph-api/) with file upload, batch request and marketing API. It can be used in Google App Engine.

API documentation can be found on [godoc](https://pkg.go.dev/github.com/huandu/facebook/v2).

Feel free to create an issue or send me a pull request if you have any "how-to" question or bug or suggestion when using this package. I'll try my best to reply to it.

## Install

If `go mod` is enabled, install this package with `go get github.com/huandu/facebook/v2`. If not, call `go get -u github.com/huandu/facebook` to get the latest master branch version.

Note that, since go1.14, [incompatible versions are omitted](https://golang.org/doc/go1.14#incompatible-versions) unless specified explicitly. Therefore, it's highly recommended to upgrade the import path to `github.com/huandu/facebook/v2` when possible to avoid any potential dependency error.

## Usage

### Quick start

Here is a sample that reads my Facebook first name by uid.

```go
package main

import (
    "fmt"
    fb "github.com/huandu/facebook/v2"
)

func main() {
    res, _ := fb.Get("/538744468", fb.Params{
        "fields": "first_name",
        "access_token": "a-valid-access-token",
    })
    fmt.Println("Here is my Facebook first name:", res["first_name"])
}
```

The type of `res` is `fb.Result` (a.k.a. `map[string]interface{}`).
This type has several useful methods to decode `res` to any Go type safely.

```go
// Decode "first_name" to a Go string.
var first_name string
res.DecodeField("first_name", &first_name)
fmt.Println("Here's an alternative way to get first_name:", first_name)

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
    "create_time": "2006-01-02T15:16:17Z",
}

// Type `*time.Time` implements `json.Unmarshaler`.
// res.DecodeField will use the interface to unmarshal data.
var tm time.Time
res.DecodeField("create_time", &tm)
```

### Read a graph `user` object with a valid access token

```go
res, err := fb.Get("/me/feed", fb.Params{
     "access_token": "a-valid-access-token",
})

if err != nil {
    // err can be a Facebook API error.
    // if so, the Error struct contains error details.
    if e, ok := err.(*Error); ok {
        fmt.Printf("facebook error. [message:%v] [type:%v] [code:%v] [subcode:%v] [trace:%v]",
            e.Message, e.Type, e.Code, e.ErrorSubcode, e.TraceID)
        return
    }

    // err can be an unmarshal error when Facebook API returns a message which is not JSON.
    if e, ok := err.(*UnmarshalError); ok {
        fmt.Printf("facebook error. [message:%v] [err:%v] [payload:%v]",
            e.Message, e.Err, string(e.Payload))
        return
    }

    return
}

// read my last feed story.
fmt.Println("My latest feed story is:", res.Get("data.0.story"))
```

### Read a graph `search` for page and decode slice of maps

```go
res, _ := fb.Get("/pages/search", fb.Params{
        "access_token": "a-valid-access-token",
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

### Use `App` and `Session`

It's recommended to use `App` and `Session` in a production app. They provide more control over all API calls. They can also make code clearer and more concise.

```go
// Create a global App var to hold app id and secret.
var globalApp = fb.New("your-app-id", "your-app-secret")

// Facebook asks for a valid redirect URI when parsing the signed request.
// It's a newly enforced policy starting as of late 2013.
globalApp.RedirectUri = "http://your.site/canvas/url/"

// Here comes a client with a Facebook signed request string in the query string.
// This will return a new session from a signed request.
session, _ := globalApp.SessionFromSignedRequest(signedRequest)

// If there is another way to get decoded access token,
// this will return a session created directly from the token.
session := globalApp.Session(token)

// This validates the access token by ensuring that the current user ID is properly returned. err is nil if the token is valid.
err := session.Validate()

// Use the new session to send an API request with the access token.
res, _ := session.Get("/me/feed", nil)
```

By default, all requests are sent to Facebook servers. If you wish to override the API base URL for unit-testing purposes - just set the respective `Session` field.

```go
testSrv := httptest.NewServer(someMux)
session.BaseURL = testSrv.URL + "/"
```

Facebook returns most timestamps in an ISO9601 format which can't be natively parsed by Go's `encoding/json`.
Setting `RFC3339Timestamps` `true` on the `Session` or at the global level will cause proper RFC3339 timestamps to be requested from Facebook.
RFC3339 is what `encoding/json` natively expects.

```go
fb.RFC3339Timestamps = true
session.RFC3339Timestamps = true
```

Setting either of these to true will cause `date_format=Y-m-d\TH:i:sP` to be sent as a parameter on every request. The format string is a PHP `date()` representation of RFC3339.
More info is available in [this issue](https://github.com/huandu/facebook/issues/95).

### Use `paging` field in response

Some Graph API responses use a special JSON structure to provide paging information. Use `Result.Paging()` to walk through all data in such results.

```go
res, _ := session.Get("/me/home", nil)

// create a paging structure.
paging, _ := res.Paging(session)

var allResults []Result

// append first page of results to slice of Result
allResults = append(allResults, paging.Data()...)

for {
  // get next page.
  noMore, err := paging.Next()
  if err != nil {
    panic(err)
  }
  if noMore {
    // No more results available
    break
  }
  // append current page of results to slice of Result
  allResults = append(allResults, paging.Data()...)
}

```

### Read Graph API response and decode result in a struct

The Facebook Graph API always uses snake case keys in API response.
This package can automatically convert from snake case to Go's camel-case-style style struct field names.

For instance, to decode the following JSON response...

```json
{
  "foo_bar": "player"
}
```

One can use the following struct.

```go
type Data struct {
    FooBar string  // "FooBar" maps to "foo_bar" in JSON automatically in this case.
}
```

The decoding of each struct field can be customized by the format string stored under the `facebook` key or the "json" key in the struct field's tag. The `facebook` key is recommended as it's specifically designed for this package.

Following is a sample that shows all possible field tags.

```go
// define a Facebook feed object.
type FacebookFeed struct {
    Id          string            `facebook:",required"`             // this field must exist in response.
                                                                     // mind the "," before "required".
    Story       string
    FeedFrom    *FacebookFeedFrom `facebook:"from"`                  // use customized field name "from".
    CreatedTime string            `facebook:"created_time,required"` // both customized field name and "required" flag.
    Omitted     string            `facebook:"-"`                     // this field is omitted when decoding.
}

type FacebookFeedFrom struct {
    Name string `json:"name"`                   // the "json" key also works as expected.
    Id string   `facebook:"id" json:"shadowed"` // if both "facebook" and "json" key are set, the "facebook" key is used.
}

// create a feed object direct from Graph API result.
var feed FacebookFeed
res, _ := session.Get("/me/feed", nil)
res.DecodeField("data.0", &feed) // read latest feed
```

### Send a batch request

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

### Using with Google App Engine

Google App Engine provides the `appengine/urlfetch` package as the standard HTTP client package.
For this reason, the default client in `net/http` won't work.
One must explicitly set the HTTP client in `Session` to make it work.

```go
import (
    "appengine"
    "appengine/urlfetch"
)

// suppose it's the AppEngine context initialized somewhere.
var context appengine.Context

// default Session object uses http.DefaultClient which is not allowed to use
// in appengine. one has to create a Session and assign it a special client.
seesion := globalApp.Session("a-access-token")
session.HttpClient = urlfetch.Client(context)

// now, the session uses AppEngine HTTP client now.
res, err := session.Get("/me", nil)
```

### Select Graph API version

See [Platform Versioning](https://developers.facebook.com/docs/apps/versions) to understand the Facebook versioning strategy.

```go
// This package uses the default version which is controlled by the Facebook app setting.
// change following global variable to specify a global default version.
fb.Version = "v3.0"

// starting with Graph API v2.0; it's not allowed to get useful information without an access token.
fb.Api("huan.du", GET, nil)

// it's possible to specify version per session.
session := &fb.Session{}
session.Version = "v3.0" // overwrite global default.
```

### Enable `appsecret_proof`

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

### Debugging API Requests

Facebook has introduced a way to debug Graph API calls. See [Debugging API Requests](https://developers.facebook.com/docs/graph-api/using-graph-api/debugging) for more details.

This package provides both a package level and per session debug flag. Set `Debug` to a `DEBUG_*` constant to change debug mode globally, or use `Session#SetDebug` to change debug mode for one session.

When debug mode is turned on, use `Result#DebugInfo` to get `DebugInfo` struct from the result.

```go
fb.Debug = fb.DEBUG_ALL

res, _ := fb.Get("/me", fb.Params{"access_token": "xxx"})
debugInfo := res.DebugInfo()

fmt.Println("http headers:", debugInfo.Header)
fmt.Println("facebook api version:", debugInfo.FacebookApiVersion)
```

### Monitoring API usage info

Call `Result#UsageInfo` to get a `UsageInfo` struct containing both app and page-level rate limit information from the result. More information about rate limiting can be found [here](https://developers.facebook.com/docs/graph-api/overview/rate-limiting).

```go
res, _ := fb.Get("/me", fb.Params{"access_token": "xxx"})
usageInfo := res.UsageInfo()

fmt.Println("App level rate limit information:", usageInfo.App)
fmt.Println("Page level rate limit information:", usageInfo.Page)
fmt.Println("Ad account rate limiting information:", usageInfo.AdAccount)
fmt.Println("Business use case usage information:", usageInfo.BusinessUseCase)
```

### Work with package `golang.org/x/oauth2`

The `golang.org/x/oauth2` package can handle the Facebook OAuth2 authentication process and access token quite well. This package can work with it by setting `Session#HttpClient` to OAuth2's client.

```go
import (
    "golang.org/x/oauth2"
    oauth2fb "golang.org/x/oauth2/facebook"
    fb "github.com/huandu/facebook/v2"
)

// Get Facebook access token.
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

### Control timeout and cancelation with `Context`

The `Session` accept a `Context`.

```go
// Create a new context.
ctx, cancel := context.WithTimeout(session.Context(), 100 * time.Millisecond)
defer cancel()

// Call an API with ctx.
// The return value of `session.WithContext` is a shadow copy of original session and
// should not be stored. It can be used only once.
result, err := session.WithContext(ctx).Get("/me", nil)
```

See [this Go blog post about context](https://blog.golang.org/context) for more details about how to use `Context`.

## Change Log

See [CHANGELOG.md](CHANGELOG.md).

## Out of Scope

1. No OAuth integration. This package only provides APIs to parse/verify access token and code generated in OAuth 2.0 authentication process.
2. No old RESTful API and FQL support. Such APIs are deprecated for years. Forget about them.

## License

This package is licensed under the MIT license. See LICENSE for details.
