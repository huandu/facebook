// A facebook graph api client in go.
// https://github.com/huandu/facebook/
// 
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE
package facebook

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/url"
    "reflect"
    "runtime"
    "strconv"
    "strings"
    "unicode"
)

// Gets a field.
//
// Field can be a dot separated string.
// It means, if field name is "a.b.c", gets field value res["a"]["b"]["c"].
//
// To access array items, use index value in field.
// For instance, field "a.0.c" means to read res["a"][0]["c"].
//
// Returns nil if field doesn't exist.
func (res Result) Get(field string) interface{} {
    if field == "" {
        return res
    }

    f := strings.Split(field, ".")
    return res.get(f)
}

func (res Result) get(fields []string) interface{} {
    var arr []interface{}

    v, ok := res[fields[0]]

    if !ok || v == nil {
        return nil
    }

    if len(fields) == 1 {
        return v
    }

    for arr, ok = v.([]interface{}); ok; arr, ok = v.([]interface{}) {
        fields = fields[1:]
        n, err := strconv.ParseUint(fields[0], 10, 0)

        if err != nil {
            return nil
        }

        if n >= uint64(len(arr)) {
            return nil
        }

        v = arr[n]

        if len(fields) == 1 {
            return v
        }
    }

    res, ok = v.(Result)

    if !ok {
        return nil
    }

    return res.get(fields[1:])
}

// Decodes full result to a struct.
// It only decodes fields defined in the struct.
//
// As all facebook response fields are lower case strings,
// it will convert all camel-case field names to lower case string.
// e.g. field name "FooBar" will be converted to "foo_bar".
// The side effect is that if a struct has 2 fields with only capital
// differences, decoder will map these fields to a same result value.
//
// Returns error if v is not a struct or any v field name absents in res.
func (res Result) Decode(v interface{}) (err error) {
    defer func() {
        if r := recover(); r != nil {
            if _, ok := r.(runtime.Error); ok {
                panic(r)
            }

            err = r.(error)
        }
    }()

    err = res.decode(reflect.ValueOf(v), "")
    return
}

// Decodes a field of result to any type, including struct.
// Field name format is defined in Result.Get().
// 
// More details about decoding struct see Result.Decode().
func (res Result) DecodeField(field string, v interface{}) error {
    f := res.Get(field)

    if f == nil {
        return fmt.Errorf("field '%v' doesn't exist in result.", field)
    }

    return decodeField(f, reflect.ValueOf(v), field)
}

func (res Result) decode(v reflect.Value, fullName string) error {
    for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
        v = v.Elem()
    }

    if v.Kind() != reflect.Struct {
        return fmt.Errorf("output value must be a struct.")
    }

    if !v.CanSet() {
        return fmt.Errorf("output value cannot be set.")
    }

    if fullName != "" {
        fullName += "."
    }

    var field reflect.Value
    var name string
    var val interface{}
    var ok bool
    var err error
    var ret error

    num := v.Type().NumField()

    for i := 0; i < num; i++ {
        field = v.Field(i)
        name = camelCaseToUnderScore(v.Type().Field(i).Name)
        val, ok = res[name]

        if !ok {
            err = fmt.Errorf("cannot find field '%v%v' in result.", fullName, name)
        } else {
            err = decodeField(val, field, fmt.Sprintf("%v%v", fullName, name))
        }
        if err != nil {
            // Log each error
            log.Println(err)
            // Return the lastest error
            ret = err
        }
    }

    return ret
}

