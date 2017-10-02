// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2015, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"bytes"
	"strings"
	"testing"
)

func TestBinaryParamsEncode(t *testing.T) {

	buf := &bytes.Buffer{}
	params := Params{}
	params["attachment"] = FileAlias("image.jpg", "LICENSE")

	contentTypeImage := "Content-Type: image/jpeg"
	if mime, err := params.Encode(buf); err != nil || !strings.Contains(mime, _MIME_FORM_DATA) || !strings.Contains(buf.String(), contentTypeImage) {
		t.Fatalf("wrong binary params encode result. expected content type is '%v'. actual is '%v'. [e:%v] [mime:%v]", contentTypeImage, buf.String(), err, mime)
	}

	// Fallback for unknown content types
	// should be application/octet-stream
	buf.Reset()
	params = Params{"attachment": FileAlias("image.unknown", "LICENSE")}
	contentTypeOctet := "Content-Type: application/octet-stream"
	if mime, err := params.Encode(buf); err != nil || !strings.Contains(mime, _MIME_FORM_DATA) || !strings.Contains(buf.String(), contentTypeOctet) {
		t.Fatalf("wrong binary params encode result. expected content type is '%v'. actual is '%v'. [e:%v] [mime:%v]", contentTypeOctet, buf.String(), err, mime)
	}
}
