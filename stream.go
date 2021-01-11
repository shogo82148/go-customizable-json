package customizablejson

import (
	"encoding/json"
	"io"
	"reflect"
)

// A Decoder reads and decodes JSON values from an input stream.
type Decoder struct {
	dec                   *json.Decoder
	myDec                 *JSONDecoder
	disallowUnknownFields bool
	useNumber             bool
	errorContext          struct { // provides context for type errors
		Struct string
		Field  string
	}
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{}
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func (dec *JSONDecoder) NewDecoder(r io.Reader) *Decoder {
	d := json.NewDecoder(r)
	d.UseNumber()
	return &Decoder{
		dec:   d,
		myDec: dec,
	}
}

// Buffered returns a reader of the data remaining in the Decoder's buffer.
// The reader is valid until the next call to Decode.
func (dec *Decoder) Buffered() io.Reader {
	return dec.dec.Buffered()
}

func (dec *Decoder) withErrorContext(err error) error {
	if dec.errorContext.Struct != "" || dec.errorContext.Field != "" {
		switch err := err.(type) {
		case *UnmarshalTypeError:
			err.Struct = dec.errorContext.Struct
			err.Field = dec.errorContext.Field
			return err
		}
	}
	return err
}

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a
// Number instead of as a float64.
func (dec *Decoder) UseNumber() {
	dec.useNumber = true
}

// DisallowUnknownFields causes the Decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination.
func (dec *Decoder) DisallowUnknownFields() {
	dec.disallowUnknownFields = true
}

// Decode xxx
func (dec *Decoder) Decode(v interface{}) error {
	var iv interface{}
	if err := dec.dec.Decode(&iv); err != nil {
		return err
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}
	return dec.decode(iv, rv)
}

// An Encoder writes JSON values to an output stream.
type Encoder struct {
	encoder   *json.Encoder
	myEncoder *JSONEncoder
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return defaultEncoder.NewEncoder(w)
}

// NewEncoder returns a new encoder that writes to w.
func (enc *JSONEncoder) NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		encoder:   json.NewEncoder(w),
		myEncoder: enc,
	}
}

// Encode xxx
func (enc *Encoder) Encode(v interface{}) error {
	state := enc.myEncoder.newState()
	ret, err := state.toInterface(v)
	if err != nil {
		return err
	}
	return enc.encoder.Encode(ret)
}

// SetIndent instructs the encoder to format each subsequent encoded
// value as if indented by the package-level function Indent(dst, src, prefix, indent).
// Calling SetIndent("", "") disables indentation.
func (enc *Encoder) SetIndent(prefix, indent string) {
	enc.encoder.SetIndent(prefix, indent)
}

// SetEscapeHTML specifies whether problematic HTML characters
// should be escaped inside JSON quoted strings.
// The default behavior is to escape &, <, and > to \u0026, \u003c, and \u003e
// to avoid certain safety problems that can arise when embedding JSON in HTML.
//
// In non-HTML settings where the escaping interferes with the readability
// of the output, SetEscapeHTML(false) disables this behavior.
func (enc *Encoder) SetEscapeHTML(on bool) {
}

// RawMessage is an alias of json.RawMessage.
type RawMessage = json.RawMessage

// Token is an alias of json.Token.
type Token = json.Token

// Delim is an alias of json.Delim.
type Delim = json.Delim
