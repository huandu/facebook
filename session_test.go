// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSession(t *testing.T) {
	if FB_TEST_VALID_ACCESS_TOKEN == "" {
		t.Skipf("skip this case as we don't have a valid access token.")
	}

	session := &Session{}
	session.SetAccessToken(FB_TEST_VALID_ACCESS_TOKEN)

	test := func(t *testing.T, session *Session) {
		id, err := session.User()

		if err != nil {
			t.Fatalf("cannot get current user id. [e:%v]", err)
		}

		t.Logf("current user id is %v", id)

		result, e := session.Api("/me", GET, Params{
			"fields": "id,email,website",
		})

		if e != nil {
			t.Fatalf("cannot get my extended info. [e:%v]", e)
		}

		if Version == "" {
			t.Log("use default facebook version.")
		} else {
			t.Logf("global facebook version: %v", Version)
		}

		if session.Version == "" {
			t.Log("use default session facebook version.")
		} else {
			t.Logf("session facebook version: %v", session.Version)
		}

		t.Logf("my extended info is: %v", result)
	}

	// Default version.
	test(t, session)

	// Global version overwrite default session version.
	func() {
		Version = FB_LATEST_VERSION
		defer func() {
			Version = ""
		}()

		test(t, session)
	}()

	// Session version overwrite default version.
	func() {
		Version = "vx.y" // an invalid version.
		session.Version = FB_LATEST_VERSION
		defer func() {
			Version = ""
		}()

		test(t, session)
	}()

	// Session with appsecret proof enabled.
	if FB_TEST_VALID_ACCESS_TOKEN != "" {
		app := New(FB_TEST_APP_ID, FB_TEST_APP_SECRET)
		app.EnableAppsecretProof = true
		session := app.Session(FB_TEST_VALID_ACCESS_TOKEN)

		_, e := session.Api("/me", GET, Params{
			"fields": "id",
		})

		if e != nil {
			t.Fatalf("cannot get my info with proof. [e:%v]", e)
		}
	}
}

func TestUploadingBinary(t *testing.T) {
	if FB_TEST_VALID_ACCESS_TOKEN == "" {
		t.Skipf("skip this case as we don't have a valid access token.")
	}

	t.Skipf("facebook doesn't support uploading photo to timeline.")

	buf := bytes.NewBufferString(FB_TEST_BINARY_JPG_FILE)
	reader := base64.NewDecoder(base64.StdEncoding, buf)

	session := &Session{}
	session.SetAccessToken(FB_TEST_VALID_ACCESS_TOKEN)

	result, e := session.Api("/317426148998929/photos", POST, Params{
		"caption":   "Test photo from https://github.com/huandu/facebook",
		"source":    Data("attachment.jpg", reader),
		"published": true,
	})

	if e != nil {
		t.Fatalf("cannot create photo on page timeline. [e:%v]", e)
	}

	var id string
	e = result.DecodeField("id", &id)

	if e != nil {
		t.Fatalf("facebook should return photo id on success. [e:%v]", e)
	}

	t.Logf("newly created photo id is %v", id)
}

func TestUploadBinaryWithBatch(t *testing.T) {
	if FB_TEST_VALID_ACCESS_TOKEN == "" {
		t.Skipf("skip this case as we don't have a valid access token.")
	}

	buf1 := bytes.NewBufferString(FB_TEST_BINARY_JPG_FILE)
	reader1 := base64.NewDecoder(base64.StdEncoding, buf1)
	buf2 := bytes.NewBufferString(FB_TEST_BINARY_JPG_FILE)
	reader2 := base64.NewDecoder(base64.StdEncoding, buf2)

	session := &Session{}
	session.SetAccessToken(FB_TEST_VALID_ACCESS_TOKEN)

	// sample comes from facebook batch api sample.
	// https://developers.facebook.com/docs/reference/api/batch/
	//
	// curl
	//     -F 'access_token=…' \
	//     -F 'batch=[{"method":"POST","relative_url":"me/photos","body":"message=My cat photo","attached_files":"file1"},{"method":"POST","relative_url":"me/photos","body":"message=My dog photo","attached_files":"file2"},]' \
	//     -F 'file1=@cat.gif' \
	//     -F 'file2=@dog.jpg' \
	//         https://graph.facebook.com
	result, e := session.Batch(Params{
		"file1": Data("cat.jpg", reader1),
		"file2": Data("dog.jpg", reader2),
	}, Params{
		"method":         POST,
		"relative_url":   "me/photos",
		"body":           "message=My cat photo",
		"attached_files": "file1",
	}, Params{
		"method":         POST,
		"relative_url":   "me/photos",
		"body":           "message=My dog photo",
		"attached_files": "file2",
	})

	if e != nil {
		t.Fatalf("cannot create photo on my timeline. [e:%v]", e)
	}

	t.Logf("batch call result. [result:%v]", result)
}

