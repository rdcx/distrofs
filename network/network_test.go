package network_test

import (
	"dfs/erasure"
	"dfs/types"
	"fmt"
	"os"
	"testing"
)

func TestReadObject(t *testing.T) {
	t.Run("can read object with one segment", func(t *testing.T) {
		dummyData := make([]byte, types.SEGMENT_SIZE)

		enc := erasure.NewReedSolomonEncoder(29, 51)

		shards, err := enc.Encode(dummyData)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fmt.Printf("len of shard[0]: %d\n", len(shards[0]))

		os.Exit(0)
	})
}

func TestReadSegment(t *testing.T) {
	t.Run("can read segment", func(t *testing.T) {
	})
}

func TestReadPiece(t *testing.T) {
	t.Run("can read piece", func(t *testing.T) {
	})
}
