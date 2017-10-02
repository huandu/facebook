// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2015, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

// Facebook API error.
type Error struct {
	Message      string
	Type         string
	Code         int
	ErrorSubcode int // subcode for authentication related errors.
}

// Returns error string.
func (e *Error) Error() string {
	return e.Message
}
