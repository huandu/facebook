A Facebook Graph API Library In Go
==================================

[![Build Status](https://travis-ci.org/huandu/facebook.png?branch=master)](https://travis-ci.org/huandu/facebook)

This is a Go library fully supports Facebook Graph API with file upload, batch request and FQL. It's simple yet powerful.

It can be used in Google App Engine. See [document](http://godoc.org/github.com/huandu/facebook) for details.

Quick Tutorial
--------------

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

Full Document
-------------

Read http://godoc.org/github.com/huandu/facebook or use `go doc`.

Get It
------

Use `go get github.com/huandu/facebook` to get and install it.

Out of Scope
------------

1. No OAuth integration. This library only provides APIs to parse/verify access token and OAuth code.
2. No old RESTful API support. Such APIs are deprecated for years. Forget about them.

License
-------

This library is licensed under MIT license. See LICENSE for details.
