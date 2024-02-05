package facebook

import (
	"fmt"
)

// Error represents Facebook API error.
type Error struct {
	Message      string
	Type         string
	Code         int
	ErrorSubcode int
	UserTitle    string `json:"error_user_title,omitempty"`
	UserMessage  string `json:"error_user_msg,omitempty"`
	IsTransient  bool   `json:"is_transient,omitempty"`
	TraceID      string `json:"fbtrace_id,omitempty"`
}

// Error returns error string.
func (e *Error) Error() string {
	return fmt.Sprintf("facebook: %s (code: %d; subcode: %d, title: %s, msg: %s)",
		e.Message, e.Code, e.ErrorSubcode, e.UserTitle, e.UserMessage)
}

// UnmarshalError represents a json decoder error.
type UnmarshalError struct {
	Payload []byte
	Message string
	Err     error
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("%s [err:%v]", e.Message, e.Err)
}
