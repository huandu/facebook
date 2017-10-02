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

func TestApiGetUserInfoV2(t *testing.T) {
	Version = "v2.2"
	defer func() {
		Version = ""
	}()

	// It's not allowed to get user info by name. So I get "me" with access token instead.
	if FB_TEST_VALID_ACCESS_TOKEN != "" {
		me, err := Api("me", GET, Params{
			"access_token": FB_TEST_VALID_ACCESS_TOKEN,
		})

		if err != nil {
			t.Fatalf("cannot get my info. [e:%v]", err)
		}

		if e := me.Err(); e != nil {
			t.Fatalf("facebook returns error. [e:%v]", e)
		}

		t.Logf("my info. %v", me)
	}
}

func TestBatchApiGetInfo(t *testing.T) {
	if FB_TEST_VALID_ACCESS_TOKEN == "" {
		t.Skipf("cannot call batch api without access token. skip this test.")
	}

	verifyBatchResult := func(t *testing.T, index int, res Result) {
		batch, err := res.Batch()

		if err != nil {
			t.Fatalf("cannot parse batch api results[%v]. [e:%v] [result:%v]", index, err, res)
		}

		if batch.StatusCode != 200 {
			t.Fatalf("facebook returns unexpected http status code in results[%v]. [code:%v] [result:%v]", index, batch.StatusCode, res)
		}

		contentType := batch.Header.Get("Content-Type")

		if contentType == "" {
			t.Fatalf("facebook returns unexpected http header in results[%v]. [header:%v]", index, batch.Header)
		}

		if batch.Body == "" {
			t.Fatalf("facebook returns unexpected http body in results[%v]. [body:%v]", index, batch.Body)
		}

		var id string
		err = batch.Result.DecodeField("id", &id)

		if err != nil {
			t.Fatalf("cannot get 'id' field in results[%v]. [result:%v]", index, res)
		}

		if id == "" {
			t.Fatalf("facebook should return account id in results[%v].", index)
		}
	}

	test := func(t *testing.T) {
		params1 := Params{
			"method":       GET,
			"relative_url": "me",
		}
		params2 := Params{
			"method":       GET,
			"relative_url": uint64(100002828925788), // id of my another facebook account
		}

		results, err := BatchApi(FB_TEST_VALID_ACCESS_TOKEN, params1, params2)

		if err != nil {
			t.Fatalf("cannot get batch result. [e:%v]", err)
		}

		if len(results) != 2 {
			t.Fatalf("batch api should return results in an array with 2 entries. [len:%v]", len(results))
		}

		if Version == "" {
			t.Log("use default facebook version.")
		} else {
			t.Logf("global facebook version: %v", Version)
		}

		for index, result := range results {
			verifyBatchResult(t, index, result)
		}
	}

	// Use default Version.
	Version = ""
	test(t)

	// User "v2.2".
	Version = "v2.2"
	defer func() {
		Version = ""
	}()
	test(t)

	// when providing an invalid access token, BatchApi should return a facebook error.
	_, err := BatchApi("an_invalid_access_token", Params{
		"method":       GET,
		"relative_url": "me",
	})

	if err == nil {
		t.Fatalf("expect an error when providing an invalid access token to BatchApi.")
	}

	if _, ok := err.(*Error); !ok {
		t.Fatalf("batch result error must be an *Error. [e:%v]", err)
	}
}

func TestGraphError(t *testing.T) {
	res, err := Get("/me", Params{
		"access_token": "fake",
	})

	if err == nil {
		t.Fatalf("facebook should return error for bad access token. [res:%v]", res)
	}

	fbErr, ok := err.(*Error)

	if !ok {
		t.Fatalf("error must be a *Error. [e:%v]", err)
	}

	t.Logf("facebook error. [e:%v] [message:%v] [type:%v] [code:%v] [subcode:%v]", err, fbErr.Message, fbErr.Type, fbErr.Code, fbErr.ErrorSubcode)
}
