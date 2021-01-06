package customizablejson

import (
	"bytes"
	"encoding/json"
)

// JSONEncoder xxx
type JSONEncoder struct{}

// Register records a type and a function for encoding.
func (enc *JSONEncoder) Register(val interface{}, f func(v interface{}) ([]byte, error)) {
}

// Marshal xxx
func (enc *JSONEncoder) Marshal(v interface{}) ([]byte, error) {
	panic("TODO: implement me!")
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