func decodeField(val interface{}, field reflect.Value, fullName string) error {
    if field.Kind() == reflect.Ptr {
        if field.IsNil() {
            field.Set(reflect.New(field.Type().Elem()))
        }

        field = field.Elem()
    }

    if !field.CanSet() {
        return fmt.Errorf("field '%v' cannot be decoded. make sure the output value is able to be set.", fullName)
    }

    kind := field.Kind()

    switch kind {
    case reflect.Bool:
        if b, ok := val.(bool); ok {
            field.SetBool(b)
        } else {
            return fmt.Errorf("field '%v' is not a bool in result.", fullName)
        }

    case reflect.Int8:
        if n, ok := val.(float64); ok {
            if n < -128 || n > 127 {
                return fmt.Errorf("field '%v' value exceeds the range of int8.", fullName)
            }

            field.SetInt(int64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Int16:
        if n, ok := val.(float64); ok {
            if n < -32768 || n > 32767 {
                return fmt.Errorf("field '%v' value exceeds the range of int16.", fullName)
            }

            field.SetInt(int64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Int32:
        if n, ok := val.(float64); ok {
            if n < -2147483648 || n > 2147483647 {
                return fmt.Errorf("field '%v' value exceeds the range of int32.", fullName)
            }

            field.SetInt(int64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Int64:
        if n, ok := val.(float64); ok {
            if n < -9223372036854775808 || n > 9223372036854775807 {
                return fmt.Errorf("field '%v' value exceeds the range of int64.", fullName)
            }

            field.SetInt(int64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Int:
        if n, ok := val.(float64); ok {
            if n < -9223372036854775808 || n > 9223372036854775807 {
                return fmt.Errorf("field '%v' value exceeds the range of int.", fullName)
            }

            field.SetInt(int64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Uint8:
        if n, ok := val.(float64); ok {
            if n < 0 || n > 0xFF {
                return fmt.Errorf("field '%v' value exceeds the range of uint8.", fullName)
            }

            field.SetUint(uint64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Uint16:
        if n, ok := val.(float64); ok {
            if n < 0 || n > 0xFFFF {
                return fmt.Errorf("field '%v' value exceeds the range of uint16.", fullName)
            }

            field.SetUint(uint64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Uint32:
        if n, ok := val.(float64); ok {
            if n < 0 || n > 0xFFFFFFFF {
                return fmt.Errorf("field '%v' value exceeds the range of uint32.", fullName)
            }

            field.SetUint(uint64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Uint64:
        if n, ok := val.(float64); ok {
            if n < 0 || n > 0xFFFFFFFFFFFFFFFF {
                return fmt.Errorf("field '%v' value exceeds the range of uint64.", fullName)
            }

            field.SetUint(uint64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Uint:
        if n, ok := val.(float64); ok {
            if n < 0 || n > 0xFFFFFFFFFFFFFFFF {
                return fmt.Errorf("field '%v' value exceeds the range of uint.", fullName)
            }

            field.SetUint(uint64(n))
        } else {
            return fmt.Errorf("field '%v' is not an integer in result.", fullName)
        }

    case reflect.Float32, reflect.Float64:
        if f, ok := val.(float64); ok {
            field.SetFloat(f)
        } else {
            return fmt.Errorf("field '%v' is not a float in result.", fullName)
        }

    case reflect.String:
        if s, ok := val.(string); ok {
            field.SetString(s)
        } else {
            return fmt.Errorf("field '%v' is not a string in result.", fullName)
        }

    case reflect.Struct:
        if r, ok := val.(map[string]interface{}); ok {
            if err := Result(r).decode(field, fullName); err != nil {
                return err
            }
        } else {
            return fmt.Errorf("field '%v' is not a json object in result.", fullName)
        }

    case reflect.Map:
        if m, ok := val.(map[string]interface{}); ok {
            // map key must be string
            if field.Type().Key().Kind() != reflect.String {
                return fmt.Errorf("field '%v' in struct is a map with non-string key type. it's not allowed.", fullName)
            }

            var needAddr bool
            valueType := field.Type().Elem()

            // shortcut for map of interface
            if valueType.Kind() == reflect.Interface {
                field.Set(reflect.ValueOf(m))
                break
            }

            if field.IsNil() {
                field.Set(reflect.MakeMap(field.Type()))
            }

            if valueType.Kind() == reflect.Ptr {
                valueType = valueType.Elem()
                needAddr = true
            }

            for key, value := range m {
                v := reflect.New(valueType)

                if err := decodeField(value, v, fmt.Sprintf("%v.%v", fullName, key)); err != nil {
                    return err
                }

                if needAddr {
                    field.SetMapIndex(reflect.ValueOf(key), v)
                } else {
                    field.SetMapIndex(reflect.ValueOf(key), v.Elem())
                }
            }
        } else {
            return fmt.Errorf("field '%v' is not a json object in result.", fullName)
        }

    case reflect.Slice, reflect.Array:
        if a, ok := val.([]interface{}); ok {
            if kind == reflect.Array {
                if field.Len() < len(a) {
                    return fmt.Errorf("cannot copy all field '%v' values to struct. expected len is %v. actual len is %v.",
                        fullName, field.Len(), len(a))
                }
            }

            var slc reflect.Value
            var needAddr bool
            valueType := field.Type().Elem()

            // shortcut for array of interface
            if valueType.Kind() == reflect.Interface {
                if kind == reflect.Array {
                    for i := 0; i < len(a); i++ {
                        field.Index(i).Set(reflect.ValueOf(a[i]))
                    }
                } else { // kind is slice
                    field.Set(reflect.ValueOf(a))
                }

                break
            }

            if kind == reflect.Array {
                slc = field.Slice(0, len(a))
            } else { // kind is slice
                slc = reflect.MakeSlice(field.Type(), len(a), len(a))
                field.Set(slc)
            }

            if valueType.Kind() == reflect.Ptr {
                needAddr = true
                valueType = valueType.Elem()
            }

            for i := 0; i < len(a); i++ {
                v := reflect.New(valueType)

                if err := decodeField(a[i], v, fmt.Sprintf("%v.%v", fullName, i)); err != nil {
                    return err
                }

                if needAddr {
                    slc.Index(i).Set(v)
                } else {
                    slc.Index(i).Set(v.Elem())
                }
            }
        } else {
            return fmt.Errorf("field '%v' is not a json array in result.", fullName)
        }

    default:
        return fmt.Errorf("field '%v' in struct uses unsupported type '%v'.", fullName, kind)
    }

    return nil
}

// Makes a new Params instance by given data.
// Data must be a struct or a map with string keys.
// MakeParams will change all struct field name to lower case name with underscore.
// e.g. "FooBar" will be changed to "foo_bar".
//
// Returns nil if data cannot be used to make a Params instance.
func MakeParams(data interface{}) (params Params) {
    if p, ok := data.(Params); ok {
        return p
    }

    defer func() {
        if r := recover(); r != nil {
            if _, ok := r.(runtime.Error); ok {
                panic(r)
            }

            params = nil
        }
    }()

    params = makeParams(reflect.ValueOf(data))
    return
}

func makeParams(value reflect.Value) (params Params) {
    for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
        value = value.Elem()
    }

    if value.Kind() != reflect.Struct {
        return
    }

    params = Params{}
    num := value.NumField()

    for i := 0; i < num; i++ {
        name := camelCaseToUnderScore(value.Type().Field(i).Name)
        field := value.Field(i)

        for field.Kind() == reflect.Ptr {
            field = field.Elem()
        }

        switch field.Kind() {
        case reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Invalid:
            // these types won't be marshalled in json.
            params = nil
            return

        case reflect.Struct:
            params[name] = makeParams(field)

        default:
            params[name] = field.Interface()
        }
    }

    return
}

// Encodes params to query string.
// If map value is not a string, Encode uses json.Marshal() to convert value to string.
//
// Encode will panic if Params contains values that cannot be marshalled to json string.
func (params Params) Encode() string {
    if params == nil || len(params) == 0 {
        return ""
    }

    buf := &bytes.Buffer{}

    for k, v := range params {
        buf.WriteString(url.QueryEscape(k))
        buf.WriteRune('=')

        if reflect.TypeOf(v).Kind() == reflect.String {
            buf.WriteString(url.QueryEscape(fmt.Sprint(v)))
        } else {
            jsonStr, err := json.Marshal(v)

            if err != nil {
                panic(err)
            }

            buf.WriteString(url.QueryEscape(string(jsonStr)))
        }

        buf.WriteRune('&')
    }

    return buf.String()[:buf.Len()-1]
}

func camelCaseToUnderScore(name string) string {
    if name == "" {
        return ""
    }

    buf := &bytes.Buffer{}
    notBeginning := false

    for _, r := range name {
        if unicode.IsUpper(r) {
            if notBeginning {
                buf.WriteRune('_')
            }

            buf.WriteRune(unicode.ToLower(r))
        } else {
            buf.WriteRune(r)
        }

        notBeginning = true
    }

    return buf.String()
}
