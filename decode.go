package customizablejson

import (
	"encoding/json"
)

// JSONDecoder xxx
type JSONDecoder struct{}

// Register records a type and a function for encoding.
func (dec *JSONDecoder) Register(val interface{}, f func(v interface{}, data []byte) error) {
}

// Unmarshal xxx
func (dec *JSONDecoder) Unmarshal(data []byte, v interface{}) error {
	panic("TODO: implement me!")
}

var defaultDecoder = new(JSONDecoder)

// Unmarshal xxx
func Unmarshal(data []byte, v interface{}) error {
	return defaultDecoder.Unmarshal(data, v)
}

// Unmarshaler is an alias of json.Unmarshaler.
type Unmarshaler = json.Unmarshaler

// UnmarshalTypeError is an alias of json.UnmarshalTypeError.
type UnmarshalTypeError = json.UnmarshalTypeError

// UnmarshalFieldError is an alias of json.UnmarshalFieldError.
type UnmarshalFieldError = json.UnmarshalFieldError

// InvalidUnmarshalError is an alias of json.InvalidUnmarshalError.
type InvalidUnmarshalError = json.InvalidUnmarshalError

// Number is an alias of json.Number.
type Number = json.Number
