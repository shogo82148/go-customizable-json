package customizablejson

import (
	"reflect"
	"testing"
	"time"
)

func ptrTime(v time.Time) *time.Time { return &v }

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
