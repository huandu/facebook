// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
)

const (
	mimeFormURLEncoded = "application/x-www-form-urlencoded"
	mimeFormData       = "multipart/form-data"
)

var (
	typeOfPointerToBinaryData = reflect.TypeOf(&BinaryData{})
	typeOfPointerToBinaryFile = reflect.TypeOf(&BinaryFile{})
)

// Params is the params used to send Facebook API request.
//
// For general uses, just use Params as an ordinary map.
//
// For advanced uses, use MakeParams to create Params from any struct.
type Params map[string]interface{}

// MakeParams makes a new Params instance by given data.
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

	// only map with string keys can be converted to Params
	if value.Kind() == reflect.Map && value.Type().Key().Kind() == reflect.String {
		params = Params{}

		for _, key := range value.MapKeys() {
			params[key.String()] = value.MapIndex(key).Interface()
		}

		return
	}

	if value.Kind() != reflect.Struct {
		return
	}

	params = Params{}
	t := value.Type()
	num := value.NumField()

	for i := 0; i < num; i++ {
		sf := t.Field(i)
		tag := sf.Tag
		name := ""
		omitEmpty := false

		// If field tag "facebook" or "json" exists, use it as field name and options.
		fbTag := tag.Get("facebook")
		jsonTag := tag.Get("json")

		if fbTag != "" || jsonTag != "" {
			optTag := jsonTag

			// If field tag "facebook" exists, it's preferred.
			if fbTag != "" {
				optTag = fbTag
			}

			opts := strings.Split(optTag, ",")

			if opts[0] != "" {
				name = opts[0]
			}

			for _, opt := range opts[1:] {
				if opt == "omitempty" {
					omitEmpty = true
				}
			}
		}

		field := value.Field(i)

		if omitEmpty && isEmptyValue(field) {
			continue
		}

		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		// If name is not set in field tag, use field name directly.
		if name == "" {
			name = camelCaseToUnderScore(sf.Name)
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

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return false
}

// Encode encodes params to query string.
// If map value is not a string, Encode uses json.Marshal() to convert value to string.
//
// Encode may panic if Params contains values that cannot be marshalled to json string.
func (params Params) Encode(writer io.Writer) (mime string, err error) {
	if len(params) == 0 {
		mime = mimeFormURLEncoded
		return
	}

	// check whether params contains any binary data.
	hasBinary := false

	for _, v := range params {
		typ := reflect.TypeOf(v)

		if typ == typeOfPointerToBinaryData || typ == typeOfPointerToBinaryFile {
			hasBinary = true
			break
		}
	}

	if hasBinary {
		return params.encodeMultipartForm(writer)
	}

	return params.encodeFormURLEncoded(writer)
}

func (params Params) encodeFormURLEncoded(writer io.Writer) (mime string, err error) {
	var jsonStr []byte
	written := false

	for k, v := range params {
		if v == nil {
			continue
		}

		if written {
			io.WriteString(writer, "&")
		}

		io.WriteString(writer, url.QueryEscape(k))
		io.WriteString(writer, "=")

		if reflect.TypeOf(v).Kind() == reflect.String {
			io.WriteString(writer, url.QueryEscape(reflect.ValueOf(v).String()))
		} else {
			jsonStr, err = json.Marshal(v)

			if err != nil {
				return
			}

			io.WriteString(writer, url.QueryEscape(string(jsonStr)))
		}

		written = true
	}

	mime = mimeFormURLEncoded
	return
}

func (params Params) encodeMultipartForm(writer io.Writer) (mime string, err error) {
	w := multipart.NewWriter(writer)
	defer func() {
		w.Close()
		mime = w.FormDataContentType()
	}()

	for k, v := range params {
		switch value := v.(type) {
		case *BinaryData:
			var dst io.Writer
			filePart := createFormFile(k, value.Filename, value.ContentType)
			dst, err = w.CreatePart(filePart)

			if err != nil {
				return
			}

			_, err = io.Copy(dst, value.Source)

			if err != nil {
				return
			}

		case *BinaryFile:
			var dst io.Writer
			var file *os.File
			var path string

			filePart := createFormFile(k, value.Filename, value.ContentType)
			dst, err = w.CreatePart(filePart)

			if err != nil {
				return
			}

			if value.Path == "" {
				path = value.Filename
			} else {
				path = value.Path
			}

			file, err = os.Open(path)

			if err != nil {
				return
			}

			_, err = io.Copy(dst, file)

			if err != nil {
				return
			}

		default:
			var dst io.Writer
			var jsonStr []byte

			dst, err = w.CreateFormField(k)

			if reflect.TypeOf(v).Kind() == reflect.String {
				io.WriteString(dst, reflect.ValueOf(v).String())
			} else {
				jsonStr, err = json.Marshal(v)

				if err != nil {
					return
				}

				_, err = dst.Write(jsonStr)

				if err != nil {
					return
				}
			}
		}
	}

	return
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func createFormFile(fieldName, fileName, contentType string) textproto.MIMEHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			quoteEscaper.Replace(fieldName), quoteEscaper.Replace(fileName)))

	if contentType == "" {
		contentType = mime.TypeByExtension(path.Ext(fileName))

		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	h.Set("Content-Type", contentType)
	return h
}
