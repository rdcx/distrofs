package types

import "github.com/google/uuid"

type ObjectID uuid.UUID

func (o ObjectID) String() string {
	return uuid.UUID(o).String()
}

func NewObjectID() ObjectID {
	return ObjectID(uuid.New())
}

type SegmentID uuid.UUID

func (s SegmentID) String() string {
	return uuid.UUID(s).String()
}

func NewSegmentID() SegmentID {
	return SegmentID(uuid.New())
}

type PieceID uuid.UUID

func (p PieceID) String() string {
	return uuid.UUID(p).String()
}

func NewPieceID() PieceID {
	return PieceID(uuid.New())
}

type NodeID uuid.UUID

func (n NodeID) String() string {
	return uuid.UUID(n).String()
}

func NewNodeID() NodeID {
	return NodeID(uuid.New())
}

type Node struct {
	ID       NodeID `json:"id"`
	HttpAddr string
	GRPCAddr string
}

type Piece struct {
	ID       PieceID `json:"id"`
	Hash     []byte  `json:"hash"`
	Position uint    `json:"position"`
	NodeID   NodeID  `json:"addr"`
}

type Segment struct {
	ID       SegmentID `json:"id"`
	ObjectID ObjectID  `json:"object_id"`
	Size     uint64    `json:"size"`
	Position uint      `json:"position"`
	Pieces   []*Piece  `json:"pieces"`
}

type Object struct {
	ID   ObjectID `json:"id"`
	Name string   `json:"name"`
	Size uint64   `json:"size"`

	Segments []*Segment `json:"segments"`
}

func NewObject(name string) Object {
	return Object{
		ID:   ObjectID(uuid.New()),
		Name: name,
	}
}

func NewSegment(objectID ObjectID, size uint64, position uint) Segment {
	return Segment{
		ID:       NewSegmentID(),
		ObjectID: objectID,
		Size:     size,
		Position: position,
	}
}
