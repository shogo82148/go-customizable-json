package customizablejson

import (
	"reflect"
	"testing"
	"time"
)

func ptrTime(v time.Time) *time.Time { return &v }
func ptrString(v string) *string     { return &v }
func ptrByteSlice(v []byte) *[]byte  { return &v }

type Small struct {
	Tag string
}

func TestUnmarshal(t *testing.T) {
	testcases := []struct {
		input    string
		ptr      interface{}
		want     interface{}
		register func() *JSONDecoder
	}{
		{
			input: `"Mon Jan  2 15:04:05 2006"`,
			ptr:   new(time.Time),
			want:  ptrTime(time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC)),
			register: func() *JSONDecoder {
				dec := new(JSONDecoder)
				dec.Register(new(time.Time), func(v interface{}, data []byte) error {
					t, err := time.Parse(`"`+time.ANSIC+`"`, string(data))
					if err != nil {
						return err
					}
					vv := v.(*time.Time)
					*vv = t
					return nil
				})
				return dec
			},
		},
		{
			input: `"Mon Jan  2 15:04:05 2006"`,
			ptr:   new(string),
			want:  ptrString("Mon Jan  2 15:04:05 2006"),
			register: func() *JSONDecoder {
				dec := new(JSONDecoder)
				return dec
			},
		},
		{
			input: `"YmFzZTY0IGVuY29kZWQgc3RyaW5n"`,
			ptr:   new([]byte),
			want:  ptrByteSlice([]byte("base64 encoded string")),
			register: func() *JSONDecoder {
				dec := new(JSONDecoder)
				return dec
			},
		},
		{
			input: `{
				"Int": 2,
				"Int8": 3,
				"Int16": 4,
				"Int32": 5,
				"Int64": 6,
				"Uint": 7,
				"Uint8": 8,
				"Uint16": 9,
				"Uint32": 10,
				"Uint64": 11,
				"Uintptr": 12,
				"Float32": 13,
				"Float64": 14
			}`,
			ptr: new(struct {
				Int     int
				Int8    int8
				Int16   int16
				Int32   int32
				Int64   int64
				Uint    uint
				Uint8   uint8
				Uint16  uint16
				Uint32  uint32
				Uint64  uint64
				Uintptr uintptr
				Float32 float32
				Float64 float64
			}),
			want: &struct {
				Int     int
				Int8    int8
				Int16   int16
				Int32   int32
				Int64   int64
				Uint    uint
				Uint8   uint8
				Uint16  uint16
				Uint32  uint32
				Uint64  uint64
				Uintptr uintptr
				Float32 float32
				Float64 float64
			}{
				Int:     2,
				Int8:    3,
				Int16:   4,
				Int32:   5,
				Int64:   6,
				Uint:    7,
				Uint8:   8,
				Uint16:  9,
				Uint32:  10,
				Uint64:  11,
				Uintptr: 12,
				Float32: 13,
				Float64: 14,
			},
			register: func() *JSONDecoder {
				dec := new(JSONDecoder)
				return dec
			},
		},
		{
			input: `{
				"slice": [ {"tag":"tag20"}, {"tag":"tag21"} ],
				"sliceP": [ {"tag":"tag22"}, {"tag":"tag23"} ]
			}`,
			ptr: new(struct {
				Slice  []Small
				SliceP []*Small
			}),
			want: &struct {
				Slice  []Small
				SliceP []*Small
			}{
				Slice:  []Small{{Tag: "tag20"}, {Tag: "tag21"}},
				SliceP: []*Small{{Tag: "tag22"}, {Tag: "tag23"}},
			},
			register: func() *JSONDecoder {
				dec := new(JSONDecoder)
				return dec
			},
		},
		{
			input: `{"null":null}`,
			ptr: &struct {
				Null map[string]interface{}
			}{
				Null: map[string]interface{}{
					"this should be cleared": "",
				},
			},
			want: new(struct {
				Null map[string]interface{}
			}),
			register: func() *JSONDecoder {
				dec := new(JSONDecoder)
				return dec
			},
		},
		{
			input: `{"True":true,"False":false}`,
			ptr: &struct {
				True  bool
				False bool
			}{
				True:  false,
				False: true,
			},
			want: &struct {
				True  bool
				False bool
			}{
				True:  true,
				False: false,
			},
			register: func() *JSONDecoder {
				dec := new(JSONDecoder)
				return dec
			},
		},
		{
			input: `{"Interface": {"foo": 42, "bar": [1, 2]}}`,
			ptr: new(struct {
				Interface interface{}
			}),
			want: &struct {
				Interface interface{}
			}{
				Interface: map[string]interface{}{
					"foo": 42.0,
					"bar": []interface{}{1.0, 2.0},
				},
			},
			register: func() *JSONDecoder {
				dec := new(JSONDecoder)
				return dec
			},
		},
		{
			input: `{"Map": {"foo": 42, "bar": [1, 2]}, "MapInt": {"one": 1, "Answer to the Ultimate Question of Life, the Universe, and Everything": 42}}`,
			ptr: new(struct {
				Map    map[string]interface{}
				MapInt map[string]int
			}),
			want: &struct {
				Map    map[string]interface{}
				MapInt map[string]int
			}{
				Map: map[string]interface{}{
					"foo": 42.0,
					"bar": []interface{}{1.0, 2.0},
				},
				MapInt: map[string]int{
					"one": 1,
					"Answer to the Ultimate Question of Life, the Universe, and Everything": 42,
				},
			},
			register: func() *JSONDecoder {
				dec := new(JSONDecoder)
				return dec
			},
		},
	}
	for i, tc := range testcases {
		ptr := tc.ptr
		dec := tc.register()
		if err := dec.Unmarshal([]byte(tc.input), ptr); err != nil {
			t.Errorf("%d: error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(ptr, tc.want) {
			t.Errorf("want %#v, got %#v", tc.want, ptr)
		}
	}
}
