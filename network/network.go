package network

import (
	"dfs/client/api"
	"dfs/progress"
	"dfs/types"
	"io"
)

type Node struct {
	Addr string
}

type NodeNetwork struct {
	nodes map[string]*Node
}

func NewNodeNetwork(api *api.Client) *NodeNetwork {
	return &NodeNetwork{
		nodes: make(map[string]*Node),
	}
}

func (nn *NodeNetwork) ReadObject(obj *types.Object, w io.WriteCloser, pc progress.Callback) error {
	var total uint64
	for _, segment := range obj.Segments {
		total += uint64(len(segment.Pieces))
	}

	for _, segment := range obj.Segments {

		var buf []byte
		nn.ReadPiece(segment.Pieces, buf, pc)
	}

	// get each segment from the network
	// and write it to the writer, all of the orders must be preserved

	return nil
}

func (nn *NodeNetwork) ReconstructSegment(pieces []types.Piece) []byte {
	// reconstruct the segment from the first 29 pieces
	return nil
}

func (nn *NodeNetwork) WriteData(obj *types.Object, r io.ReadCloser, pc progress.Callback) error {
	var total uint64
	for _, segment := range obj.Segments {
		total += uint64(len(segment.Pieces))
	}

	var completed uint64 = total
	var progress = float64(completed) / float64(total)
	pc(progress, 0)

	// split the data into pieces
	// and send each piece to the network

	return nil
}
