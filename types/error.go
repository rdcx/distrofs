package types

import "errors"

var ErrCouldNotGetObjectFromAPI = errors.New("could not get object from API")
var ErrCouldNotPutObjectToAPI = errors.New("could not put object to API")

var ErrPieceHashMismatch = errors.New("piece hash mismatch")
var ErrNotEnoughPieces = errors.New("not enough pieces")
