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