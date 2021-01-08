package customizablejson

import (
	"bytes"
	"encoding"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unicode"
)

type encodeState struct {
	enc *JSONEncoder
}

type toInterfaceFunc func(state *encodeState, v reflect.Value) (interface{}, error)

// JSONEncoder xxx
type JSONEncoder struct {
	cache      sync.Map // map[reflect.Type]toInterfaceFunc
	fieldCache sync.Map // map[reflect.Type]structFields
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

func invalidValueToInterface(state *encodeState, v reflect.Value) (interface{}, error) {
	return nil, nil
}

func (enc *JSONEncoder) valueEncoder(v reflect.Value) toInterfaceFunc {
	if !v.IsValid() {
		return invalidValueToInterface
	}
	return enc.typeEncoder(v.Type())
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
	case reflect.Interface:
		return enc.interfaceToInterface
	case reflect.Struct:
		return enc.newStructToInterface(t)
	case reflect.Map:
		return enc.newMapToInterface(t)
	case reflect.Slice:
		return enc.newSliceToInterface(t)
	case reflect.Array:
		return enc.newArrayToInterface(t)
	}
	return interfaceToInterface
}

func (enc *JSONEncoder) interfaceToInterface(state *encodeState, v reflect.Value) (interface{}, error) {
	if v.IsNil() {
		return nil, nil
	}
	return state.reflectToInterface(v.Elem())
}

func interfaceToInterface(state *encodeState, v reflect.Value) (interface{}, error) {
	return v.Interface(), nil
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

type structEncoder struct {
	fields structFields
}

type structFields struct {
	list      []field
	nameIndex map[string]int
}

func (enc *structEncoder) toInterface(state *encodeState, v reflect.Value) (interface{}, error) {
	ret := map[string]interface{}{}

FieldLoop:
	for i := range enc.fields.list {
		f := &enc.fields.list[i]

		// Find the nested struct field by following f.index.
		fv := v
		for _, i := range f.index {
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					continue FieldLoop
				}
				fv = fv.Elem()
			}
			fv = fv.Field(i)
		}

		if f.omitEmpty && isEmptyValue(fv) {
			continue
		}
		vv, err := f.encoder(state, fv)
		if err != nil {
			return nil, err
		}
		ret[f.name] = vv
	}
	return ret, nil
}

func (enc *JSONEncoder) newStructToInterface(t reflect.Type) toInterfaceFunc {
	e := &structEncoder{enc.cachedTypeFields(t)}
	return e.toInterface
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

func (enc *JSONEncoder) newSliceToInterface(t reflect.Type) toInterfaceFunc {
	// Byte slices get special treatment; arrays don't.
	if t.Elem().Kind() == reflect.Uint8 {
		p := reflect.PtrTo(t.Elem())
		if !p.Implements(marshalerType) && !p.Implements(textMarshalerType) {
			return interfaceToInterface
		}
	}
	e := sliceEncoder{enc.newArrayToInterface(t)}
	return e.toInterface
}

// sliceEncoder just wraps an arrayEncoder, checking to make sure the value isn't nil.
type sliceEncoder struct {
	arrayEnc toInterfaceFunc
}

func (enc sliceEncoder) toInterface(state *encodeState, v reflect.Value) (interface{}, error) {
	if v.IsNil() {
		return nil, nil
	}
	return enc.arrayEnc(state, v)
}

func (enc *JSONEncoder) newArrayToInterface(t reflect.Type) toInterfaceFunc {
	e := arrayEncoder{enc.typeEncoder(t.Elem())}
	return e.toInterface
}

type arrayEncoder struct {
	elemEnc toInterfaceFunc
}

func (enc arrayEncoder) toInterface(state *encodeState, v reflect.Value) (interface{}, error) {
	n := v.Len()
	ret := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		elem, err := enc.elemEnc(state, v.Index(i))
		if err != nil {
			return nil, err
		}
		ret = append(ret, elem)
	}
	return ret, nil
}

func (state *encodeState) reflectToInterface(v reflect.Value) (interface{}, error) {
	return state.enc.valueEncoder(v)(state, v)
}

func (state *encodeState) toInterface(v interface{}) (interface{}, error) {
	return state.reflectToInterface(reflect.ValueOf(v))
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

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
			return false
		}
	}
	return true
}

func typeByIndex(t reflect.Type, index []int) reflect.Type {
	for _, i := range index {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		t = t.Field(i).Type
	}
	return t
}

// A field represents a single field found in a struct.
type field struct {
	name      string
	nameBytes []byte                 // []byte(name)
	equalFold func(s, t []byte) bool // bytes.EqualFold or equivalent

	nameEscHTML string // HTMLEscape(name)

	tag       bool
	index     []int
	typ       reflect.Type
	omitEmpty bool
	quoted    bool

	encoder toInterfaceFunc
}

