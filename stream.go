package customizablejson

import (
	"encoding/json"
	"io"
)

// A Decoder reads and decodes JSON values from an input stream.
type Decoder struct {
	// TODO: fill me
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{}
}

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a
// Number instead of as a float64.
func (dec *Decoder) UseNumber() {}

// DisallowUnknownFields causes the Decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination.
func (dec *Decoder) DisallowUnknownFields() {}

// Decode xxx
func (dec *Decoder) Decode(v interface{}) error {
	panic("TODO: implement me")
}

// An Encoder writes JSON values to an output stream.
type Encoder struct {
	// TODO: fill me
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{}
}

// Encode xxx
func (enc *Encoder) Encode(v interface{}) error {
	panic("TODO: implement me")
}

// SetIndent instructs the encoder to format each subsequent encoded
// value as if indented by the package-level function Indent(dst, src, prefix, indent).
// Calling SetIndent("", "") disables indentation.
func (enc *Encoder) SetIndent(prefix, indent string) {
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
