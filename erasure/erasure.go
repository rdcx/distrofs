package erasure

import (
	"bytes"

	"github.com/klauspost/reedsolomon"
)

type ReedSolomonEncoder struct {
	dataShards   int
	parityShards int
}

func NewReedSolomonEncoder(dataShards, parityShards int) ReedSolomonEncoder {
	return ReedSolomonEncoder{
		dataShards:   dataShards,
		parityShards: parityShards,
	}
}

func (rse ReedSolomonEncoder) Encode(data []byte) ([][]byte, error) {
	enc, err := reedsolomon.New(rse.dataShards, rse.parityShards)

	if err != nil {
		return nil, err
	}

	shards, err := enc.Split(data)

	if err != nil {
		return nil, err
	}

	err = enc.Encode(shards)

	if err != nil {
		return nil, err
	}

	return shards, nil
}

func (rse ReedSolomonEncoder) Reconstruct(shards [][]byte) ([]byte, error) {
	enc, err := reedsolomon.New(rse.dataShards, rse.parityShards)

	if err != nil {
		return nil, err
	}

	err = enc.Reconstruct(shards)

	if err != nil {
		return nil, err
	}

	return bytes.Join(shards[:rse.dataShards], nil), nil
}
