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
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const (
	ERROR_CODE_UNKNOWN = -1 // unknown facebook graph api error code.

	debugInfoKey   = "__debug__"
	debugProtoKey  = "__proto__"
	debugHeaderKey = "__header__"

	usageInfoKey = "__usage__"

	facebookApiVersionHeader = "facebook-api-version"
	facebookDebugHeader      = "x-fb-debug"
	facebookRevHeader        = "x-fb-rev"
)

var (
	typeOfJSONNumber = reflect.TypeOf(json.Number(""))
	typeOfInt        = reflect.TypeOf(Int(0))
	typeOfInt8       = reflect.TypeOf(Int8(0))
	typeOfInt16      = reflect.TypeOf(Int16(0))
	typeOfInt32      = reflect.TypeOf(Int32(0))
	typeOfInt64      = reflect.TypeOf(Int64(0))
	typeOfUint       = reflect.TypeOf(Uint(0))
	typeOfUint8      = reflect.TypeOf(Uint8(0))
	typeOfUint16     = reflect.TypeOf(Uint16(0))
	typeOfUint32     = reflect.TypeOf(Uint32(0))
	typeOfUint64     = reflect.TypeOf(Uint64(0))
	typeOfFloat32    = reflect.TypeOf(Float32(0))
	typeOfFloat64    = reflect.TypeOf(Float64(0))

	facebookSuccessJsonBytes = []byte("true")
)

// Result is Facebook API call result.
type Result map[string]interface{}

// PagingResult represents facebook API call result with paging information.
type PagingResult struct {
	session  *Session
	paging   pagingData
	previous string
	next     string
}

// BatchResult represents facebook batch API call result.
// See https://developers.facebook.com/docs/graph-api/making-multiple-requests/#multiple_methods.
type BatchResult struct {
	StatusCode int         // HTTP status code.
	Header     http.Header // HTTP response headers.
	Body       string      // Raw HTTP response body string.
	Result     Result      // Facebook api result parsed from body.
}

// DebugInfo is the debug information returned by facebook when debug mode is enabled.
type DebugInfo struct {
	Messages []DebugMessage // debug messages. it can be nil if there is no message.
	Header   http.Header    // all HTTP headers for this response.
	Proto    string         // HTTP protocol name for this response.

	// Facebook debug HTTP headers.
	FacebookApiVersion string // the actual graph API version provided by facebook-api-version HTTP header.
	FacebookDebug      string // the X-FB-Debug HTTP header.
	FacebookRev        string // the x-fb-rev HTTP header.
}

// UsageInfo is the app usage (rate limit) information returned by facebook when rate limits are possible.
type UsageInfo struct {
	App struct {
		CallCount    int `json:"call_count"`
		TotalTime    int `json:"total_time"`
		TotalCPUTime int `json:"total_cputime"`
	} `json:"app"`
	Page struct {
		CallCount    int `json:"call_count"`
		TotalTime    int `json:"total_time"`
		TotalCPUTime int `json:"total_cputime"`
	} `json:"page"`
}

// DebugMessage is one debug message in "__debug__" of graph API response.
type DebugMessage struct {
	Type    string
	Message string
	Link    string
}

// Special number types which can be decoded from either a number or a string.
// If developers intend to use a string in JSON as a number, these types can parse
// string to a number implicitly in `Result#Decode` or `Result#DecodeField`.
//
// Caveats: Parsing a string to a number may lose accuracy or shadow some errors.
type (
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Float32 float32
	Float64 float64
)

// MakeResult makes a Result from facebook Graph API response.
func MakeResult(jsonBytes []byte) (Result, error) {
	res := Result{}
	err := makeResult(jsonBytes, &res)

	if err != nil {
		return nil, err
	}

	// facebook may return an error
	return res, res.Err()
}

func makeResult(jsonBytes []byte, res interface{}) error {
	if bytes.Equal(jsonBytes, facebookSuccessJsonBytes) {
		return nil
	}

	jsonReader := bytes.NewReader(jsonBytes)
	dec := json.NewDecoder(jsonReader)

	// issue #19
	// app_scoped user_id in a post-Facebook graph 2.0 would exceeds 2^53.
	// use Number instead of float64 to avoid precision lost.
	dec.UseNumber()

	err := dec.Decode(res)

	if err != nil {
		typ := reflect.TypeOf(res)

		if typ != nil {
			// if res is a slice, jsonBytes may be a facebook error.
			// try to decode it as Error.
			kind := typ.Kind()

			if kind == reflect.Ptr {
				typ = typ.Elem()
				kind = typ.Kind()
			}

			if kind == reflect.Array || kind == reflect.Slice {
				var errRes Result
				err = makeResult(jsonBytes, &errRes)

				if err != nil {
					return err
				}

				err = errRes.Err()

				if err == nil {
					err = fmt.Errorf("cannot format facebook response; expect an array but get an object")
				}

				return err
			}
		}

		return fmt.Errorf("cannot format facebook response. %v", err)
	}

	return nil
}

