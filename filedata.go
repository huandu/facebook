// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2015, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"io"
)

// Binary data.
type binaryData struct {
	Filename    string    // filename used in multipart form writer.
	Source      io.Reader // file data source.
	ContentType string    // content type of the data.
}

// Binary file.
type binaryFile struct {
	Filename    string // filename used in multipart form writer.
	Path        string // path to file. must be readable.
	ContentType string // content type of the file.
}

// Creates new binary data holder.
func Data(filename string, source io.Reader) *binaryData {
	return &binaryData{
		Filename: filename,
		Source:   source,
	}
}

// Creates new binary data holder with arbitrary content type.
func DataWithContentType(filename string, source io.Reader, contentType string) *binaryData {
	return &binaryData{
		Filename:    filename,
		Source:      source,
		ContentType: contentType,
	}
}

// Creates a binary file holder.
func File(filename string) *binaryFile {
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

// Creates a new binary file holder with arbitrary content type.
func FileAliasWithContentType(filename, path, contentType string) *binaryFile {
	if path == "" {
		path = filename
	}

	return &binaryFile{
		Filename:    filename,
		Path:        path,
		ContentType: contentType,
	}
}
