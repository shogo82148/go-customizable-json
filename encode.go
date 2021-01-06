package customizablejson

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sync"
)

type encoderFunc func(v interface{}) (interface{}, error)

// JSONEncoder xxx
type JSONEncoder struct {
	cache sync.Map
}

// Register records a type and a function for encoding.
func (enc *JSONEncoder) Register(val interface{}, f func(v interface{}) ([]byte, error)) {
	typ := reflect.TypeOf(val)
	enc.cache.Store(typ, encoderFunc(func(v interface{}) (interface{}, error) {
		data, err := f(v)
		if err != nil {
			return nil, err
		}
		var ret interface{}
		if err := json.Unmarshal(data, &ret); err != nil {
			return nil, err
		}
		return ret, nil
	}))
}

// Marshal xxx
func (enc *JSONEncoder) Marshal(v interface{}) ([]byte, error) {
	typ := reflect.TypeOf(v)
	f, ok := enc.cache.Load(typ)
	if !ok {
		panic("TODO: implement me!")
	}
	ret, err := f.(encoderFunc)(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ret)
}

// MarshalIndent is like Marshal but applies Indent to format the output.
func (enc *JSONEncoder) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	panic("TODO: implement me!")
}

var defaultEncoder = new(JSONEncoder)

// Marshal xxx
func Marshal(v interface{}) ([]byte, error) {
	return defaultEncoder.Marshal(v)
}

// MarshalIndent is like Marshal but applies Indent to format the output.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return defaultEncoder.MarshalIndent(v, prefix, indent)
}

// HTMLEscape is an alias of json.HTMLEscape
func HTMLEscape(dst *bytes.Buffer, src []byte) {
	json.HTMLEscape(dst, src)
}

// Marshaler is an alias of json.Marshaler
type Marshaler = json.Marshaler

// UnsupportedTypeError is an alias of json.UnsupportedTypeError
type UnsupportedTypeError = json.UnsupportedTypeError

// UnsupportedValueError is an alias of UnsupportedValueError
type UnsupportedValueError = json.UnsupportedValueError

// MarshalerError is an alias of json.MarshalerError
type MarshalerError = json.MarshalerError
