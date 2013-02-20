A Facebook Graph API Library In Go
=================================

This is a Go library supports Facebook Graph API and FQL. It's simple but powerful.

Quick Tutorial
---------------

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

      // It's easy to decode "username" to a go string too.
      var username string
      res.DecodeField("username", &username)
      fmt.Println("alternative way to get username:", username)

      // It's also possible to decode the whole result into a predefined struct.
      // Assume the struct is defined as following:
      //     type User struct {
      //         Username string
      //     }
      //
      // Then, you can use following code to fill struct with values in result.
      //     var user User
      //     res.Decode(&user)
  }
```

For more document, please read http://godoc.org/github.com/huandu/facebook or use `go doc`.

Get It
------

Use `go get github.com/huandu/facebook` to get and install it.

Out of Scope
------------

1. No OAuth integration.
2. No old RESTful API support.

License
-------

This library is licensed under MIT license.
