// A facebook graph api client in go.
// https://github.com/huandu/facebook/
// 
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE
package facebook

import (
    "encoding/json"
    "testing"
)

const (
    FB_TEST_APP_ID      = "169186383097898"
    FB_TEST_APP_SECRET  = "b2e4262c306caa3c7f5215d2d099b319"
    FB_TEST_MY_USERNAME = "huan.du"

    // remeber to change it to a valid token to run test
    //FB_TEST_VALID_ACCESS_TOKEN = "AAACEdEose0cBAEVR6GZAEcja9ZBIdmYrF4I7nAFFbYZAe9tdYEv6uZCHGPpPFcbNYo49ya6qmAChPUZBO2UYmotTdWiDZBMQZCGXm8lA9qjCQZDZD"
    FB_TEST_VALID_ACCESS_TOKEN = ""

    // remember to change it to a valid signed request to run test
    //FB_TEST_VALID_SIGNED_REQUEST = "ZAxP-ILRQBOwKKxCBMNlGmVraiowV7WFNg761OYBNGc.eyJhbGdvcml0aG0iOiJITUFDLVNIQTI1NiIsImV4cGlyZXMiOjEzNDM0OTg0MDAsImlzc3VlZF9hdCI6MTM0MzQ5MzI2NSwib2F1dGhfdG9rZW4iOiJBQUFDWkEzOFpBRDhDb0JBRFpCcmZ5TFpDanBNUVczdThVTWZmRldSWkNpZGw5Tkx4a1BsY2tTcXZaQnpzTW9OWkF2bVk2RUd2NG1hUUFaQ0t2VlpBWkJ5VXA5a0FCU2x6THFJejlvZTdOdHBzdzhyQVpEWkQiLCJ1c2VyIjp7ImNvdW50cnkiOiJ1cyIsImxvY2FsZSI6ImVuX1VTIiwiYWdlIjp7Im1pbiI6MjF9fSwidXNlcl9pZCI6IjUzODc0NDQ2OCJ9"
    FB_TEST_VALID_SIGNED_REQUEST = ""
)

type AllTypes struct {
    Int          int
    Int8         int8
    Int16        int16
    Int32        int32
    Int64        int64
    Uint         uint
    Uint8        uint8
    Uint16       uint16
    Uint32       uint32
    Uint64       uint64
    Float32      float32
    Float64      float64
    String       string
    ArrayOfInt   []int
    MapOfString  map[string]string
    NestedStruct *NestedStruct
}

type NestedStruct struct {
    Int           int
    String        string
    ArrayOfString []string
}

type ParamsStruct struct {
    Foo string
    Bar *ParamsNestedStruct
}

type ParamsNestedStruct struct {
    AAA int
    BBB string
    CCC bool
}

func TestApiGetUserInfo(t *testing.T) {
    me, err := Api(FB_TEST_MY_USERNAME, GET, nil)

    if err != nil {
        t.Errorf("cannot get my info. [e:%v]", err)
        return
    }

    t.Logf("my info. %v", me)
}

func TestBatchApiGetInfo(t *testing.T) {
    if FB_TEST_VALID_ACCESS_TOKEN == "" {
        t.Logf("cannot call batch api without access token. skip this test.")
        return
    }

    params1 := Params{
        "method": GET,
        "relative_url": FB_TEST_MY_USERNAME,
    }
    params2 := Params{
        "method": GET,
        "relative_url": uint64(100002828925788), // id of my another facebook id
    }

    me, err := BatchApi(FB_TEST_VALID_ACCESS_TOKEN, params1, params2)

    if err != nil {
        t.Errorf("cannot get batch result. [e:%v]", err)
        return
    }

    t.Logf("my info. %v", me)
}

func TestApiParseSignedRequest(t *testing.T) {
    if FB_TEST_VALID_SIGNED_REQUEST == "" {
        t.Logf("skip this case as we don't have a valid signed request.")
        return
    }

    app := New(FB_TEST_APP_ID, FB_TEST_APP_SECRET)
    res, err := app.ParseSignedRequest(FB_TEST_VALID_SIGNED_REQUEST)

    if err != nil {
        t.Errorf("cannot parse signed request. [e:%v]", err)
        return
    }

    t.Logf("signed request is '%v'.", res)
}

func TestSession(t *testing.T) {
    if FB_TEST_VALID_ACCESS_TOKEN == "" {
        t.Logf("skip this case as we don't have a valid access token.")
        return
    }

    session := &Session{}
    session.SetAccessToken(FB_TEST_VALID_ACCESS_TOKEN)
    id, err := session.User()

    if err != nil {
        t.Errorf("cannot get current user id. [e:%v]", err)
        return
    }

    t.Logf("current user id is %v", id)

    result, e := session.Api("/me", GET, Params{
        "fields": "id,email,website",
    })

    if e != nil {
        t.Errorf("cannot get my extended info. [e:%v]", e)
        return
    }

    t.Logf("my extended info is: %v", result)
}

