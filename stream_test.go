package customizablejson

import (
	"bytes"
	"testing"
	"time"
)

func TestEncoder(t *testing.T) {
	enc := new(JSONEncoder)
	enc.Register(time.Time{}, func(v interface{}) ([]byte, error) {
		return []byte(`"` + v.(time.Time).Format(time.RFC3339) + `"`), nil
	})

	var buf bytes.Buffer
	e := enc.NewEncoder(&buf)
	err := e.Encode(map[string]interface{}{
		"foo": time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = e.Encode(map[string]interface{}{
		"foo": time.Date(2006, time.January, 2, 15, 4, 5, 123456000, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	want := "{\"foo\":\"2006-01-02T15:04:05Z\"}\n{\"foo\":\"2006-01-02T15:04:05Z\"}\n"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}
