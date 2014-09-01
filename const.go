// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2014, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"reflect"
	"regexp"
)

// Facebook graph api methods.
const (
	GET    Method = "GET"
	POST   Method = "POST"
	DELETE Method = "DELETE"
	PUT    Method = "PUT"
)

const (
	ERROR_CODE_UNKNOWN = -1 // unknown facebook graph api error code.

	_MIME_FORM_URLENCODED = "application/x-www-form-urlencoded"
)

var (
	// Maps aliases to Facebook domains.
	// Copied from Facebook PHP SDK.
	domainMap = map[string]string{
		"api":         "https://api.facebook.com/",
		"api_video":   "https://api-video.facebook.com/",
		"api_read":    "https://api-read.facebook.com/",
		"graph":       "https://graph.facebook.com/",
		"graph_video": "https://graph-video.facebook.com/",
		"www":         "https://www.facebook.com/",
	}

	// checks whether it's a video post.
	regexpIsVideoPost = regexp.MustCompile(`/^(\/)(.+)(\/)(videos)$/`)

	// default facebook session.
	defaultSession = &Session{}

	typeOfPointerToBinaryData = reflect.TypeOf(&binaryData{})
	typeOfPointerToBinaryFile = reflect.TypeOf(&binaryFile{})

	facebookSuccessJsonBytes = []byte("true")
)
