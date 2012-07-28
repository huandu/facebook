// A facebook graph api client in go.
// https://github.com/huandu/facebook/
// 
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE
package facebook

// Holds facebook application information.
type App struct {
    AppId     string // facebook app id
    AppSecret string // facebook app secret
}

type Session struct {
    accessToken string // facebook access token. can be empty.
    app         *App
    id          string
}

// Api HTTP method.
// Can be GET, POST or DELETE.
type Method string

// Api params.
// 
// For general uses, just use Params as a ordinary map.
//
// For advanced uses, use MakeParams to create Params from any struct.
type Params map[string]interface{}

// Facebook api call result.
type Result map[string]interface{}
