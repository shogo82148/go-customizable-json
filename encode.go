package customizablejson

import (
	"bytes"
	"encoding"
	"encoding/json"
	"reflect"
	"sync"
)

type encodeState struct {
	enc *JSONEncoder
}

type toInterfaceFunc func(state *encodeState, v reflect.Value) (interface{}, error)

// JSONEncoder xxx
type JSONEncoder struct {
	cache sync.Map
}

// Register records a type and a function for encoding.
func (enc *JSONEncoder) Register(val interface{}, f func(v interface{}) ([]byte, error)) {
	typ := reflect.TypeOf(val)
	enc.cache.Store(typ, toInterfaceFunc(func(_ *encodeState, v reflect.Value) (interface{}, error) {
		data, err := f(v.Interface())
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
	state := &encodeState{enc: enc}
	ret, err := state.toInterface(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ret)
}

func (enc *JSONEncoder) typeEncoder(t reflect.Type) toInterfaceFunc {
	if fi, ok := enc.cache.Load(t); ok {
		return fi.(toInterfaceFunc)
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it. This indirect
	// func is only used for recursive types.
	var (
		wg sync.WaitGroup
		f  toInterfaceFunc
	)
	wg.Add(1)
	fi, loaded := enc.cache.LoadOrStore(t, toInterfaceFunc(func(state *encodeState, v reflect.Value) (interface{}, error) {
		wg.Wait()
		return f(state, v)
	}))
	if loaded {
		return fi.(toInterfaceFunc)
	}

	// Compute the real encoder and replace the indirect func with it.
	f = enc.newTypeEncoder(t, true)
	wg.Done()
	enc.cache.Store(t, f)
	return f
}

var (
	marshalerType     = reflect.TypeOf((*Marshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func (enc *JSONEncoder) newTypeEncoder(t reflect.Type, allowAddr bool) toInterfaceFunc {
	if t.Implements(marshalerType) {
		return interfaceToInterface
	}
	if t.Implements(textMarshalerType) {
		return interfaceToInterface
	}
	switch t.Kind() {
	case reflect.Struct:
		return newStructToInterface(t)
	case reflect.Map:
		return enc.newMapToInterface(t)
	case reflect.Slice:
		panic("TODO: implement me")
	case reflect.Array:
		panic("TODO: implement me")
	}
	return interfaceToInterface
}

func interfaceToInterface(state *encodeState, v reflect.Value) (interface{}, error) {
	return v.Interface(), nil
}

func newStructToInterface(t reflect.Type) toInterfaceFunc {
	return interfaceToInterface // TODO: implement me
}

type mapEncoder struct {
	retType reflect.Type
	elemEnc toInterfaceFunc
}

func (enc *mapEncoder) toInterface(state *encodeState, v reflect.Value) (interface{}, error) {
	if v.IsNil() {
		return nil, nil
	}
	ret := reflect.MakeMap(enc.retType)
	keys := v.MapKeys()
	for _, key := range keys {
		elem, err := enc.elemEnc(state, v.MapIndex(key))
		if err != nil {
			return nil, err
		}
		ret.SetMapIndex(key, reflect.ValueOf(elem))
	}
	return ret.Interface(), nil
}

func (enc *JSONEncoder) newMapToInterface(t reflect.Type) toInterfaceFunc {
	e := &mapEncoder{
		retType: reflect.MapOf(t.Key(), reflect.TypeOf((*interface{})(nil)).Elem()),
		elemEnc: enc.typeEncoder(t.Elem()),
	}
	return e.toInterface
}

func (state *encodeState) toInterface(v interface{}) (interface{}, error) {
	typ := reflect.TypeOf(v)
	f := state.enc.typeEncoder(typ)
	return f(state, reflect.ValueOf(v))
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