// byIndex sorts field by index sequence.
type byIndex []field

func (x byIndex) Len() int { return len(x) }

func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}

// typeFields returns a list of fields that JSON should recognize for the given type.
// The algorithm is breadth-first search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
func (enc *JSONEncoder) typeFields(t reflect.Type) structFields {
	// Anonymous fields to explore at the current level and the next.
	current := []field{}
	next := []field{{typ: t}}

	// Count of queued names for current level and the next.
	var count, nextCount map[reflect.Type]int

	// Types already visited at an earlier level.
	visited := map[reflect.Type]bool{}

	// Fields found.
	var fields []field

	// Buffer to run HTMLEscape on field names.
	var nameEscBuf bytes.Buffer

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// Scan f.typ for fields to include.
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				isUnexported := sf.PkgPath != ""
				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					}
					if isUnexported && t.Kind() != reflect.Struct {
						// Ignore embedded fields of unexported non-struct types.
						continue
					}
					// Do not ignore embedded fields of unexported struct types
					// since they may have exported fields.
				} else if isUnexported {
					// Ignore unexported non-embedded fields.
					continue
				}
				tag := sf.Tag.Get("json")
				if tag == "-" {
					continue
				}
				name, opts := parseTag(tag)
				if !isValidTag(name) {
					name = ""
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					// Follow pointer.
					ft = ft.Elem()
				}

				// Only strings, floats, integers, and booleans can be quoted.
				quoted := false
				if opts.Contains("string") {
					switch ft.Kind() {
					case reflect.Bool,
						reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
						reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
						reflect.Float32, reflect.Float64,
						reflect.String:
						quoted = true
					}
				}

				// Record found field and index sequence.
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					field := field{
						name:      name,
						tag:       tagged,
						index:     index,
						typ:       ft,
						omitEmpty: opts.Contains("omitempty"),
						quoted:    quoted,
					}
					field.nameBytes = []byte(field.name)
					field.equalFold = foldFunc(field.nameBytes)

					// Build nameEscHTML and nameNonEsc ahead of time.
					nameEscBuf.Reset()
					HTMLEscape(&nameEscBuf, field.nameBytes)
					field.nameEscHTML = nameEscBuf.String()

					fields = append(fields, field)
					if count[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, field{name: ft.Name(), index: index, typ: ft})
				}
			}
		}
	}

	sort.Slice(fields, func(i, j int) bool {
		x := fields
		// sort field by name, breaking ties with depth, then
		// breaking ties with "name came from json tag", then
		// breaking ties with index sequence.
		if x[i].name != x[j].name {
			return x[i].name < x[j].name
		}
		if len(x[i].index) != len(x[j].index) {
			return len(x[i].index) < len(x[j].index)
		}
		if x[i].tag != x[j].tag {
			return x[i].tag
		}
		return byIndex(x).Less(i, j)
	})

	// Delete all fields that are hidden by the Go rules for embedded fields,
	// except that fields with JSON tags are promoted.

	// The fields are sorted in primary order of name, secondary order
	// of field index length. Loop over names; for each name, delete
	// hidden fields by choosing the one dominant field that survives.
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		// One iteration per name.
		// Find the sequence of fields with the name of this first field.
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one field with this name
			out = append(out, fi)
			continue
		}
		dominant, ok := dominantField(fields[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	fields = out
	sort.Sort(byIndex(fields))

	for i := range fields {
		f := &fields[i]
		f.encoder = enc.typeEncoder(typeByIndex(t, f.index))
	}
	nameIndex := make(map[string]int, len(fields))
	for i, field := range fields {
		nameIndex[field.name] = i
	}
	return structFields{fields, nameIndex}
}

// dominantField looks through the fields, all of which are known to
// have the same name, to find the single field that dominates the
// others using Go's embedding rules, modified by the presence of
// JSON tags. If there are multiple top-level fields, the boolean
// will be false: This condition is an error in Go and we skip all
// the fields.
func dominantField(fields []field) (field, bool) {
	// The fields are sorted in increasing index-length order, then by presence of tag.
	// That means that the first field is the dominant one. We need only check
	// for error cases: two fields at top level, either both tagged or neither tagged.
	if len(fields) > 1 && len(fields[0].index) == len(fields[1].index) && fields[0].tag == fields[1].tag {
		return field{}, false
	}
	return fields[0], true
}

func (enc *JSONEncoder) cachedTypeFields(t reflect.Type) structFields {
	if f, ok := enc.fieldCache.Load(t); ok {
		return f.(structFields)
	}
	f, _ := enc.fieldCache.LoadOrStore(t, enc.typeFields(t))
	return f.(structFields)
}
