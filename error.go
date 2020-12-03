// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"fmt"
)

// Error represents Facebook API error.
type Error struct {
	Message      string
	Type         string
	Code         int
	ErrorSubcode int    // subcode for authentication related errors.
	UserTitle    string `json:"error_user_title,omitempty"`
	UserMessage  string `json:"error_user_msg,omitempty"`
	IsTransient  bool   `json:"is_transient,omitempty"`
	TraceID      string `json:"fbtrace_id,omitempty"`
}

// Error returns error string.
func (e *Error) Error() string {
	return fmt.Sprintf("facebook: %s (code: %d; error_subcode: %d, error_user_title: %s, error_user_msg: %s)",
		e.Message, e.Code, e.ErrorSubcode, e.UserTitle, e.UserMessage)
}

// UnmarshalError represents a json decoder error.
type UnmarshalError struct {
	Payload []byte // Body of the HTTP response.
	Message string // Verbose message for debug.
	Err     error  // The error returned by json decoder. It can be nil.
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("%s [err:%v]", e.Message, e.Err)
}
