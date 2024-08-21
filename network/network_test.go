package network_test

import (
	"bytes"
	"dfs/erasure"
	"dfs/hashutil"
	"dfs/network"
	"dfs/types"
	"testing"

	"github.com/h2non/gock"
)

func TestReadObject(t *testing.T) {
	t.Run("can read object with one segment", func(t *testing.T) {
		dummyData := make([]byte, types.SEGMENT_SIZE)

		enc := erasure.NewReedSolomonEncoder(29, 51)

		shards, err := enc.Encode(dummyData)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		obj := types.NewObject("test")

		segment := types.Segment{
			ID:       types.NewSegmentID(),
			Size:     uint64(len(dummyData)),
			Position: 0,
		}

		for i := 0; i < 80; i++ {
			piece := types.Piece{
				ID:       types.NewPieceID(),
				Hash:     hashutil.Blake3(shards[i]),
				Position: uint(i),
				Addr:     "http://localhost:8080",
			}

			segment.Pieces = append(segment.Pieces, piece)
		}

		obj.Segments = append(obj.Segments, segment)

		defer gock.Off()

		for i, shard := range shards {
			gock.New("http://localhost:8080").
				Get("/piece/" + segment.Pieces[i].ID.String()).
				Reply(200).
				Body(bytes.NewBuffer(shard))
		}

		nn := network.NewNodeNetwork()

		var buf bytes.Buffer

		err = nn.ReadObject(&obj, &buf, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result := buf.Bytes()

		if !bytes.Equal(result, dummyData) {
			t.Fatalf("expected data to be equal")
		}
	})
}

func TestReadSegment(t *testing.T) {
	t.Run("can read segment", func(t *testing.T) {

		enc := erasure.NewReedSolomonEncoder(29, 51)

		data := []byte("hello world")

		shards, err := enc.Encode(data)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		segment := types.Segment{
			ID:       types.NewSegmentID(),
			Size:     uint64(len(data)),
			Position: 0,
		}

		for i := 0; i < 80; i++ {
			piece := types.Piece{
				ID:       types.NewPieceID(),
				Hash:     hashutil.Blake3(shards[i]),
				Position: uint(i),
				Addr:     "http://localhost:8080",
			}

			segment.Pieces = append(segment.Pieces, piece)
		}

		defer gock.Off()

		for i, shard := range shards {
			gock.New("http://localhost:8080").
				Get("/piece/" + segment.Pieces[i].ID.String()).
				Reply(200).
				Body(bytes.NewBuffer(shard))
		}

		nn := network.NewNodeNetwork()

		var buf bytes.Buffer

		err = nn.ReadSegment(segment, &buf, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result := buf.Bytes()

		if !bytes.Equal(result, data) {
			t.Fatalf("expected data to be equal to %s got %s", data, result)
		}
	})
}

func TestReadPiece(t *testing.T) {
	t.Run("can read piece", func(t *testing.T) {

		data := []byte("hello world")

		pieceID := types.NewPieceID()

		defer gock.Off()

		gock.New("http://localhost:8080").
			Get("/piece/" + pieceID.String()).
			Reply(200).
			Body(bytes.NewBuffer(data))

		nn := network.NewNodeNetwork()

		piece := types.Piece{
			ID:       pieceID,
			Position: 0,
			Addr:     "http://localhost:8080",
		}

		result, err := nn.ReadPiece(piece)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !bytes.Equal(result, data) {
			t.Fatalf("expected data to be equal")
		}
	})
}