func TestResultDecode(t *testing.T) {
    strNormal := `{
        "int": 1234,
        "int8": 23,
        "int16": 12345,
        "int32": -127372843,
        "int64": 192438483489298,
        "uint": 1283829,
        "uint8": 233,
        "uint16": 62121,
        "uint32": 3083747392,
        "uint64": 2034857382993849,
        "float32": 9382.38429,
        "float64": 3984.293848292,
        "map_of_string": {"a": "1", "b": "2"},
        "array_of_int": [12, 34, 56],
        "string": "abcd",
        "notused": 1234,
        "nested_struct": {
            "string": "hello",
            "int": 123,
            "array_of_string": ["a", "b", "c"]
        }
    }`
    strOverflow := `{
        "int": 1234,
        "int8": 23,
        "int16": 12345,
        "int32": -127372843,
        "int64": 192438483489298,
        "uint": 1283829,
        "uint8": 233,
        "uint16": 62121,
        "uint32": 383083747392,
        "uint64": 2034857382993849,
        "float32": 9382.38429,
        "float64": 3984.293848292,
        "string": "abcd",
        "map_of_string": {"a": "1", "b": "2"},
        "array_of_int": [12, 34, 56],
        "string": "abcd",
        "notused": 1234,
        "nested_struct": {
            "string": "hello",
            "int": 123,
            "array_of_string": ["a", "b", "c"]
        }
    }`
    strMissAField := `{
        "int": 1234,
        "int8": 23,
        "int16": 12345,
        "int32": -127372843,

        "missed": "int64",

        "uint": 1283829,
        "uint8": 233,
        "uint16": 62121,
        "uint32": 383083747392,
        "uint64": 2034857382993849,
        "float32": 9382.38429,
        "float64": 3984.293848292,
        "string": "abcd",
        "map_of_string": {"a": "1", "b": "2"},
        "array_of_int": [12, 34, 56],
        "string": "abcd",
        "notused": 1234,
        "nested_struct": {
            "string": "hello",
            "int": 123,
            "array_of_string": ["a", "b", "c"]
        }
    }`
    var result Result
    var err error
    var normal, withError AllTypes
    var anInt int

    err = json.Unmarshal([]byte(strNormal), &result)

    if err != nil {
        t.Errorf("cannot unmarshal json string. [e:%v]", err)
        return
    }

    err = result.Decode(&normal)

    if err != nil {
        t.Errorf("cannot decode normal struct. [e:%v]", err)
        return
    }

    err = json.Unmarshal([]byte(strOverflow), &result)

    if err != nil {
        t.Errorf("cannot unmarshal json string. [e:%v]", err)
        return
    }

    err = result.Decode(&withError)

    if err == nil {
        t.Errorf("struct should be overflow")
        return
    }

    t.Logf("overflow struct. e:%v", err)

    err = json.Unmarshal([]byte(strMissAField), &result)

    if err != nil {
        t.Errorf("cannot unmarshal json string. [e:%v]", err)
        return
    }

    err = result.Decode(&withError)

    if err == nil {
        t.Errorf("a field in struct should absent in json map.")
        return
    }

    t.Logf("miss-a-field struct. e:%v", err)

    err = result.DecodeField("array_of_int.2", &anInt)

    if err != nil {
        t.Errorf("cannot decode array item. [e:%v]", err)
        return
    }

    if anInt != 56 {
        t.Errorf("invalid array value. expected 56, actual %v", anInt)
        return
    }
}

func TestParamsEncode(t *testing.T) {
    var params Params

    if params.Encode() != "" {
        t.Errorf("empty params must encode to an empty string. actual is '%v'.", params.Encode())
        return
    }

    params = Params{}
    params["need_escape"] = "&=+"
    expectedEncoding := "need_escape=%26%3D%2B"

    if params.Encode() != expectedEncoding {
        t.Errorf("wrong params encode result. expected is '%v'. actual is '%v'.", expectedEncoding, params.Encode())
        return
    }

    data := ParamsStruct{
        Foo: "hello, world!",
        Bar: &ParamsNestedStruct{
            AAA: 1234,
            BBB: "bbb",
            CCC: true,
        },
    }
    params = MakeParams(data)
    /* there is no easy way to compare two maps. so i just write expect map here, not test it.
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
        t.Errorf("make params error.")
        return
    }

    t.Logf("complex encode result is '%v'.", params.Encode())
}
