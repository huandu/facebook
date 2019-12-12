// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import "fmt"

// Error represents Facebook API error.
type Error struct {
	Message      string
	Type         string
	Code         int
	ErrorSubcode int    // subcode for authentication related errors.
	UserTitle    string `json:"error_user_title"`
	UserMessage  string `json:"error_user_msg"`
	IsTransient  bool   `json:"is_transient"`
	TraceID      string `json:"fbtrace_id"`
}

// Error returns error string.
func (e *Error) Error() string {
	return fmt.Sprintf("message: %s, error_user_title: %s, error_user_msg: %s", e.Message, e.UserTitle, e.UserMessage)
}