// Get gets a field from Result.
//
// Field can be a dot separated string.
// If field name is "a.b.c", it will try to return value of res["a"]["b"]["c"].
//
// To access array items, use index value in field.
// For instance, field "a.0.c" means to read res["a"][0]["c"].
//
// It doesn't work with Result which has a key contains dot. Use GetField in this case.
//
// Returns nil if field doesn't exist.
func (res Result) Get(field string) interface{} {
	if field == "" {
		return res
	}

	f := strings.Split(field, ".")
	return res.get(f)
}

// GetField gets a field from Result.
//
// Arguments are treated as keys to access value in Result.
// If arguments are "a","b","c", it will try to return value of res["a"]["b"]["c"].
//
// To access array items, use index value as a string.
// For instance, args of "a", "0", "c" means to read res["a"][0]["c"].
//
// Returns nil if field doesn't exist.
func (res Result) GetField(fields ...string) interface{} {
	if len(fields) == 0 {
		return res
	}

	return res.get(fields)
}

func (res Result) get(fields []string) interface{} {
	v, ok := res[fields[0]]

	if !ok || v == nil {
		return nil
	}

	if len(fields) == 1 {
		return v
	}

	value := getValueField(reflect.ValueOf(v), fields[1:])

	if !value.IsValid() {
		return nil
	}

	return value.Interface()
}

func getValueField(value reflect.Value, fields []string) reflect.Value {
	valueType := value.Type()
	kind := valueType.Kind()
	field := fields[0]

	switch kind {
	case reflect.Array, reflect.Slice:
		// field must be a number.
		n, err := strconv.ParseUint(field, 10, 0)

		if err != nil {
			return reflect.Value{}
		}

		if n >= uint64(value.Len()) {
			return reflect.Value{}
		}

		// work around a reflect package pitfall.
		value = reflect.ValueOf(value.Index(int(n)).Interface())

	case reflect.Map:
		v := value.MapIndex(reflect.ValueOf(field))

		if !v.IsValid() {
			return v
		}

		// get real value type.
		value = reflect.ValueOf(v.Interface())

	default:
		return reflect.Value{}
	}

	if len(fields) == 1 {
		return value
	}

	return getValueField(value, fields[1:])
}

// Decode decodes full result to a struct.
// It only decodes fields defined in the struct.
//
// As all facebook response fields are lower case strings,
// Decode will convert all camel-case field names to lower case string.
// e.g. field name "FooBar" will be converted to "foo_bar".
// The side effect is that if a struct has 2 fields with only capital
// differences, decoder will map these fields to a same result value.
//
// If a field is missing in the result, Decode keeps it unchanged by default.
//
// The decoding of each struct field can be customized by the format string stored
// under the "facebook" key or the "json" key in the struct field's tag.
// The "facebook" key is recommended as it's specifically designed for this package.
//
// Examples:
//
//     type Foo struct {
//         // "id" must exist in response. note the leading comma.
//         Id string `facebook:",required"`
//
//         // use "name" as field name in response.
//         TheName string `facebook:"name"`
//
//         // the "json" key also works as expected.
//         Key string `json:"my_key"`
//
//         // if both "facebook" and "json" key are set, the "facebook" key is used.
//         Value string `facebook:"value" json:"shadowed"`
//     }
//
// To change default behavior, set a struct tag `facebook:",required"` to fields
// should not be missing.
//
// Returns error if v is not a struct or any required v field name absents in res.
func (res Result) Decode(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}

			if errStr, ok := r.(string); ok {
				err = errors.New(errStr)
				return
			}

			if errErr, ok := r.(error); ok {
				err = errErr
				return
			}

			panic(r)
		}
	}()

	err = res.decode(reflect.ValueOf(v), "")
	return
}

// DecodeField decodes a field of result to any type, including struct.
// Field name format is defined in Result.Get().
//
// More details about decoding struct see Result.Decode().
func (res Result) DecodeField(field string, v interface{}) error {
	f := res.Get(field)

	if f == nil {
		return fmt.Errorf("field '%v' doesn't exist in result", field)
	}

	return decodeField(reflect.ValueOf(f), reflect.ValueOf(v), field)
}

