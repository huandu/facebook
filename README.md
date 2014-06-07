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

### Read a graph `user` object without access token ###

```go
    res, _ := facebook.Get("/huandu", nil)
    fmt.Println("my facebook id is", res["id"])
```

### Read a graph `user` object with a valid access token ###

```go
    res, err := facebook.Get("/me/feed", facebook.Params{
         "access_token": "a-valid-access-token",
    })
    
    if err != nil {
        // err can be an facebook API error.
        // if so, the Error struct contains error details.
        if e, ok := err.(*Error); ok {
            fmt.Logf("facebook error. [message:%v] [type:%v] [code:%v] [subcode:%v]",
                fbErr.Message, fbErr.Type, fbErr.Code, fbErr.ErrorSubcode)
            return
        }
    }
    
    // read my last feed.
    fmt.Println("my latest feed story is:", res.Get("data.0.story"))
```

### Use `App` and `Session` ###

It's recommended to use `App` and `Session` in a production app.

```go
    // create a global App var to hold your app id and secret.
    var globalApp = facebook.New("your-app-id", "your-app-secret")
    
    // facebook asks for a valid redirect uri when parsing signed request.
    // it's a new enforced policy starting in late 2013.
    // it can be omitted in a mobile app server.
    globalApp.RedirectUri = "http://your-site-canvas-url/"
    
    // here comes a client with a facebook signed request string in query string.
    // creates a new session with signed request.
    session, _ := globalApp.SessionFromSignedRequest(signedRequest)
    
    // or, you just get a valid access token in other way.
    // creates a session directly.
    seesion := globalApp.Session(token)
    
    // use session to send api request with your access token.
    res, _ := session.Get("/me/feed", nil)
    
    // validate access token. err is nil if token is valid.
    err := session.Validate()
```

### Read graph api response and decode result into a struct ###

As facebook Graph API always uses lower case words as keys in API response. This library can convert go's camel-case-style struct field name to underscore-style API key name.

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

Like `encoding/json` package, struct can have tag definitions, which is compatible with the JSON package.

Following is a full sample wrap up everything about struct decoding.

```go
    // define a facebook feed object.
    type FacebookFeed struct {
        Id string `facebook:",required"`         // must exist
        Story string
        From *FacebookFeedFrom `facebook:"from"` // use customized field name
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
        "method": facebook.GET,
        "relative_url": "huandu",
    }
    params2 := Params{
        "method": facebook.GET,
        "relative_url": uint64(100002828925788),
    }
    res, err := facebook.BatchApi(your_access_token, params1, params2)
    
    // res is a []Result. if err is nil, res[0] and res[1] are response to
    // params1 and params2 respectively.
```

### Send FQL query ###

```go
    results, _ := FQL("SELECT username FROM page WHERE page_id = 20531316728")
    fmt.Println(results[0]["username"]) // print "facebook"
    
    // most FQL query requires access token. create session to hold access token.
    session := &Session{}
    session.SetAccessToken("A-VALID-ACCESS-TOKEN")
    results, _ := session.FQL("SELECT username FROM page WHERE page_id = 20531316728")
    fmt.Println(results[0]["username"]) // print "facebook"
```

### Make multi-FQL ###

```go
    res, _ := MultiFQL(Params{
        "query1": "SELECT username FROM page WHERE page_id = 20531316728",
        "query2": "SELECT uid FROM user WHERE uid = 538744468",
    })
    var query1, query2 []Result
    
    // get response for query1 and query2.
    res.DecodeField("query1", &query1)
    res.DecodeField("query2", &query2)
    
    // most FQL query requires access token. create session to hold access token.
    session := &Session{}
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
    
    // default Session object uses http.DefaultClient which is not supported
    // by appengine. we have to create a Session and assign it a special client.
    seesion := globalApp.Session("a-access-token")
    session.HttpClient = urlfetch.Client(context)
    
    // now, session uses appengine http client now.
    res, err := session.Get("/me", nil)
```

### Select Graph API version ###
See [Platform Versioning](https://developers.facebook.com/docs/apps/versions) to understand facebook versioning strategy.

```go
    // this library uses default version by default.
    // change following global variable to specific a global default version.
    Version = "v2.0"
    
    // now you will get an error as v2.0 api doesn't allow you to do so.
    Api("huan.du", GET, nil)
    
    // you can also specify version per session.
    session := &Session{}
    session.Version = "v2.0" // overwrite global default.
```

I've try my best to add enough information in every public method and type. If you still have any question or suggestion, feel free to create an issue or send pull request to me. Thank you.

## Get It ##

Use `go get github.com/huandu/facebook` to get and install it.

## Out of Scope ##

1. No OAuth integration. This library only provides APIs to parse/verify access token and code generated in OAuth 2.0 authentication process.
2. No old RESTful API support. Such APIs are deprecated for years. Forget about them.

## License ##

This library is licensed under MIT license. See LICENSE for details.
