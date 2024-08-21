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
	return &NodeNetwork{
		nodes: make(map[string]*Node),
	}
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

func (nn *NodeNetwork) ReadSegment(segment types.Segment, w io.Writer, pc progress.BytesRead) error {

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

func (nn *NodeNetwork) ReadPiece(piece types.Piece) ([]byte, error) {
	res, err := http.Get(piece.Addr + "/piece/" + piece.ID.String())

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
