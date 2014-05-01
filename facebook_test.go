// A facebook graph api client in go.
// https://github.com/huandu/facebook/
// 
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
    "bytes"
    "encoding/base64"
    "encoding/json"
    "testing"
)

const (
    FB_TEST_APP_ID      = "169186383097898"
    FB_TEST_APP_SECRET  = "b2e4262c306caa3c7f5215d2d099b319"
    FB_TEST_MY_USERNAME = "huan.du"

    // remeber to change it to a valid token to run test
    //FB_TEST_VALID_ACCESS_TOKEN = "CAACZA38ZAD8CoBAItCaMZAZCIMZAl1HWRDEDZAhLrce1X9IHzl6slmPSdMeZCyT45p71gsOuAVB5fZAcNUcrp6eZAXDguaFZAjNDbnfpY1m5f942cnI3ZAATOgJORWoDjRB6u7vb04ZC8oAu2A6kzKl1EfxrZBg4NhZAvINrYdv9F79dZCsOzTPJNQekczMz0rIvdBEpKwZD"
    FB_TEST_VALID_ACCESS_TOKEN = ""

    // remember to change it to a valid signed request to run test
    //FB_TEST_VALID_SIGNED_REQUEST = "ZAxP-ILRQBOwKKxCBMNlGmVraiowV7WFNg761OYBNGc.eyJhbGdvcml0aG0iOiJITUFDLVNIQTI1NiIsImV4cGlyZXMiOjEzNDM0OTg0MDAsImlzc3VlZF9hdCI6MTM0MzQ5MzI2NSwib2F1dGhfdG9rZW4iOiJBQUFDWkEzOFpBRDhDb0JBRFpCcmZ5TFpDanBNUVczdThVTWZmRldSWkNpZGw5Tkx4a1BsY2tTcXZaQnpzTW9OWkF2bVk2RUd2NG1hUUFaQ0t2VlpBWkJ5VXA5a0FCU2x6THFJejlvZTdOdHBzdzhyQVpEWkQiLCJ1c2VyIjp7ImNvdW50cnkiOiJ1cyIsImxvY2FsZSI6ImVuX1VTIiwiYWdlIjp7Im1pbiI6MjF9fSwidXNlcl9pZCI6IjUzODc0NDQ2OCJ9"
    FB_TEST_VALID_SIGNED_REQUEST = ""

    // test binary file base64 value
    FB_TEST_BINARY_JPG_FILE = "/9j/4AAQSkZJRgABAQEASABIAAD/4gv4SUNDX1BST0ZJTEUAAQEAAAvoAAAAAAIAAABtbnRy" +
        "UkdCIFhZWiAH2QADABsAFQAkAB9hY3NwAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAA" +
        "9tYAAQAAAADTLQAAAAAp+D3er/JVrnhC+uTKgzkNAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
        "AAAAAAAAABBkZXNjAAABRAAAAHliWFlaAAABwAAAABRiVFJDAAAB1AAACAxkbWRkAAAJ4AAA" +
        "AIhnWFlaAAAKaAAAABRnVFJDAAAB1AAACAxsdW1pAAAKfAAAABRtZWFzAAAKkAAAACRia3B0" +
        "AAAKtAAAABRyWFlaAAAKyAAAABRyVFJDAAAB1AAACAx0ZWNoAAAK3AAAAAx2dWVkAAAK6AAA" +
        "AId3dHB0AAALcAAAABRjcHJ0AAALhAAAADdjaGFkAAALvAAAACxkZXNjAAAAAAAAAB9zUkdC" +
        "IElFQzYxOTY2LTItMSBibGFjayBzY2FsZWQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
        "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
        "WFlaIAAAAAAAACSgAAAPhAAAts9jdXJ2AAAAAAAABAAAAAAFAAoADwAUABkAHgAjACgALQAy" +
        "ADcAOwBAAEUASgBPAFQAWQBeAGMAaABtAHIAdwB8AIEAhgCLAJAAlQCaAJ8ApACpAK4AsgC3" +
        "ALwAwQDGAMsA0ADVANsA4ADlAOsA8AD2APsBAQEHAQ0BEwEZAR8BJQErATIBOAE+AUUBTAFS" +
        "AVkBYAFnAW4BdQF8AYMBiwGSAZoBoQGpAbEBuQHBAckB0QHZAeEB6QHyAfoCAwIMAhQCHQIm" +
        "Ai8COAJBAksCVAJdAmcCcQJ6AoQCjgKYAqICrAK2AsECywLVAuAC6wL1AwADCwMWAyEDLQM4" +
        "A0MDTwNaA2YDcgN+A4oDlgOiA64DugPHA9MD4APsA/kEBgQTBCAELQQ7BEgEVQRjBHEEfgSM" +
        "BJoEqAS2BMQE0wThBPAE/gUNBRwFKwU6BUkFWAVnBXcFhgWWBaYFtQXFBdUF5QX2BgYGFgYn" +
        "BjcGSAZZBmoGewaMBp0GrwbABtEG4wb1BwcHGQcrBz0HTwdhB3QHhgeZB6wHvwfSB+UH+AgL" +
        "CB8IMghGCFoIbgiCCJYIqgi+CNII5wj7CRAJJQk6CU8JZAl5CY8JpAm6Cc8J5Qn7ChEKJwo9" +
        "ClQKagqBCpgKrgrFCtwK8wsLCyILOQtRC2kLgAuYC7ALyAvhC/kMEgwqDEMMXAx1DI4MpwzA" +
        "DNkM8w0NDSYNQA1aDXQNjg2pDcMN3g34DhMOLg5JDmQOfw6bDrYO0g7uDwkPJQ9BD14Peg+W" +
        "D7MPzw/sEAkQJhBDEGEQfhCbELkQ1xD1ERMRMRFPEW0RjBGqEckR6BIHEiYSRRJkEoQSoxLD" +
        "EuMTAxMjE0MTYxODE6QTxRPlFAYUJxRJFGoUixStFM4U8BUSFTQVVhV4FZsVvRXgFgMWJhZJ" +
        "FmwWjxayFtYW+hcdF0EXZReJF64X0hf3GBsYQBhlGIoYrxjVGPoZIBlFGWsZkRm3Gd0aBBoq" +
        "GlEadxqeGsUa7BsUGzsbYxuKG7Ib2hwCHCocUhx7HKMczBz1HR4dRx1wHZkdwx3sHhYeQB5q" +
        "HpQevh7pHxMfPh9pH5Qfvx/qIBUgQSBsIJggxCDwIRwhSCF1IaEhziH7IiciVSKCIq8i3SMK" +
        "IzgjZiOUI8Ij8CQfJE0kfCSrJNolCSU4JWgllyXHJfcmJyZXJocmtyboJxgnSSd6J6sn3CgN" +
        "KD8ocSiiKNQpBik4KWspnSnQKgIqNSpoKpsqzysCKzYraSudK9EsBSw5LG4soizXLQwtQS12" +
        "Last4S4WLkwugi63Lu4vJC9aL5Evxy/+MDUwbDCkMNsxEjFKMYIxujHyMioyYzKbMtQzDTNG" +
        "M38zuDPxNCs0ZTSeNNg1EzVNNYc1wjX9Njc2cjauNuk3JDdgN5w31zgUOFA4jDjIOQU5Qjl/" +
        "Obw5+To2OnQ6sjrvOy07azuqO+g8JzxlPKQ84z0iPWE9oT3gPiA+YD6gPuA/IT9hP6I/4kAj" +
        "QGRApkDnQSlBakGsQe5CMEJyQrVC90M6Q31DwEQDREdEikTORRJFVUWaRd5GIkZnRqtG8Ec1" +
        "R3tHwEgFSEtIkUjXSR1JY0mpSfBKN0p9SsRLDEtTS5pL4kwqTHJMuk0CTUpNk03cTiVObk63" +
        "TwBPSU+TT91QJ1BxULtRBlFQUZtR5lIxUnxSx1MTU19TqlP2VEJUj1TbVShVdVXCVg9WXFap" +
        "VvdXRFeSV+BYL1h9WMtZGllpWbhaB1pWWqZa9VtFW5Vb5Vw1XIZc1l0nXXhdyV4aXmxevV8P" +
        "X2Ffs2AFYFdgqmD8YU9homH1YklinGLwY0Njl2PrZEBklGTpZT1lkmXnZj1mkmboZz1nk2fp" +
        "aD9olmjsaUNpmmnxakhqn2r3a09rp2v/bFdsr20IbWBtuW4SbmtuxG8eb3hv0XArcIZw4HE6" +
        "cZVx8HJLcqZzAXNdc7h0FHRwdMx1KHWFdeF2Pnabdvh3VnezeBF4bnjMeSp5iXnnekZ6pXsE" +
        "e2N7wnwhfIF84X1BfaF+AX5ifsJ/I3+Ef+WAR4CogQqBa4HNgjCCkoL0g1eDuoQdhICE44VH" +
        "hauGDoZyhteHO4efiASIaYjOiTOJmYn+imSKyoswi5aL/IxjjMqNMY2Yjf+OZo7OjzaPnpAG" +
        "kG6Q1pE/kaiSEZJ6kuOTTZO2lCCUipT0lV+VyZY0lp+XCpd1l+CYTJi4mSSZkJn8mmia1ZtC" +
        "m6+cHJyJnPedZJ3SnkCerp8dn4uf+qBpoNihR6G2oiailqMGo3aj5qRWpMelOKWpphqmi6b9" +
        "p26n4KhSqMSpN6mpqhyqj6sCq3Wr6axcrNCtRK24ri2uoa8Wr4uwALB1sOqxYLHWskuywrM4" +
        "s660JbSctRO1irYBtnm28Ldot+C4WbjRuUq5wro7urW7LrunvCG8m70VvY++Cr6Evv+/er/1" +
        "wHDA7MFnwePCX8Lbw1jD1MRRxM7FS8XIxkbGw8dBx7/IPci8yTrJuco4yrfLNsu2zDXMtc01" +
        "zbXONs62zzfPuNA50LrRPNG+0j/SwdNE08bUSdTL1U7V0dZV1tjXXNfg2GTY6Nls2fHadtr7" +
        "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"
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

type FieldTagStruct struct {
    Field1    string `facebook:"field2"`
    Required  string `facebook:",required"`
    Foo       string `facebook:"bar,required"`
    CanAbsent string
}

func TestApiGetUserInfo(t *testing.T) {
    me, err := Api(FB_TEST_MY_USERNAME, GET, nil)

    if err != nil {
        t.Errorf("cannot get my info. [e:%v]", err)
        return
    }

    if e := me.Err(); e != nil {
        t.Errorf("facebook returns error. [e:%v]", e)
        return
    }

    t.Logf("my info. %v", me)
}

func TestApiGetUserInfoV2(t *testing.T) {
    Version = "v2.0"
    defer func() {
        Version = ""
    }()

    // It's not allowed to get user info by name. So I get "me" with access token instead.
    if FB_TEST_VALID_ACCESS_TOKEN != "" {
        me, err := Api("me", GET, Params{
            "access_token": FB_TEST_VALID_ACCESS_TOKEN,
        })

        if err != nil {
            t.Errorf("cannot get my info. [e:%v]", err)
            return
        }

        if e := me.Err(); e != nil {
            t.Errorf("facebook returns error. [e:%v]", e)
            return
        }

        t.Logf("my info. %v", me)
    }
}

func TestBatchApiGetInfo(t *testing.T) {
    if FB_TEST_VALID_ACCESS_TOKEN == "" {
        t.Logf("cannot call batch api without access token. skip this test.")
        return
    }

    test := func(t *testing.T) {
        params1 := Params{
            "method":       GET,
            "relative_url": FB_TEST_MY_USERNAME,
        }
        params2 := Params{
            "method":       GET,
            "relative_url": uint64(100002828925788), // id of my another facebook id
        }

        me, err := BatchApi(FB_TEST_VALID_ACCESS_TOKEN, params1, params2)

        if err != nil {
            t.Errorf("cannot get batch result. [e:%v]", err)
            return
        }

        if Version == "" {
            t.Log("use default facebook version.")
        } else {
            t.Logf("global facebook version: %v", Version)
        }

        t.Logf("my info. %v", me)
    }

    // Use default Version.
    Version = ""
    test(t)

    // User "v2.0".
    Version = "v2.0"
    defer func() {
        Version = ""
    }()
    test(t)
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

    test := func(t *testing.T, session *Session) {
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
        Version = "v2.0"
        defer func() {
            Version = ""
        }()

        test(t, session)
    }()

    // Session version overwrite default version.
    func() {
        Version = "vx.y" // an invalid version.
        session.Version = "v2.0"
        defer func() {
            Version = ""
        }()

        test(t, session)
    }()
}

func TestUploadingBinary(t *testing.T) {
    if FB_TEST_VALID_ACCESS_TOKEN == "" {
        t.Logf("skip this case as we don't have a valid access token.")
        return
    }

    buf := bytes.NewBufferString(FB_TEST_BINARY_JPG_FILE)
    reader := base64.NewDecoder(base64.StdEncoding, buf)

    session := &Session{}
    session.SetAccessToken(FB_TEST_VALID_ACCESS_TOKEN)

    result, e := session.Api("/me/photos", POST, Params{
        "message": "Test photo from https://github.com/huandu/facebook",
        "source":  Data("my_profile.jpg", reader),
    })

    if e != nil {
        t.Errorf("cannot create photo on my timeline. [e:%v]", e)
        return
    }

    var id string
    e = result.DecodeField("id", &id)

    if e != nil {
        t.Errorf("facebook should return photo id on success. [e:%v]", e)
        return
    }

    t.Logf("newly created photo id is %v", id)
}

func TestUploadBinaryWithBatch(t *testing.T) {
    if FB_TEST_VALID_ACCESS_TOKEN == "" {
        t.Logf("skip this case as we don't have a valid access token.")
        return
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
        t.Errorf("cannot create photo on my timeline. [e:%v]", e)
        return
    }

    t.Logf("batch call result. [result:%v]", result)
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

    err = result.DecodeField("nested_struct.int", &anInt)

    if err != nil {
        t.Errorf("cannot decode nested struct item. [e:%v]", err)
        return
    }

    if anInt != 123 {
        t.Errorf("invalid array value. expected 123, actual %v", anInt)
        return
    }
}

func TestParamsEncode(t *testing.T) {
    var params Params
    buf := &bytes.Buffer{}

    if mime, err := params.Encode(buf); err != nil || mime != _MIME_FORM_URLENCODED || buf.Len() != 0 {
        t.Errorf("empty params must encode to an empty string. actual is [e:%v] [str:%v] [mime:%v]", err, buf.String(), mime)
        return
    }

    buf.Reset()
    params = Params{}
    params["need_escape"] = "&=+"
    expectedEncoding := "need_escape=%26%3D%2B"

    if mime, err := params.Encode(buf); err != nil || mime != _MIME_FORM_URLENCODED || buf.String() != expectedEncoding {
        t.Errorf("wrong params encode result. expected is '%v'. actual is '%v'. [e:%v] [mime:%v]", expectedEncoding, buf.String(), err, mime)
        return
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
        t.Errorf("make params error.")
        return
    }

    mime, err := params.Encode(buf)
    t.Logf("complex encode result is '%v'. [e:%v] [mime:%v]", buf.String(), err, mime)
}

func TestStructFieldTag(t *testing.T) {
    strNormalField := `{
        "field2": "hey",
        "required": "my",
        "bar": "dear"
    }`
    strMissingField2Field := `{
        "field1": "hey",
        "required": "my",
        "bar": "dear"
    }`
    strMissingRequiredField := `{
        "field1": "hey",
        "bar": "dear",
        "can_absent": "babe"
    }`
    strMissingBarField := `{
        "field1": "hey",
        "required": "my"
    }`

    var result Result
    var value FieldTagStruct
    var err error

    err = json.Unmarshal([]byte(strNormalField), &result)

    if err != nil {
        t.Errorf("cannot unmarshal json string. [e:%v]", err)
        return
    }

    err = result.Decode(&value)

    if err != nil {
        t.Errorf("cannot decode struct. [e:%v]", err)
        return
    }

    result = Result{}
    value = FieldTagStruct{}
    err = json.Unmarshal([]byte(strMissingField2Field), &result)

    if err != nil {
        t.Errorf("cannot unmarshal json string. [e:%v]", err)
        return
    }

    err = result.Decode(&value)

    if err != nil {
        t.Errorf("cannot decode struct. [e:%v]", err)
        return
    }

    if value.Field1 != "" {
        t.Errorf("value field1 should be kept unchanged. [field1:%v]", value.Field1)
        return
    }

    result = Result{}
    value = FieldTagStruct{}
    err = json.Unmarshal([]byte(strMissingRequiredField), &result)

    if err != nil {
        t.Errorf("cannot unmarshal json string. [e:%v]", err)
        return
    }

    err = result.Decode(&value)

    if err == nil {
        t.Errorf("should fail to decode struct.")
        return
    }

    t.Logf("expected decode error. [e:%v]", err)

    result = Result{}
    value = FieldTagStruct{}
    err = json.Unmarshal([]byte(strMissingBarField), &result)

    if err != nil {
        t.Errorf("cannot unmarshal json string. [e:%v]", err)
        return
    }

    err = result.Decode(&value)

    if err == nil {
        t.Errorf("should fail to decode struct.")
        return
    }

    t.Logf("expected decode error. [e:%v]", err)
}

func TestGraphError(t *testing.T) {
    res, err := Get("/me", Params{
        "access_token": "fake",
    })

    if err == nil {
        t.Errorf("facebook should return error for bad access token. [res:%v]", res)
        return
    }

    fbErr, ok := err.(*Error)

    if !ok {
        t.Errorf("error must be a *Error. [e:%v]", err)
        return
    }

    t.Logf("facebook error. [e:%v] [message:%v] [type:%v] [code:%v] [subcode:%v]", err, fbErr.Message, fbErr.Type, fbErr.Code, fbErr.ErrorSubcode)
}
