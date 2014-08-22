// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2014, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"bytes"
	"io"
	"unicode"
)

func camelCaseToUnderScore(name string) string {
	if name == "" {
		return ""
	}

	buf := &bytes.Buffer{}

	for _, r := range name {
		if unicode.IsUpper(r) {
			if buf.Len() != 0 {
				buf.WriteRune('_')
			}

			buf.WriteRune(unicode.ToLower(r))
		} else {
			buf.WriteRune(r)
		}
	}

	return buf.String()
}

// Returns error string.
func (e *Error) Error() string {
	return e.Message
}

// Creates a new binary data holder.
func Data(filename string, source io.Reader) *binaryData {
	return &binaryData{
		Filename: filename,
		Source:   source,
	}
}

// Creates a binary file holder.
func File(filename, path string) *binaryFile {
	return &binaryFile{
		Filename: filename,
	}
}

// Creates a binary file holder and specific a different path for reading.
func FileAlias(filename, path string) *binaryFile {
	return &binaryFile{
		Filename: filename,
		Path:     path,
	}
}