// Err returns an error if Result is a Graph API error.
//
// The returned error can be converted to Error by type assertion.
//     err := res.Err()
//     if err != nil {
//         if e, ok := err.(*Error); ok {
//             // read more details in e.Message, e.Code and e.Type
//         }
//     }
//
// For more information about Graph API Errors, see
// https://developers.facebook.com/docs/reference/api/errors/
func (res Result) Err() error {
	var err Error
	e := res.DecodeField("error", &err)

	// no "error" in result. result is not an error.
	if e != nil {
		return nil
	}

	// code may be missing in error.
	// assign a non-zero value to it.
	if err.Code == 0 {
		err.Code = ERROR_CODE_UNKNOWN
	}

	return &err
}

// Paging creates a PagingResult for this Result and
// returns error if the Result cannot be used for paging.
//
// Facebook uses following JSON structure to response paging information.
// If "data" doesn't present in Result, Paging will return error.
//     {
//         "data": [...],
//         "paging": {
//             "previous": "https://graph.facebook.com/...",
//             "next": "https://graph.facebook.com/..."
//         }
//     }
func (res Result) Paging(session *Session) (*PagingResult, error) {
	return newPagingResult(session, res)
}

// Batch creates a BatchResult for this result and
// returns error if the Result is not a batch api response.
//
// See BatchApi document for a sample usage.
func (res Result) Batch() (*BatchResult, error) {
	return newBatchResult(res)
}

// DebugInfo creates a DebugInfo for this result if this result
// has "__debug__" key.
func (res Result) DebugInfo() *DebugInfo {
	var info Result
	err := res.DecodeField(debugInfoKey, &info)

	if err != nil {
		return nil
	}

	debugInfo := &DebugInfo{}
	info.DecodeField("messages", &debugInfo.Messages)

	if proto, ok := info[debugProtoKey]; ok {
		if v, ok := proto.(string); ok {
			debugInfo.Proto = v
		}
	}

	if header, ok := info[debugHeaderKey]; ok {
		if v, ok := header.(http.Header); ok {
			debugInfo.Header = v

			debugInfo.FacebookApiVersion = v.Get(facebookApiVersionHeader)
			debugInfo.FacebookDebug = v.Get(facebookDebugHeader)
			debugInfo.FacebookRev = v.Get(facebookRevHeader)
		}
	}

	return debugInfo
}

// UsageInfo returns app and page usage info (rate limits)
func (res Result) UsageInfo() *UsageInfo {
	if usageInfo, ok := res[usageInfoKey]; ok {
		if usage, ok := usageInfo.(*UsageInfo); ok {
			return usage
		}
	}

	return nil
}

func (res Result) decode(v reflect.Value, fullName string) error {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("output value must be a struct")
	}

	if !v.CanSet() {
		return fmt.Errorf("output value cannot be set")
	}

	var field reflect.Value
	var fieldInfo reflect.StructField
	var name, dot string
	var val interface{}
	var ok, required bool
	var err error

	if fullName != "" {
		dot = "."
	}

	vType := v.Type()
	num := vType.NumField()

	for i := 0; i < num; i++ {
		name = ""
		required = false
		field = v.Field(i)
		fieldInfo = vType.Field(i)

		// parse struct field tag.
		if fbTag := fieldInfo.Tag.Get("facebook"); fbTag != "" {
			if fbTag == "-" {
				continue
			}

			index := strings.IndexRune(fbTag, ',')

			if index == -1 {
				name = fbTag
			} else {
				name = fbTag[:index]

				if fbTag[index:] == ",required" {
					required = true
				}
			}
		} else {
			// compatible with json tag.
			fbTag = fieldInfo.Tag.Get("json")

			if fbTag == "-" {
				continue
			}

			index := strings.IndexRune(fbTag, ',')

			if index == -1 {
				name = fbTag
			} else {
				name = fbTag[:index]
			}
		}

		// embedded field is "expanded" when decoding.
		// special case: treat it as a normal field if the name is not empty.
		if fieldInfo.Anonymous && name == "" {
			if err = decodeField(reflect.ValueOf(res), field, fullName); err != nil {
				return err
			}

			continue
		}

		if name == "" {
			name = camelCaseToUnderScore(fieldInfo.Name)
		}

		val, ok = res[name]

		if !ok {
			// check whether the field is required. if so, report error.
			if required {
				return fmt.Errorf("cannot find field '%v%v%v' in result", fullName, dot, name)
			}

			continue
		}

		if err = decodeField(reflect.ValueOf(val), field, fmt.Sprintf("%v%v%v", fullName, dot, name)); err != nil {
			return err
		}
	}

	return nil
}

