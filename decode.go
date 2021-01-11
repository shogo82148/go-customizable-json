package customizablejson

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"sort"
	"sync"
)

type unmarshalFunc func(v interface{}, data []byte) error

// JSONDecoder xxx
type JSONDecoder struct {
	unmarshalFuncs sync.Map // map[reflect]unmarshalFunc
	fieldCache     sync.Map // map[reflect.Type]structFields
}

// Register records a type and a function for encoding.
func (dec *JSONDecoder) Register(val interface{}, f func(v interface{}, data []byte) error) {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr {
		panic("customizablejson: val must be a pointer")
	}
	typ := v.Type()
	dec.unmarshalFuncs.Store(typ, unmarshalFunc(f))
}

// Unmarshal xxx
func (dec *JSONDecoder) Unmarshal(data []byte, v interface{}) error {
	d := dec.NewDecoder(bytes.NewReader(data))
	return d.Decode(v)
}

type unmarshaler struct {
	v interface{}
	f unmarshalFunc
}

func (v unmarshaler) UnmarshalJSON(data []byte) error {
	return v.f(v.v, data)
}

// from the encoding/json package.
// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
// if it encounters an Unmarshaler, indirect stops and returns that.
// if decodingNull is true, indirect stops at the last pointer so it can be set to nil.
func (dec *JSONDecoder) indirect(v reflect.Value, decodingNull bool) (Unmarshaler, encoding.TextUnmarshaler, reflect.Value) {
	// Issue #24153 indicates that it is generally not a guaranteed property
	// that you may round-trip a reflect.Value by calling Value.Addr().Elem()
	// and expect the value to still be settable for values derived from
	// unexported embedded struct fields.
	//
	// The logic below effectively does this when it first addresses the value
	// (to satisfy possible pointer methods) and continues to dereference
	// subsequent pointers as necessary.
	//
	// After the first round-trip, we set v back to the original value to
	// preserve the original RW flags contained in reflect.Value.
	v0 := v
	haveAddr := false

	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		haveAddr = true
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				haveAddr = false
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.Elem().Kind() != reflect.Ptr && decodingNull && v.CanSet() {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if f, ok := dec.unmarshalFuncs.Load(v.Type()); ok {
			u := unmarshaler{
				v: v.Interface(),
				f: f.(unmarshalFunc),
			}
			return u, nil, reflect.Value{}
		}
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(Unmarshaler); ok {
				return u, nil, reflect.Value{}
			}
			if !decodingNull {
				if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
					return nil, u, reflect.Value{}
				}
			}
		}

		if haveAddr {
			v = v0 // restore original value after round-trip Value.Addr().Elem()
			haveAddr = false
		} else {
			v = v.Elem()
		}
	}
	return nil, nil, v
}

func (dec *Decoder) decode(in interface{}, out reflect.Value) error {
	if !out.IsValid() {
		return nil
	}

	u, ut, pv := dec.myDec.indirect(out, in == nil)
	if u != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		return u.UnmarshalJSON(data)
	}
	if ut != nil {
		switch v := in.(type) {
		default:
			return dec.withErrorContext(&UnmarshalTypeError{Type: out.Type()})
		case nil:
			return dec.withErrorContext(&UnmarshalTypeError{Value: "null", Type: out.Type()})
		case bool:
			return dec.withErrorContext(&UnmarshalTypeError{Value: "bool", Type: out.Type()})
		case Number:
			return dec.withErrorContext(&UnmarshalTypeError{Value: "number", Type: out.Type()})
		case string:
			return ut.UnmarshalText([]byte(v))
		}
	}

	out = pv
	switch v := in.(type) {
	case string:
		switch out.Kind() {
		default:
			return dec.withErrorContext(&UnmarshalTypeError{Value: "string", Type: out.Type()})
		case reflect.Slice:
			if out.Type().Elem().Kind() == reflect.Uint8 {
				b, err := base64.StdEncoding.DecodeString(v)
				if err != nil {
					return err
				}
				out.SetBytes(b)
				break
			}
			return dec.withErrorContext(&UnmarshalTypeError{Value: "string", Type: out.Type()})
		case reflect.String:
			out.SetString(v)
		case reflect.Interface:
			if out.NumMethod() == 0 {
				out.Set(reflect.ValueOf(v))
			} else {
				return dec.withErrorContext(&UnmarshalTypeError{Value: "string", Type: out.Type()})
			}
		}
	default:
		panic("TODO: implement me!")
	}
	return nil
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

// typeFields returns a list of fields that JSON should recognize for the given type.
// The algorithm is breadth-first search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
func (dec *JSONDecoder) typeFields(t reflect.Type) structFields {
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

	nameIndex := make(map[string]int, len(fields))
	for i, field := range fields {
		nameIndex[field.name] = i
	}
	return structFields{fields, nameIndex}
}

func (dec *JSONDecoder) cachedTypeFields(t reflect.Type) structFields {
	if f, ok := dec.fieldCache.Load(t); ok {
		return f.(structFields)
	}
	f, _ := dec.fieldCache.LoadOrStore(t, dec.typeFields(t))
	return f.(structFields)
}
