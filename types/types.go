package types

import "github.com/google/uuid"

type ObjectID uuid.UUID

func NewObjectID() ObjectID {
	return ObjectID(uuid.New())
}

type SegmentID uuid.UUID

func NewSegmentID() SegmentID {
	return SegmentID(uuid.New())
}

type PieceID uuid.UUID

func NewPieceID() PieceID {
	return PieceID(uuid.New())
}

type Piece struct {
	ID PieceID `json:"id"`
}

type Segment struct {
	ID       SegmentID `json:"id"`
	Position uint      `json:"position"`
	Pieces   []Piece   `json:"pieces"`
}

type Object struct {
	ID   ObjectID `json:"id"`
	Name string   `json:"name"`

	Segments []Segment `json:"segments"`
}

func NewObject(name string) Object {
	return Object{
		ID:   ObjectID(uuid.New()),
		Name: name,
	}
}
