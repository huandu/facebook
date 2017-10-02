// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2015, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"testing"
)

func TestApiParseSignedRequest(t *testing.T) {
	if FB_TEST_VALID_SIGNED_REQUEST == "" {
		t.Logf("skip this case as we don't have a valid signed request.")
		return
	}

	app := New(FB_TEST_APP_ID, FB_TEST_APP_SECRET)
	res, err := app.ParseSignedRequest(FB_TEST_VALID_SIGNED_REQUEST)

	if err != nil {
		t.Fatalf("cannot parse signed request. [e:%v]", err)
	}

	t.Logf("signed request is '%v'.", res)
}
