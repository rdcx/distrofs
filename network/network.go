package network

import (
	"bytes"
	"dfs/client/api"
	"dfs/erasure"
	"dfs/hashutil"
	"dfs/progress"
	"dfs/types"
	"io"
	"math/rand"
	"net/http"
)

type Network struct {
	api   *api.Client
	nodes []*types.Node
}

func (nn *Network) RandomNodesList(n int) ([]*types.Node, error) {

	if len(nn.nodes) < n {
		return nil, types.ErrNotEnoughNodesAvailable
	}
	newList := make([]*types.Node, len(nn.nodes))

	rand.Shuffle(len(nn.nodes), func(i, j int) {
		newList[j], newList[i] = nn.nodes[i], nn.nodes[j]
	})

	return newList[:n], nil
}

func (n *Network) GetNode(nodeID types.NodeID) (*types.Node, error) {
	var found *types.Node

	for _, node := range n.nodes {
		if node.ID == nodeID {
			found = node
			break
		}
	}

	if found == nil {
		return nil, types.ErrNodeNotFound
	}

	return found, nil
}

func WithApiClient(api *api.Client) func(*Network) {
	return func(nn *Network) {
		nn.api = api
	}
}

func WithNodes(nodes []*types.Node) func(*Network) {
	return func(nn *Network) {
		nn.nodes = nodes
	}
}

func NewNetwork(opts ...func(*Network)) *Network {
	nn := &Network{
		nodes: make([]*types.Node, 0),
	}

	for _, opt := range opts {
		opt(nn)
	}

	return nn
}

func (nn *Network) WriteObject(obj *types.Object, r io.Reader, progress progress.BytesReadWithTotal) error {
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

func (nn *Network) WriteSegment(segment *types.Segment, r io.Reader, pc progress.BytesRead) error {

	// check there are enough nodes available
	randomNodes, err := nn.RandomNodesList(80)

	if err != nil {
		return err
	}

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
			NodeID:   randomNodes[i].ID,
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

func (nn *Network) WritePiece(piece *types.Piece, data []byte) error {
	node, err := nn.GetNode(piece.NodeID)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", node.HttpAddr+"/pieces/"+piece.ID.String(), bytes.NewBuffer(data))

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

func (nn *Network) ReadObject(obj *types.Object, w io.Writer, progress progress.BytesReadWithTotal) error {
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

func (nn *Network) ReadSegment(segment *types.Segment, w io.Writer, pc progress.BytesRead) error {

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

func (nn *Network) ReadPiece(piece *types.Piece) ([]byte, error) {
	node, err := nn.GetNode(piece.NodeID)

	if err != nil {
		return nil, err
	}

	res, err := http.Get(node.HttpAddr + "/pieces/" + piece.ID.String())

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