func decodeField(val reflect.Value, field reflect.Value, fullName string) error {
	if field.Kind() == reflect.Ptr {
		// reset Ptr field if val is nil.
		if !val.IsValid() {
			if !field.IsNil() && field.CanSet() {
				field.Set(reflect.Zero(field.Type()))
			}

			return nil
		}

		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		field = field.Elem()
	}

	if !field.CanSet() {
		return fmt.Errorf("field '%v' cannot be decoded; make sure the output value is able to be set", fullName)
	}

	if !val.IsValid() {
		return fmt.Errorf("field '%v' is not a pointer; fail to assign nil to it", fullName)
	}

	// if field implements Unmarshaler, let field unmarshals data itself.
	if unmarshaler := indirect(field); unmarshaler != nil {
		data, err := json.Marshal(val.Interface())

		if err != nil {
			return fmt.Errorf("fail to marshal value for field '%v' with error %v", fullName, err)
		}

		return unmarshaler.UnmarshalJSON(data)
	}

	kind := field.Kind()
	fieldType := field.Type()
	valType := val.Type()

	switch kind {
	case reflect.Bool:
		if valType.Kind() == reflect.Bool {
			field.SetBool(val.Bool())
		} else {
			return fmt.Errorf("field '%v' is not a bool in result", fullName)
		}

	case reflect.Int8:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()

			if n < math.MinInt8 || n > math.MaxInt64 {
				return fmt.Errorf("field '%v' value exceeds the range of int8", fullName)
			}

			field.SetInt(int64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()

			if n > math.MaxInt8 {
				return fmt.Errorf("field '%v' value exceeds the range of int8", fullName)
			}

			field.SetInt(int64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < math.MinInt8 || n > math.MaxInt8 {
				return fmt.Errorf("field '%v' value exceeds the range of int8", fullName)
			}

			field.SetInt(int64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Int8.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfInt8 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseInt(val.String(), 10, 8)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid int8", fullName)
			}

			field.SetInt(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Int16:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()

			if n < math.MinInt16 || n > math.MaxInt16 {
				return fmt.Errorf("field '%v' value exceeds the range of int16", fullName)
			}

			field.SetInt(int64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()

			if n > math.MaxInt16 {
				return fmt.Errorf("field '%v' value exceeds the range of int16", fullName)
			}

			field.SetInt(int64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < math.MinInt16 || n > math.MaxInt16 {
				return fmt.Errorf("field '%v' value exceeds the range of int16", fullName)
			}

			field.SetInt(int64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Int16.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfInt16 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseInt(val.String(), 10, 16)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid int16", fullName)
			}

			field.SetInt(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Int32:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()

			if n < math.MinInt32 || n > math.MaxInt32 {
				return fmt.Errorf("field '%v' value exceeds the range of int32", fullName)
			}

			field.SetInt(int64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()

			if n > math.MaxInt32 {
				return fmt.Errorf("field '%v' value exceeds the range of int32", fullName)
			}

			field.SetInt(int64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < math.MinInt32 || n > math.MaxInt32 {
				return fmt.Errorf("field '%v' value exceeds the range of int32", fullName)
			}

			field.SetInt(int64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Int32.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfInt32 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseInt(val.String(), 10, 32)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid int32", fullName)
			}

			field.SetInt(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Int64:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()
			field.SetInt(n)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()

			if n > math.MaxInt64 {
				return fmt.Errorf("field '%v' value exceeds the range of int64", fullName)
			}

			field.SetInt(int64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < math.MinInt64 || n > math.MaxInt64 {
				return fmt.Errorf("field '%v' value exceeds the range of int64", fullName)
			}

			field.SetInt(int64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Int64.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfInt64 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseInt(val.String(), 10, 64)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid int64", fullName)
			}

			field.SetInt(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Int:
		bits := field.Type().Bits()

		var min, max int64

		if bits == 32 {
			min = math.MinInt32
			max = math.MaxInt32
		} else if bits == 64 {
			min = math.MinInt64
			max = math.MaxInt64
		}

		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()

			if n < min || n > max {
				return fmt.Errorf("field '%v' value exceeds the range of int", fullName)
			}

			field.SetInt(int64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()

			if n > uint64(max) {
				return fmt.Errorf("field '%v' value exceeds the range of int", fullName)
			}

			field.SetInt(int64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < float64(min) || n > float64(max) {
				return fmt.Errorf("field '%v' value exceeds the range of int", fullName)
			}

			field.SetInt(int64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Int.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfInt {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseInt(val.String(), 10, bits)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid int%v", fullName, bits)
			}

			field.SetInt(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Uint8:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()

			if n < 0 || n > math.MaxUint8 {
				return fmt.Errorf("field '%v' value exceeds the range of uint8", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()

			if n > math.MaxUint8 {
				return fmt.Errorf("field '%v' value exceeds the range of uint8", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < 0 || n > math.MaxUint8 {
				return fmt.Errorf("field '%v' value exceeds the range of uint8", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Uint8.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfUint8 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseUint(val.String(), 10, 8)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid uint8", fullName)
			}

			field.SetUint(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Uint16:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()

			if n < 0 || n > math.MaxUint16 {
				return fmt.Errorf("field '%v' value exceeds the range of uint16", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()

			if n > math.MaxUint16 {
				return fmt.Errorf("field '%v' value exceeds the range of uint16", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < 0 || n > math.MaxUint16 {
				return fmt.Errorf("field '%v' value exceeds the range of uint16", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Uint16.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfUint16 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseUint(val.String(), 10, 16)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid uint16", fullName)
			}

			field.SetUint(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Uint32:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()

			if n < 0 || n > math.MaxUint32 {
				return fmt.Errorf("field '%v' value exceeds the range of uint32", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()

			if n > math.MaxUint32 {
				return fmt.Errorf("field '%v' value exceeds the range of uint32", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < 0 || n > math.MaxUint32 {
				return fmt.Errorf("field '%v' value exceeds the range of uint32", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Uint32.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfUint32 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseUint(val.String(), 10, 32)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid uint32", fullName)
			}

			field.SetUint(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Uint64:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()

			if n < 0 {
				return fmt.Errorf("field '%v' value exceeds the range of uint64", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()
			field.SetUint(n)

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < 0 || n > math.MaxUint64 {
				return fmt.Errorf("field '%v' value exceeds the range of uint64", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Uint64.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfUint64 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseUint(val.String(), 10, 64)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid uint64", fullName)
			}

			field.SetUint(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Uint:
		bits := field.Type().Bits()

		var max uint64

		if bits == 32 {
			max = math.MaxUint32
		} else if bits == 64 {
			max = math.MaxUint64
		}

		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()

			if n < 0 || uint64(n) > max {
				return fmt.Errorf("field '%v' value exceeds the range of uint", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()

			if n > max {
				return fmt.Errorf("field '%v' value exceeds the range of uint", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if n < 0 || n > float64(max) {
				return fmt.Errorf("field '%v' value exceeds the range of uint", fullName)
			}

			field.SetUint(uint64(n))

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Uint.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfUint {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseUint(val.String(), 10, bits)

			if err != nil {
				return fmt.Errorf("field '%v' value is not a valid uint%v", fullName, bits)
			}

			field.SetUint(n)

		default:
			return fmt.Errorf("field '%v' is not an integer in result", fullName)
		}

	case reflect.Float32:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()
			field.SetFloat(float64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()
			field.SetFloat(float64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()

			if math.Abs(n) > math.MaxFloat32 {
				return fmt.Errorf("field '%v' value exceeds the range of float32", fullName)
			}

			field.SetFloat(n)

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Float32.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfFloat32 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseFloat(val.String(), 32)

			if err != nil {
				return fmt.Errorf("field '%v' is not a valid float32", fullName)
			}

			field.SetFloat(n)

		default:
			return fmt.Errorf("field '%v' is not a float in result", fullName)
		}

	case reflect.Float64:
		switch valType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := val.Int()
			field.SetFloat(float64(n))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n := val.Uint()
			field.SetFloat(float64(n))

		case reflect.Float32, reflect.Float64:
			n := val.Float()
			field.SetFloat(n)

		case reflect.String:
			// val is allowed to be used as number only if val is json.Number or field is fb.Float64.
			if val.Type() != typeOfJSONNumber && fieldType != typeOfFloat64 {
				return fmt.Errorf("field '%v' value is string, not a number", fullName)
			}

			n, err := strconv.ParseFloat(val.String(), 64)

			if err != nil {
				return fmt.Errorf("field '%v' is not a valid float64", fullName)
			}

			field.SetFloat(n)

		default:
			return fmt.Errorf("field '%v' is not a float in result", fullName)
		}

	case reflect.String:
		if valType.Kind() != reflect.String {
			return fmt.Errorf("field '%v' is not a string in result", fullName)
		}

		field.SetString(val.String())

	case reflect.Struct:
		if valType.Kind() != reflect.Map || valType.Key().Kind() != reflect.String {
			return fmt.Errorf("field '%v' is not a json object in result", fullName)
		}

		// safe convert val to Result. type assertion doesn't work in this case.
		var r Result
		reflect.ValueOf(&r).Elem().Set(val)

		if err := r.decode(field, fullName); err != nil {
			return err
		}

	case reflect.Map:
		if valType.Kind() != reflect.Map || valType.Key().Kind() != reflect.String {
			return fmt.Errorf("field '%v' is not a json object in result", fullName)
		}

		// map key must be string
		if field.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("field '%v' in struct must be a map whose key type is string", fullName)
		}

		var needAddr bool
		valueType := field.Type().Elem()

		// shortcut for map[string]interface{}.
		if valueType.Kind() == reflect.Interface {
			field.Set(val)
			break
		}

		if field.IsNil() {
			field.Set(reflect.MakeMap(field.Type()))
		}

		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
			needAddr = true
		}

		for _, key := range val.MapKeys() {
			// val.MapIndex(key) returns a Value with wrong type.
			// use following trick to get correct Value.
			value := reflect.ValueOf(val.MapIndex(key).Interface())
			newValue := reflect.New(valueType)

			if err := decodeField(value, newValue, fmt.Sprintf("%v.%v", fullName, key)); err != nil {
				return err
			}

			if needAddr {
				field.SetMapIndex(key, newValue)
			} else {
				field.SetMapIndex(key, newValue.Elem())
			}
		}

	case reflect.Slice, reflect.Array:
		if valType.Kind() != reflect.Slice && valType.Kind() != reflect.Array {
			return fmt.Errorf("field '%v' is not a json array in result", fullName)
		}

		valLen := val.Len()

		if kind == reflect.Array {
			if field.Len() < valLen {
				return fmt.Errorf("cannot copy all field '%v' values to struct; expected len is %v but actual is %v",
					fullName, field.Len(), valLen)
			}
		}

		var slc reflect.Value
		var needAddr bool

		valueType := field.Type().Elem()

		// shortcut for array of interface
		if valueType.Kind() == reflect.Interface {
			if kind == reflect.Array {
				for i := 0; i < valLen; i++ {
					field.Index(i).Set(val.Index(i))
				}
			} else { // kind is slice
				field.Set(val)
			}

			break
		}

		if kind == reflect.Array {
			slc = field.Slice(0, valLen)
		} else {
			// kind is slice
			slc = reflect.MakeSlice(field.Type(), valLen, valLen)
			field.Set(slc)
		}

		if valueType.Kind() == reflect.Ptr {
			needAddr = true
			valueType = valueType.Elem()
		}

		for i := 0; i < valLen; i++ {
			// val.Index(i) returns a Value with wrong type.
			// use following trick to get correct Value.
			valIndexValue := reflect.ValueOf(val.Index(i).Interface())
			newValue := reflect.New(valueType)

			if err := decodeField(valIndexValue, newValue, fmt.Sprintf("%v.%v", fullName, i)); err != nil {
				return err
			}

			if needAddr {
				slc.Index(i).Set(newValue)
			} else {
				slc.Index(i).Set(newValue.Elem())
			}
		}

	default:
		return fmt.Errorf("field '%v' in struct uses unsupported type '%v'", fullName, kind)
	}

	return nil
}

// Indirect walks down v allocating pointers as needed until it gets to a non-pointer.
// If v implements json.Unmarshaler, indrect stops and returns it.
//
// This implementation is a modified version of http://golang.org/src/encoding/json/decode.go.
func indirect(v reflect.Value) json.Unmarshaler {
	// if v is a struct field and v's pointer may implement json.Unmarshaler,
	// try to discover this case.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}

	for {
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()

			if e.Kind() == reflect.Ptr && !e.IsNil() && e.Elem().Kind() == reflect.Ptr {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.Elem().Kind() != reflect.Ptr && v.CanSet() {
			break
		}

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(json.Unmarshaler); ok {
				return u
			}
		}

		v = v.Elem()
	}

	if v.Type().NumMethod() > 0 {
		if u, ok := v.Interface().(json.Unmarshaler); ok {
			return u
		}
	}

	return nil
}
