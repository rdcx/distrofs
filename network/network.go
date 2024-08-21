package network

import (
	"bytes"
	"dfs/client/api"
	"dfs/erasure"
	"dfs/hashutil"
	"dfs/progress"
	"dfs/types"
	"io"
	"net/http"
)

type Node struct {
	HttpAddr string
	GRPCAddr string
}

type NodeNetwork struct {
	api   *api.Client
	nodes map[string]*Node
}

func WithApiClient(api *api.Client) func(*NodeNetwork) {
	return func(nn *NodeNetwork) {
		nn.api = api
	}
}

func WithNodes(nodes map[string]*Node) func(*NodeNetwork) {
	return func(nn *NodeNetwork) {
		nn.nodes = nodes
	}
}

func NewNodeNetwork(opts ...func(*NodeNetwork)) *NodeNetwork {
	nn := &NodeNetwork{
		nodes: make(map[string]*Node),
	}

	for _, opt := range opts {
		opt(nn)
	}

	return nn
}

func (nn *NodeNetwork) WriteObject(obj *types.Object, r io.Reader, progress progress.BytesReadWithTotal) error {
	var totalBytesRead uint64 = 0
	var segmentProgress = func(bytesRead uint64) error {
		totalBytesRead += bytesRead
		progress(totalBytesRead, obj.Size)
		return nil
	}

	for _, segment := range obj.Segments {
		err := nn.WriteSegment(segment, r, segmentProgress)

		if err != nil {
			return err
		}
	}

	return nil
}

func (nn *NodeNetwork) WriteSegment(segment *types.Segment, r io.Reader, pc progress.BytesRead) error {
	data, err := io.ReadAll(io.LimitReader(r, int64(segment.Size)))

	if err != nil {
		return err
	}

	enc := erasure.NewReedSolomonEncoder(29, 51)

	shards, err := enc.Encode(data)

	if err != nil {
		return err
	}

	for i, shard := range shards {
		piece := &types.Piece{
			ID:       types.NewPieceID(),
			Hash:     hashutil.Blake3(shard),
			Position: uint(i),
			Addr:     "http://localhost:8080",
		}

		segment.Pieces = append(segment.Pieces, piece)
	}

	for i, shard := range shards {
		err = nn.WritePiece(segment.Pieces[i], shard)

		if err != nil {
			return err
		}
	}

	err = nn.api.CreateSegment(segment)

	if err != nil {
		return err
	}

	return nil
}

func (nn *NodeNetwork) WritePiece(piece *types.Piece, data []byte) error {
	req, err := http.NewRequest("POST", piece.Addr+"/pieces/"+piece.ID.String(), bytes.NewBuffer(data))

	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	return nil
}

func (nn *NodeNetwork) ReadObject(obj *types.Object, w io.Writer, progress progress.BytesReadWithTotal) error {
	var totalBytesRead uint64 = 0
	var segmentProgress = func(bytesRead uint64) error {
		totalBytesRead += bytesRead
		progress(totalBytesRead, obj.Size)
		return nil
	}

	for _, segment := range obj.Segments {

		err := nn.ReadSegment(segment, w, segmentProgress)

		if err != nil {
			return err
		}
	}

	return nil
}

func (nn *NodeNetwork) ReadSegment(segment *types.Segment, w io.Writer, pc progress.BytesRead) error {

	var segData [][]byte = make([][]byte, 80)

	var successPiecesCount uint = 0
	for _, piece := range segment.Pieces {
		if successPiecesCount == 29 {
			break
		}
		data, err := nn.ReadPiece(piece)
		if err != nil {
			segData[piece.Position] = nil
			continue
		}

		segData[piece.Position] = data
		successPiecesCount++
	}

	if successPiecesCount < 29 {
		return types.ErrNotEnoughPieces
	}

	enc := erasure.NewReedSolomonEncoder(29, 51)

	data, err := enc.Reconstruct(segData)

	if err != nil {
		return err
	}

	_, err = w.Write(data[:segment.Size])

	if err != nil {
		return err
	}

	return nil
}

func (nn *NodeNetwork) ReadPiece(piece *types.Piece) ([]byte, error) {
	res, err := http.Get(piece.Addr + "/pieces/" + piece.ID.String())

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	if hash := hashutil.Blake3(data); !bytes.Equal(hash, piece.Hash) {
		return nil, types.ErrPieceHashMismatch
	}

	return data, nil
}
