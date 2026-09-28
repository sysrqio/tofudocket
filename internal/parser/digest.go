package parser

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Digest returns a sha256:... digest over plan bytes.
func SHA256Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
