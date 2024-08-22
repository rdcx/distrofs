package network_test

import (
	"bytes"
	"dfs/client/api"
	"dfs/erasure"
	"dfs/hashutil"
	"dfs/network"
	"dfs/types"
	"fmt"
	"testing"

	"github.com/h2non/gock"
)

func TestRandomNodesList(t *testing.T) {
	t.Run("can get random nodes list", func(t *testing.T) {
		nodes := []*types.Node{
			{
				ID:       types.NewNodeID(),
				HttpAddr: "http://node1",
			},
			{
				ID:       types.NewNodeID(),
				HttpAddr: "http://node2",
			},
			{
				ID:       types.NewNodeID(),
				HttpAddr: "http://node3",
			},
		}

		nn := network.NewNetwork(
			network.WithNodes(nodes),
		)

		list, err := nn.RandomNodesList(2)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(list) != 2 {
			t.Fatalf("expected list to have 2 nodes")
		}

		for _, node := range list {
			if node == nil {
				t.Fatalf("expected node to be not nil")
			}
		}
	})
}

func TestWriteSegment(t *testing.T) {
	t.Run("can write segment", func(t *testing.T) {
		enc := erasure.NewReedSolomonEncoder(29, 51)

		data := []byte("hello world")

		shards, err := enc.Encode(data)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		segment := &types.Segment{
			ID:       types.NewSegmentID(),
			ObjectID: types.NewObjectID(),
			Size:     uint64(len(data)),
			Position: 0,
		}

		nodes := []*types.Node{}

		for i := 0; i < 80; i++ {
			node := &types.Node{
				ID:       types.NewNodeID(),
				HttpAddr: fmt.Sprintf("http://node-%d", i),
			}
			nodes = append(nodes, node)
			piece := &types.Piece{
				ID:       types.NewPieceID(),
				Hash:     hashutil.Blake3(shards[i]),
				Position: uint(i),
				NodeID:   node.ID,
			}

			segment.Pieces = append(segment.Pieces, piece)
		}

		defer gock.Off()

		gock.New("http://localhost:8080").
			Post("/objects/" + segment.ObjectID.String() + "/segments").
			Reply(200).
			JSON(`{"status":"ok"}`)

		for i, shard := range shards {
			gock.New(fmt.Sprintf("http://node-%d", i)).
				Post("/pieces").
				Reply(200).
				Body(bytes.NewBuffer(shard))
		}

		nn := network.NewNetwork(
			network.WithApiClient(
				api.NewClient("http://localhost:8080", "test"),
			),
			network.WithNodes(nodes),
		)

		err = nn.WriteSegment(segment, bytes.NewReader(data), nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestWritePiece(t *testing.T) {
	t.Run("can write piece", func(t *testing.T) {
		data := []byte("hello world")

		pieceID := types.NewPieceID()

		defer gock.Off()

		gock.New("http://localhost:9090").
			Post("/pieces").
			Reply(200).
			Body(bytes.NewBuffer(data))

		nodes := []*types.Node{
			{
				ID:       types.NewNodeID(),
				HttpAddr: "http://localhost:9090",
			}}

		nn := network.NewNetwork(
			network.WithNodes(nodes),
		)

		piece := &types.Piece{
			ID:       pieceID,
			Hash:     hashutil.Blake3(data),
			Position: 0,
			NodeID:   nodes[0].ID,
		}

		err := nn.WritePiece(piece, data)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestReadObject(t *testing.T) {
	t.Run("can read object with one segment", func(t *testing.T) {
		dummyData := make([]byte, types.SEGMENT_SIZE)

		enc := erasure.NewReedSolomonEncoder(29, 51)

		shards, err := enc.Encode(dummyData)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		obj := types.NewObject("test")

		segment := &types.Segment{
			ID:       types.NewSegmentID(),
			Size:     uint64(len(dummyData)),
			Position: 0,
		}

		nodes := []*types.Node{}

		for i := 0; i < 80; i++ {
			node := &types.Node{
				ID:       types.NewNodeID(),
				HttpAddr: fmt.Sprintf("http://node-%d", i),
			}
			nodes = append(nodes, node)
			piece := &types.Piece{
				ID:       types.NewPieceID(),
				Hash:     hashutil.Blake3(shards[i]),
				Position: uint(i),
				NodeID:   node.ID,
			}

			// simulate 51 corrupted pieces
			if i > 3 && i < 55 {
				piece.Hash = hashutil.Blake3([]byte("corrupted"))
			}

			segment.Pieces = append(segment.Pieces, piece)
		}

		obj.Segments = append(obj.Segments, segment)

		defer gock.Off()

		for i, shard := range shards {
			gock.New(fmt.Sprintf("http://node-%d", i)).
				Get("/pieces/" + segment.Pieces[i].ID.String()).
				Reply(200).
				Body(bytes.NewBuffer(shard))
		}

		nn := network.NewNetwork(
			network.WithNodes(nodes),
		)

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

		segment := &types.Segment{
			ID:       types.NewSegmentID(),
			Size:     uint64(len(data)),
			Position: 0,
		}
		nodes := []*types.Node{}
		for i := 0; i < 80; i++ {
			node := &types.Node{
				ID:       types.NewNodeID(),
				HttpAddr: "http://localhost:8080",
			}
			nodes = append(nodes, node)
			piece := &types.Piece{
				ID:       types.NewPieceID(),
				Hash:     hashutil.Blake3(shards[i]),
				Position: uint(i),
				NodeID:   node.ID,
			}

			segment.Pieces = append(segment.Pieces, piece)
		}

		defer gock.Off()

		for i, shard := range shards {
			gock.New("http://localhost:8080").
				Get("/pieces/" + segment.Pieces[i].ID.String()).
				Reply(200).
				Body(bytes.NewBuffer(shard))
		}

		nn := network.NewNetwork(
			network.WithNodes(nodes),
		)

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

		gock.New("http://node:8080").
			Get("/pieces/" + pieceID.String()).
			Reply(200).
			Body(bytes.NewBuffer(data))

		nodes := []*types.Node{
			{
				ID:       types.NewNodeID(),
				HttpAddr: "http://node:8080",
			},
		}

		nn := network.NewNetwork(
			network.WithNodes(nodes),
		)

		piece := &types.Piece{
			ID:       pieceID,
			Hash:     hashutil.Blake3(data),
			Position: 0,
			NodeID:   nodes[0].ID,
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
