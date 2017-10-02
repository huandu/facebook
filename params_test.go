// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2015, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"bytes"
	"testing"
)

func TestParamsEncode(t *testing.T) {
	var params Params
	buf := &bytes.Buffer{}

	if mime, err := params.Encode(buf); err != nil || mime != _MIME_FORM_URLENCODED || buf.Len() != 0 {
		t.Fatalf("empty params must encode to an empty string. actual is [e:%v] [str:%v] [mime:%v]", err, buf.String(), mime)
	}

	buf.Reset()
	params = Params{}
	params["need_escape"] = "&=+"
	expectedEncoding := "need_escape=%26%3D%2B"

	if mime, err := params.Encode(buf); err != nil || mime != _MIME_FORM_URLENCODED || buf.String() != expectedEncoding {
		t.Fatalf("wrong params encode result. expected is '%v'. actual is '%v'. [e:%v] [mime:%v]", expectedEncoding, buf.String(), err, mime)
	}

	buf.Reset()
	data := ParamsStruct{
		Foo: "hello, world!",
		Bar: &ParamsNestedStruct{
			AAA: 1234,
			BBB: "bbb",
			CCC: true,
		},
	}
	params = MakeParams(data)
	/* there is no easy way to compare two encoded maps. so i just write expect map here, not test it.
	   expectedParams := Params{
	       "foo": "hello, world!",
	       "bar": map[string]interface{}{
	           "aaa": 1234,
	           "bbb": "bbb",
	           "ccc": true,
	       },
	   }
	*/

	if params == nil {
		t.Fatalf("make params error.")
	}

	mime, err := params.Encode(buf)
	t.Logf("complex encode result is '%v'. [e:%v] [mime:%v]", buf.String(), err, mime)
}
