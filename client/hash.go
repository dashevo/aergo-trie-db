package main

import (
	"crypto/sha256"
)

// hash256 function with 256 bit outputs.
func hash256(m []byte) []byte {
	h := sha256.New()
	h.Write(m)
	return h.Sum(nil)
}
