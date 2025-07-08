package codecs_test

import (
	"reflect"
	"testing"

	"github.com/dimmerz92/sittella/codecs"
)

func TestJSONCodec(t *testing.T) {
	encoder := codecs.JSONCodec{}

	t.Run("valid data", func(t *testing.T) {
		expected := map[string]string{"hello": "world"}

		encoded, err := encoder.Marshal(expected)
		if err != nil {
			t.Fatalf("expected successful marshal, got %v", err)
		}

		var out map[string]string
		if err := encoder.Unmarshal(encoded, &out); err != nil {
			t.Fatalf("expected successful unmarshal, got %v", err)
		}

		if !reflect.DeepEqual(expected, out) {
			t.Fatalf("expected %+v, got %+v", expected, out)
		}
	})

	t.Run("invalid data", func(t *testing.T) {
		if _, err := encoder.Marshal(make(chan struct{})); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
