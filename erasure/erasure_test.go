package erasure

import (
	"fmt"
	"testing"
)

func TestErasure(t *testing.T) {
	t.Run("can encode data", func(t *testing.T) {
		data := []byte("hello world")
		rs := NewReedSolomonEncoder(3, 2)

		shards, err := rs.Encode(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(shards) != 5 {
			t.Errorf("expected 5 shards, got %d", len(shards))
		}
	})

	t.Run("can decode data", func(t *testing.T) {
		data := []byte("hello world")
		rs := NewReedSolomonEncoder(3, 2)

		shards, err := rs.Encode(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		actual, err := rs.Reconstruct(shards)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fmt.Printf("len of data: %d\n", len(data))
		fmt.Printf("len of decoded: %d\n", len(actual))

		if string(actual) != string(data) {
			t.Errorf("expected %s, got %s", string(data), string(actual))
		}
	})

	t.Run("can can correct errors", func(t *testing.T) {
		raw := []byte("hello world does this work nicely?")
		rs := NewReedSolomonEncoder(29, 51)

		shards, err := rs.Encode(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// corrupt the data
		shards[0] = nil
		shards[4] = nil

		data, err := rs.Reconstruct(shards)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(data[:len(raw)]) != string(raw) {
			t.Errorf("expected %s, got %s", string(raw), string(data))
		}
	})
}