func TestGraphDebuggingAPI(t *testing.T) {
	if FB_TEST_VALID_ACCESS_TOKEN == "" {
		t.Skipf("cannot call batch api without access token. skip this test.")
	}

	test := func(t *testing.T, session *Session) {
		session.SetAccessToken(FB_TEST_VALID_ACCESS_TOKEN)
		defer session.SetAccessToken("")

		// test app must not grant "read_friends" permission.
		// otherwise there is no way to get a warning from facebook.
		res, _ := session.Get("/me/friends", nil)

		if res == nil {
			t.Fatalf("res must not be nil.")
		}

		debugInfo := res.DebugInfo()

		if debugInfo == nil {
			t.Fatalf("debug info must exist.")
		}

		t.Logf("facebook response is: %v", res)
		t.Logf("debug info is: %v", *debugInfo)

		if len(debugInfo.Messages) == 0 {
			t.Fatalf("facebook must warn me for the permission issue.")
		}

		msg := debugInfo.Messages[0]

		if msg.Type == "" || msg.Message == "" {
			t.Fatalf("facebook must say something. [msg:%v]", msg)
		}

		if debugInfo.FacebookApiVersion == "" {
			t.Fatalf("facebook must tell me api version.")
		}

		if debugInfo.FacebookDebug == "" {
			t.Fatalf("facebook must tell me X-FB-Debug.")
		}

		if debugInfo.FacebookRev == "" {
			t.Fatalf("facebook must tell me x-fb-rev.")
		}
	}

	defer func() {
		Debug = DEBUG_OFF
		Version = ""
	}()

	Version = FB_LATEST_VERSION
	Debug = DEBUG_ALL
	test(t, defaultSession)
	session := &Session{}
	session.SetDebug(DEBUG_ALL)
	test(t, session)

	// test changing debug mode.
	old := session.SetDebug(DEBUG_OFF)

	if old != DEBUG_ALL {
		t.Fatalf("debug mode must be DEBUG_ALL. [debug:%v]", old)
	}

	if session.Debug() != DEBUG_ALL {
		t.Fatalf("debug mode must be DEBUG_ALL [debug:%v]", session.Debug())
	}

	Debug = DEBUG_OFF

	if session.Debug() != DEBUG_OFF {
		t.Fatalf("debug mode must be DEBUG_OFF. [debug:%v]", session.Debug())
	}
}

func TestInspectValidToken(t *testing.T) {
	if FB_TEST_VALID_ACCESS_TOKEN == "" {
		t.Skipf("skip this case as we don't have a valid access token.")
	}

	session := testGlobalApp.Session(FB_TEST_VALID_ACCESS_TOKEN)
	result, err := session.Inspect()

	if err != nil {
		t.Fatalf("cannot inspect a valid access token. [e:%v]", err)
	}

	var isValid bool
	err = result.DecodeField("is_valid", &isValid)

	if err != nil {
		t.Fatalf("cannot get 'is_valid' in inspect result. [e:%v]", err)
	}

	if !isValid {
		t.Fatalf("inspect result shows access token is invalid. why? [result:%v]", result)
	}
}

