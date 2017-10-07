// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2015, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type AllTypes struct {
	AnonymousStruct1
	*AnonymousStruct2

	Int           int
	Int8          int8
	Int16         int16
	Int32         int32
	Int64         int64
	Uint          uint
	Uint8         uint8
	Uint16        uint16
	Uint32        uint32
	Uint64        uint64
	Float32       float32
	Float64       float64
	StringInt     Int
	StringInt8    Int8
	StringInt16   Int16
	StringInt32   Int32
	StringInt64   Int64
	StringUint    Uint
	StringUint8   Uint8
	StringUint16  Uint16
	StringUint32  Uint32
	StringUint64  Uint64
	StringFloat32 Float32
	StringFloat64 Float64
	String        string
	ArrayOfInt    []int
	MapOfString   map[string]string
	NestedStruct  *NestedStruct
}

type AnonymousStruct1 struct {
	AnonymousInt1           int
	AnonymousString1        string
	AnonymousArrayOfString1 []string
}

type AnonymousStruct2 struct {
	AnonymousInt2           int
	AnonymousString2        string
	AnonymousArrayOfString2 []string
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

type MessageTag struct {
	Id   string
	Name string
	Type string
}

type MessageTags map[string][]*MessageTag

type NullStruct struct {
	Null *int
}

// custom unmarshaler.
type CustomMarshaler struct {
	Name  string
	Hash  int64
	Extra bool `facebook:",required"` // this field is ignored due to Unmarshaler.
}

type customMarshaler struct {
	Name string
	Hash int64
}

func (cm *CustomMarshaler) UnmarshalJSON(data []byte) error {
	var c customMarshaler
	err := json.Unmarshal(data, &c)

	if err != nil {
		return err
	}

	cm.Name = c.Name
	cm.Hash = c.Hash

	if cm.Name == "bar" {
		return fmt.Errorf("sorry but i don't like `bar`.")
	}

	return nil
}

type CustomMarshalerStruct struct {
	Marshaler CustomMarshaler
}

func TestResultDecode(t *testing.T) {
	strNormal := `{
		"anonymous_int1": 123,
		"anonymous_string2": "abc",
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
		"string_int": "1234",
		"string_int8": "23",
		"string_int16": "12345",
		"string_int32": "-127372843",
		"string_int64": "192438483489298",
		"string_uint": "1283829",
		"string_uint8": "233",
		"string_uint16": "62121",
		"string_uint32": "3083747392",
		"string_uint64": "2034857382993849",
		"string_float32": "9382.38429",
		"string_float64": "3984.293848292",
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
		"string_int": "1234",
		"string_int8": "23",
		"string_int16": "12345",
		"string_int32": "-127372843",
		"string_int64": "192438483489298",
		"string_uint": "1283829",
		"string_uint8": "233",
		"string_uint16": "62121",
		"string_uint32": "383083747392",
		"string_uint64": "2034857382993849",
		"string_float32": "9382.38429",
		"string_float64": "3984.293848292",
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
		"string_uint": "789",
		"string_float32": "10.234",
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
	var aStringUint Uint
	var aStringFloat32 Float32

	err = json.Unmarshal([]byte(strNormal), &result)

	if err != nil {
		t.Fatalf("cannot unmarshal json string. [e:%v]", err)
	}

	err = result.Decode(&normal)

	if err != nil {
		t.Fatalf("cannot decode normal struct. [e:%v]", err)
	}

	if normal.AnonymousInt1 != 123 {
		t.Fatalf("Fail to decode AnonymousInt1. [value:%v]", normal.AnonymousInt1)
	}

	if normal.AnonymousString2 != "abc" {
		t.Fatalf("Fail to decode AnonymousString2. [value:%v]", normal.AnonymousString2)
	}

	err = json.Unmarshal([]byte(strOverflow), &result)

	if err != nil {
		t.Fatalf("cannot unmarshal json string. [e:%v]", err)
	}

	err = result.Decode(&withError)

	if err == nil {
		t.Fatalf("struct should be overflow")
	}

	t.Logf("overflow struct. e:%v", err)

	err = json.Unmarshal([]byte(strMissAField), &result)

	if err != nil {
		t.Fatalf("cannot unmarshal json string. [e:%v]", err)
	}

	err = result.Decode(&withError)

	if err == nil {
		t.Fatalf("a field in struct should absent in json map.")
	}

	t.Logf("miss-a-field struct. e:%v", err)

	err = result.DecodeField("array_of_int.2", &anInt)

	if err != nil {
		t.Fatalf("cannot decode array item. [e:%v]", err)
	}

	if anInt != 56 {
		t.Fatalf("invalid array value. expected 56, actual %v", anInt)
	}

	err = result.DecodeField("nested_struct.int", &anInt)

	if err != nil {
		t.Fatalf("cannot decode nested struct item. [e:%v]", err)
	}

	if anInt != 123 {
		t.Fatalf("invalid array value. expected 123, actual %v", anInt)
	}

	err = result.DecodeField("string_uint", &aStringUint)

	if err != nil {
		t.Fatalf("cannot decode `string_uint`. [e:%v]", err)
	}

	if aStringUint != 789 {
		t.Fatalf("invalid uint value. expected 789, actual %v", aStringUint)
	}

	err = result.DecodeField("string_float32", &aStringFloat32)

	if err != nil {
		t.Fatalf("cannot decode `string_float32`. [e:%v]", err)
	}

	if aStringFloat32 != 10.234 {
		t.Fatalf("invalid uint value. expected 10.234, actual %v", aStringFloat32)
	}
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
		t.Fatalf("cannot unmarshal json string. [e:%v]", err)
	}

	err = result.Decode(&value)

	if err != nil {
		t.Fatalf("cannot decode struct. [e:%v]", err)
	}

	result = Result{}
	value = FieldTagStruct{}
	err = json.Unmarshal([]byte(strMissingField2Field), &result)

	if err != nil {
		t.Fatalf("cannot unmarshal json string. [e:%v]", err)
	}

	err = result.Decode(&value)

	if err != nil {
		t.Fatalf("cannot decode struct. [e:%v]", err)
	}

	if value.Field1 != "" {
		t.Fatalf("value field1 should be kept unchanged. [field1:%v]", value.Field1)
	}

	result = Result{}
	value = FieldTagStruct{}
	err = json.Unmarshal([]byte(strMissingRequiredField), &result)

	if err != nil {
		t.Fatalf("cannot unmarshal json string. [e:%v]", err)
	}

	err = result.Decode(&value)

	if err == nil {
		t.Fatalf("should fail to decode struct.")
	}

	t.Logf("expected decode error. [e:%v]", err)

	result = Result{}
	value = FieldTagStruct{}
	err = json.Unmarshal([]byte(strMissingBarField), &result)

	if err != nil {
		t.Fatalf("cannot unmarshal json string. [e:%v]", err)
	}

	err = result.Decode(&value)

	if err == nil {
		t.Fatalf("should fail to decode struct.")
	}

	t.Logf("expected decode error. [e:%v]", err)
}

type myTime struct {
	time.Time
}

func TestDecodeField(t *testing.T) {
	date1 := "2015-01-03T11:15:01Z"
	date2 := "2014-03-04T11:15:01.123Z"
	jsonStr := `{
        "int": 1234,
        "array": ["abcd", "efgh"],
        "map": {
            "key1": 5678,
            "nested_map": {
                "key2": "ijkl",
                "key3": [{
                    "key4": "mnop"
                }, {
                    "key5": 9012
                }]
            }
        },
        "message_tags": {
            "2": [
                {
                    "id": "4838901",
                    "name": "Foo Bar",
                    "type": "page"
                },
                {
                    "id": "293450302",
                    "name": "Player Rocks",
                    "type": "page"
                }
            ]
        },
        "nullStruct": {
        	"null": null
        },
        "timestamp": "` + date1 + `",
        "custom_timestamp": "` + date2 + `"
    }`

	var result Result
	var err error
	var anInt int
	var aString string
	var aSlice []string
	var subResults []Result
	var aNull NullStruct = NullStruct{
		Null: &anInt,
	}
	var aTimestamp time.Time
	var aCustomTimestamp myTime

	err = json.Unmarshal([]byte(jsonStr), &result)

	if err != nil {
		t.Fatalf("invalid json string. [e:%v]", err)
	}

	err = result.DecodeField("int", &anInt)

	if err != nil {
		t.Fatalf("cannot decode int field. [e:%v]", err)
	}

	if anInt != 1234 {
		t.Fatalf("expected int value is 1234. [int:%v]", anInt)
	}

	err = result.DecodeField("array.0", &aString)

	if err != nil {
		t.Fatalf("cannot decode array.0 field. [e:%v]", err)
	}

	if aString != "abcd" {
		t.Fatalf("expected array.0 value is 'abcd'. [string:%v]", aString)
	}

	err = result.DecodeField("array.1", &aString)

	if err != nil {
		t.Fatalf("cannot decode array.1 field. [e:%v]", err)
	}

	if aString != "efgh" {
		t.Fatalf("expected array.1 value is 'abcd'. [string:%v]", aString)
	}

	err = result.DecodeField("array.2", &aString)

	if err == nil {
		t.Fatalf("array.2 doesn't exist. expect an error.")
	}

	err = result.DecodeField("map.key1", &anInt)

	if err != nil {
		t.Fatalf("cannot decode map.key1 field. [e:%v]", err)
	}

	if anInt != 5678 {
		t.Fatalf("expected map.key1 value is 5678. [int:%v]", anInt)
	}

	err = result.DecodeField("map.nested_map.key2", &aString)

	if err != nil {
		t.Fatalf("cannot decode map.nested_map.key2 field. [e:%v]", err)
	}

	if aString != "ijkl" {
		t.Fatalf("expected map.nested_map.key2 value is 'ijkl'. [string:%v]", aString)
	}

	err = result.DecodeField("array", &aSlice)

	if err != nil {
		t.Fatalf("cannot decode array field. [e:%v]", err)
	}

	if len(aSlice) != 2 || aSlice[0] != "abcd" || aSlice[1] != "efgh" {
		t.Fatalf("expected array value is ['abcd', 'efgh']. [slice:%v]", aSlice)
	}

	err = result.DecodeField("map.nested_map.key3", &subResults)

	if err != nil {
		t.Fatalf("cannot decode map.nested_map.key3 field. [e:%v]", err)
	}

	if len(subResults) != 2 {
		t.Fatalf("expected sub results len is 2. [len:%v] [results:%v]", len(subResults), subResults)
	}

	err = subResults[0].DecodeField("key4", &aString)

	if err != nil {
		t.Fatalf("cannot decode key4 field in sub result. [e:%v]", err)
	}

	if aString != "mnop" {
		t.Fatalf("expected map.nested_map.key2 value is 'mnop'. [string:%v]", aString)
	}

	err = subResults[1].DecodeField("key5", &anInt)

	if err != nil {
		t.Fatalf("cannot decode key5 field. [e:%v]", err)
	}

	if anInt != 9012 {
		t.Fatalf("expected key5 value is 9012. [int:%v]", anInt)
	}

	err = result.DecodeField("message_tags.2.0.id", &aString)

	if err != nil {
		t.Fatalf("cannot decode message_tags.2.0.id field. [e:%v]", err)
	}

	if aString != "4838901" {
		t.Fatalf("expected message_tags.2.0.id value is '4838901'. [string:%v]", aString)
	}

	var messageTags MessageTags
	err = result.DecodeField("message_tags", &messageTags)

	if err != nil {
		t.Fatalf("cannot decode message_tags field. [e:%v]", err)
	}

	if len(messageTags) != 1 {
		t.Fatalf("expect messageTags have only 1 element. [len:%v]", len(messageTags))
	}

	aString = messageTags["2"][1].Id

	if aString != "293450302" {
		t.Fatalf("expect messageTags.2.1.id value is '293450302'. [value:%v]", aString)
	}

	err = result.DecodeField("nullStruct", &aNull)

	if err != nil {
		t.Fatalf("cannot decode nullStruct field. [e:%v]", err)
	}

	if aNull.Null != nil {
		t.Fatalf("expect aNull.Null is reset to nil.")
	}

	err = result.DecodeField("timestamp", &aTimestamp)

	if err != nil {
		t.Fatalf("cannot decode timestamp field. [e:%v]", err)
	}

	t1, _ := time.Parse(time.RFC3339, date1)

	if !aTimestamp.Equal(t1) {
		t.Fatalf("expect aTimestamp date to be %v [value:%v]", date1, aTimestamp.String())
	}

	err = result.DecodeField("custom_timestamp", &aCustomTimestamp)

	if err != nil {
		t.Fatalf("cannot decode custom_timestamp field. [e:%v]", err)
	}

	t2, _ := time.Parse(time.RFC3339, date2)

	if !aCustomTimestamp.Equal(t2) {
		t.Fatalf("expect aCustomTimestamp date to be %v [value:%v]", date2, aCustomTimestamp.String())
	}

	var timeStruct struct {
		Timestamp       time.Time `facebook:",required"`
		CustomTimestamp myTime    `facebook:",required"`
	}

	err = result.Decode(&timeStruct)

	if err != nil {
		t.Fatalf("cannot decode time struct. [e:%v]", err)
	}

	if !timeStruct.Timestamp.Equal(t1) {
		t.Fatalf("expect timeStruct.Timestamp date to be %v [value:%v]", date1, aTimestamp.String())
	}

	if !timeStruct.CustomTimestamp.Equal(t2) {
		t.Fatalf("expect timeStruct.CustomTimestamp date to be %v [value:%v]", date2, aCustomTimestamp.Time.String())
	}
}

type FacebookFriend struct {
	Id   string `facebook:",required"`
	Name string `facebook:",required"`
}

type FacebookFriends struct {
	Friends []FacebookFriend `facebook:"data,required"`
}

func TestPagingResultDecode(t *testing.T) {
	res := Result{
		"data": []interface{}{
			map[string]interface{}{
				"name": "friend 1",
				"id":   "1",
			},
			map[string]interface{}{
				"name": "friend 2",
				"id":   "2",
			},
		},
		"paging": map[string]interface{}{
			"next": "https://graph.facebook.com/...",
		},
	}
	paging, err := newPagingResult(nil, res)
	if err != nil {
		t.Fatalf("cannot create paging result. [e:%v]", err)
	}
	var friends FacebookFriends
	if err := paging.Decode(&friends); err != nil {
		t.Fatalf("cannot decode paging result. [e:%v]", err)
	}
	if len(friends.Friends) != 2 {
		t.Fatalf("expect to have 2 friends. [len:%v]", len(friends.Friends))
	}
	if friends.Friends[0].Name != "friend 1" {
		t.Fatalf("expect name to be 'friend 1'. [name:%v]", friends.Friends[0].Name)
	}
	if friends.Friends[0].Id != "1" {
		t.Fatalf("expect id to be '1'. [id:%v]", friends.Friends[0].Id)
	}
	if friends.Friends[1].Name != "friend 2" {
		t.Fatalf("expect name to be 'friend 2'. [name:%v]", friends.Friends[1].Name)
	}
	if friends.Friends[1].Id != "2" {
		t.Fatalf("expect id to be '2'. [id:%v]", friends.Friends[1].Id)
	}
}

func TestPagingResult(t *testing.T) {
	if FB_TEST_VALID_ACCESS_TOKEN == "" {
		t.Skipf("skip this case as we don't have a valid access token.")
	}

	session := &Session{}
	session.SetAccessToken(FB_TEST_VALID_ACCESS_TOKEN)
	res, err := session.Get("/me/home", Params{
		"limit": 2,
	})

	if err != nil {
		t.Fatalf("cannot get my home post. [e:%v]", err)
	}

	paging, err := res.Paging(session)

	if err != nil {
		t.Fatalf("cannot get paging information. [e:%v]", err)
	}

	data := paging.Data()

	if len(data) != 2 {
		t.Fatalf("expect to have only 2 post. [len:%v]", len(data))
	}

	t.Logf("result: %v", res)
	t.Logf("previous: %v", paging.previous)

	noMore, err := paging.Previous()

	if err != nil {
		t.Fatalf("cannot get paging information. [e:%v]", err)
	}

	if !noMore {
		t.Fatalf("should have no more post. %v", *paging.paging.Paging)
	}

	noMore, err = paging.Next()

	if err != nil {
		t.Fatalf("cannot get paging information. [e:%v]", err)
	}

	data = paging.Data()

	if len(data) != 2 {
		t.Fatalf("expect to have only 2 post. [len:%v]", len(data))
	}

	noMore, err = paging.Next()

	if err != nil {
		t.Fatalf("cannot get paging information. [e:%v]", err)
	}

	if len(paging.Data()) != 2 {
		t.Fatalf("expect to have only 2 post. [len:%v]", len(paging.Data()))
	}
}

func TestDecodeLargeInteger(t *testing.T) {
	bigIntegers := []int64{
		1<<53 - 2,
		1<<53 - 1,
		1 << 53,
		1<<53 + 1,
		1<<53 + 2,

		1<<54 - 2,
		1<<54 - 1,
		1 << 54,
		1<<54 + 1,
		1<<54 + 2,

		1<<60 - 2,
		1<<60 - 1,
		1 << 60,
		1<<60 + 1,
		1<<60 + 2,

		1<<63 - 2,
		1<<63 - 1,

		-(1<<53 - 2),
		-(1<<53 - 1),
		-(1 << 53),
		-(1<<53 + 1),
		-(1<<53 + 2),

		-(1<<54 - 2),
		-(1<<54 - 1),
		-(1 << 54),
		-(1<<54 + 1),
		-(1<<54 + 2),

		-(1<<60 - 2),
		-(1<<60 - 1),
		-(1 << 60),
		-(1<<60 + 1),
		-(1<<60 + 2),

		-(1<<53 - 2),
		-(1<<63 - 1),
		-(1 << 63),
	}
	jsonStr := `{
		"integers": [%v]
	}`

	buf := &bytes.Buffer{}

	for _, v := range bigIntegers {
		buf.WriteString(fmt.Sprintf("%v", v))
		buf.WriteRune(',')
	}

	buf.WriteRune('0')
	json := fmt.Sprintf(jsonStr, buf.String())

	res, err := MakeResult([]byte(json))

	if err != nil {
		t.Fatalf("cannot make result on test json string. [e:%v]", err)
	}

	var actualIntegers []int64
	err = res.DecodeField("integers", &actualIntegers)

	if err != nil {
		t.Fatalf("cannot decode integers from json. [e:%v]", err)
	}

	if len(actualIntegers) != len(bigIntegers)+1 {
		t.Fatalf("count of decoded integers is not correct. [expected:%v] [actual:%v]", len(bigIntegers)+1, len(actualIntegers))
	}

	for k, _ := range bigIntegers {
		if bigIntegers[k] != actualIntegers[k] {
			t.Logf("expected integers: %v", bigIntegers)
			t.Logf("actual integers:   %v", actualIntegers)
			t.Fatalf("a decoded integer is not expected. [expected:%v] [actual:%v]", bigIntegers[k], actualIntegers[k])
		}
	}
}

func TestMakeSliceResult(t *testing.T) {
	jsonStr := `{
		"error": {
			"message": "Invalid OAuth access token.",
			"type": "OAuthException",
			"code": 190
		}
	}`
	var res []Result
	err := makeResult([]byte(jsonStr), &res)

	if err == nil {
		t.Fatalf("makeResult must fail")
	}

	fbErr, ok := err.(*Error)

	if !ok {
		t.Fatalf("error must be a facebook error. [e:%v]", err)
	}

	if fbErr.Code != 190 {
		t.Fatalf("invalid facebook error. [e:%v]", fbErr.Error())
	}
}

func TestMakeSliceResultWithNilElements(t *testing.T) {
	jsonStr := `[
		null,
		{
			"foo": "bar"
		},
		null
	]`
	var res []Result
	err := makeResult([]byte(jsonStr), &res)

	if err != nil {
		t.Fatalf("fail to decode results. [e:%v]", err)
	}

	if len(res) != 3 {
		t.Fatalf("expect 3 elements in res. [res:%v]", res)
	}

	if res[0] != nil || res[1] == nil || res[2] != nil {
		t.Fatalf("decoded res is not expected. [res:%v]", res)
	}

	if res[1]["foo"].(string) != "bar" {
		t.Fatalf("decode res is not expected. [res:%v]", res)
	}
}

// case for #39.
func TestResultDecodeNumberString(t *testing.T) {
	res := Result{
		"int":    json.Number("1234"),
		"string": json.Number("1234"),
		"float":  json.Number("1234"),
	}

	var intValue int64
	var strValue string
	var floatValue float32
	var err error

	err = res.DecodeField("int", &intValue)

	if err != nil {
		t.Fatalf("fail to decode field `int`. [e:%v]", err)
	}

	if intValue != 1234 {
		t.Fatalf("unexpected int value. [expect:1234] [actual:%v]", intValue)
	}

	err = res.DecodeField("string", &strValue)

	if err != nil {
		t.Fatalf("fail to decode field `string`. [e:%v]", err)
	}

	if strValue != "1234" {
		t.Fatalf("unexpected string value. [expect:1234] [actual:%v]", strValue)
	}

	err = res.DecodeField("float", &floatValue)

	if err != nil {
		t.Fatalf("fail to decode field `float`. [e:%v]", err)
	}

	if floatValue != 1234 {
		t.Fatalf("unexpected float value. [expect:1234] [actual:%v]", floatValue)
	}
}

func TestResultDecodeUnmarshaler(t *testing.T) {
	correctJSON := `{
		"marshaler": {
			"name": "foo",
			"hash": 123456
		}
	}`

	errorJSON := `{
		"marshaler": {
			"name": "bar",
			"hash": 123456
		}
	}`

	var correctResult, errorResult Result
	err1 := makeResult([]byte(correctJSON), &correctResult)
	err2 := makeResult([]byte(errorJSON), &errorResult)

	if err1 != nil || err2 != nil {
		t.Fatalf("invalid test case input. [e1:%v] [e2:%v]", err1, err2)
	}

	var cms CustomMarshalerStruct

	err := correctResult.Decode(&cms)

	if err != nil {
		t.Fatalf("fail to decode input. [e:%v]", err)
	}

	err = errorResult.Decode(&cms)

	if err == nil {
		t.Fatalf("input should fail due to Unmarshaler.")
	}
}

type MixedTagStruct struct {
	Foo        int     `facebook:"bar" json:"player"`
	FirstTest  string  `facebook:"" json:"first"`
	SecondTest float64 `json:"second"`
	ThirdTest  uint32  `json:"-"`
	FourthTest int64   `facebook:",required" json:"fourth"`
}

func TestCompatibleWithJSONUnmarshal(t *testing.T) {
	res := Result{
		"foo":    1,
		"bar":    2,
		"player": 3,

		"first_test": "first",
		"first":      "test",

		"second_test": 1.2,
		"second":      3.4,

		"third_test": 5,
		"-":          6,

		"fourth_test": 7,
		"fourth":      8,
	}
	mts := &MixedTagStruct{}
	err := res.Decode(mts)

	if err != nil {
		t.Fatalf("fail to decode result. [e:%v]", err)
	}

	if expected := 2; mts.Foo != expected {
		t.Fatalf("mts.Foo is incorrect. [expected:%v] [actual:%v]", expected, mts.Foo)
	}

	if expected := "test"; mts.FirstTest != expected {
		t.Fatalf("mts.FirstTest is incorrect. [expected:%v] [actual:%v]", expected, mts.FirstTest)
	}

	if expected := 3.4; mts.SecondTest != expected {
		t.Fatalf("mts.SecondTest is incorrect. [expected:%v] [actual:%v]", expected, mts.SecondTest)
	}

	if expected := uint32(0); mts.ThirdTest != expected {
		t.Fatalf("mts.ThirdTest is incorrect. [expected:%v] [actual:%v]", expected, mts.ThirdTest)
	}

	if expected := int64(7); mts.FourthTest != expected {
		t.Fatalf("mts.FourthTest is incorrect. [expected:%v] [actual:%v]", expected, mts.FourthTest)
	}
}
