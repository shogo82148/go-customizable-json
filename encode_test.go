package customizablejson

import (
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	testcases := []struct {
		input    interface{}
		want     string
		register func() *JSONEncoder
	}{
		{
			input: time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			want:  `"2006-01-02T15:04:05Z"`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},
		{
			input: time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			want:  `"2006-01-02T15:04:05.123456000Z"`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format("2006-01-02T15:04:05.000000000Z07:00") + `"`), nil
				})
				return enc
			},
		},

		// maps
		{
			input: map[string]time.Time{
				"foo": time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			},
			want: `{"foo":"2006-01-02T15:04:05Z"}`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},
		{
			input: map[string]map[string]time.Time{
				"foo": {
					"bar": time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
				},
			},
			want: `{"foo":{"bar":"2006-01-02T15:04:05Z"}}`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},
		{
			input: map[string]interface{}{
				"foo": time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			},
			want: `{"foo":"2006-01-02T15:04:05Z"}`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},

		// struct
		{
			input: struct {
				Foo time.Time
			}{
				Foo: time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			},
			want: `{"Foo":"2006-01-02T15:04:05Z"}`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},
		{
			input: &struct {
				Foo time.Time `json:"foo"`
			}{
				Foo: time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			},
			want: `{"foo":"2006-01-02T15:04:05Z"}`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},
		{
			input: struct {
				Foo interface{}
			}{
				Foo: time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			},
			want: `{"Foo":"2006-01-02T15:04:05Z"}`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},

		// slice
		{
			input: []time.Time{
				time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
				time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			},
			want: `["2006-01-02T15:04:05Z","2006-01-02T15:04:05Z"]`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},
		{
			input: []interface{}{
				time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
				time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			},
			want: `["2006-01-02T15:04:05Z","2006-01-02T15:04:05Z"]`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},
		{
			input: []byte("special case: it will be encoded into base64 string"),
			want:  `"c3BlY2lhbCBjYXNlOiBpdCB3aWxsIGJlIGVuY29kZWQgaW50byBiYXNlNjQgc3RyaW5n"`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},

		// string tag
		{
			input: struct {
				Int int `json:"int,string"`
			}{
				Int: 42,
			},
			want: `{"int":"42"}`,
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				return enc
			},
		},
	}
	for i, tc := range testcases {
		enc := tc.register()
		got, err := enc.Marshal(tc.input)
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
		if string(got) != tc.want {
			t.Errorf("#%d: want %q, got %q", i, tc.want, got)
		}
	}
}

func TestMarshal_Cycle(t *testing.T) {
	type Cycle struct {
		Next *Cycle
	}
	cycle := new(Cycle)
	cycle.Next = cycle

	enc := new(JSONEncoder)
	_, err := enc.Marshal(cycle)
	if _, ok := err.(*UnsupportedValueError); !ok {
		t.Errorf("want UnsupportedValueError, got %T", err)
	}
}

func TestMarshalIndent(t *testing.T) {
	testcases := []struct {
		input    interface{}
		want     string
		register func() *JSONEncoder
	}{
		{
			input: map[string]interface{}{
				"foo": time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			},
			want: "{\n        \"foo\": \"2006-01-02T15:04:05Z\"\n    }",
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},

		{
			input: struct {
				Foo interface{}
			}{
				Foo: time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
			},
			want: "{\n        \"Foo\": \"2006-01-02T15:04:05Z\"\n    }",
			register: func() *JSONEncoder {
				enc := new(JSONEncoder)
				enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
					return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
				})
				return enc
			},
		},
	}
	for i, tc := range testcases {
		enc := tc.register()
		got, err := enc.MarshalIndent(tc.input, "    ", "    ")
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
		if string(got) != tc.want {
			t.Errorf("#%d: want %q, got %q", i, tc.want, got)
		}
	}
}