func TestInspectInvalidToken(t *testing.T) {
	invalidToken := "CAACZA38ZAD8CoBAe2bDC6EdThnni3b56scyshKINjZARoC9ZAuEUTgYUkYnKdimqfA2ZAXcd2wLd7Rr8jLmMXTY9vqAhQGqObZBIUz1WwbqVoCsB3AAvLtwoWNhsxM76mK0eiJSLXHZCdPVpyhmtojvzXA7f69Bm6b5WZBBXia8iOpPZAUHTGp1UQLFMt47c7RqJTrYIl3VfAR0deN82GMFL2"
	session := testGlobalApp.Session(invalidToken)
	result, err := session.Inspect()

	if err == nil {
		t.Fatalf("facebook should indicate it's an invalid token. why not? [result:%v]", result)
	}

	if _, ok := err.(*Error); !ok {
		t.Fatalf("inspect error should be a standard facebook error. why not? [e:%v]", err)
	}

	isValid := true
	err = result.DecodeField("is_valid", &isValid)

	if err != nil {
		t.Fatalf("cannot get 'is_valid' in inspect result. [e:%v]", err)
	}

	if isValid {
		t.Fatalf("inspect result shows access token is valid. why? [result:%v]", result)
	}
}

func TestSessionCancelationWithContext(t *testing.T) {
	session := &Session{}
	ctx, cancel := context.WithCancel(context.Background())
	newSession := session.WithContext(ctx)

	if session == newSession {
		t.Fatalf("session.WithContext must return a new session instance.")
	}

	if session.Context() != context.Background() {
		t.Fatalf("default session context must be context.Background().")
	}

	if ctx != newSession.Context() {
		t.Fatalf("ctx is not set to new session.")
	}

	if FB_TEST_VALID_ACCESS_TOKEN == "" {
		t.Skipf("skip this case as we don't have a valid access token.")
	}

	cancel()
	newSession.SetAccessToken(FB_TEST_VALID_ACCESS_TOKEN)
	_, err := newSession.Inspect()

	if err == nil {
		t.Fatalf("http request must fail as cancelled.")
	}

	t.Logf("http request error should fail as cancelled. [e:%v]", err)
}

func TestInspectAppAccessToken(t *testing.T) {
	app := New(FB_TEST_APP_ID, FB_TEST_APP_SECRET)
	session := app.Session(app.AppAccessToken())

	_, err := session.Inspect()

	if err != nil {
		t.Fatalf("fail to inspect app access token. [e:%v]", err)
	}
}

func TestSessionWithCustomBaseUrl(t *testing.T) {
	testMux := http.NewServeMux()
	numCalls := 0
	testMux.HandleFunc("/v3.0/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name": "some-user"}`))
		numCalls++
	})

	srv := httptest.NewServer(testMux)
	defer srv.Close()

	session := &Session{
		Version: "v3.0",
		BaseURL: srv.URL + "/",
	}
	_, err := session.Get("/me", nil)
	if err != nil {
		t.Fatalf("request to custom base URL failed: %v", err)
	}
	if numCalls != 1 {
		t.Fatal("no call to mock server")
	}
}

func TestSessionGetWithQueryString(t *testing.T) {
	if FB_TEST_VALID_ACCESS_TOKEN == "" {
		t.Skipf("skip this case as we don't have a valid access token.")
	}

	session := &Session{}
	session.SetAccessToken(FB_TEST_VALID_ACCESS_TOKEN)

	id, err := session.User()

	if err != nil {
		t.Fatalf("cannot get current user id. [e:%v]", err)
	}

	t.Logf("current user id is %v", id)

	result, e := session.Api("me?fields=name,email", GET, Params{
		"fields": "id,name",
	})

	if e != nil {
		t.Fatalf("cannot get my extended info. [e:%v]", e)
	}

	if result.Get("name") == nil || result.Get("email") == nil {
		t.Fatalf("fail to get my extend info. [result:%v]", result)
	}

	t.Logf("my extended info is: %v", result)
}
