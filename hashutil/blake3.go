package hashutil

import (
	"github.com/zeebo/blake3"
)

func Blake3(data []byte) []byte {
	// Create a new BLAKE3 hasher
	hasher := blake3.New()

	// Write the data to the hasher
	hasher.Write([]byte(data))

	// Compute the hash
	hash := hasher.Sum(nil)

	return hash
}
