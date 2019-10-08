package main

import (
	"crypto/sha256"

	"golang.org/x/crypto/blake2b"
)

// Sha256 exports single sha256 hash function for trie
var Sha256 = func(data ...[]byte) []byte {
	hasher := sha256.New()
	for i := 0; i < len(data); i++ {
		hasher.Write(data[i])
	}
	return hasher.Sum(nil)
}

// Blake2b exports Blake2b hash function for trie
var Blake2b = func(data ...[]byte) []byte {
	hasher, _ := blake2b.New(32, nil)
	for i := 0; i < len(data); i++ {
		hasher.Write(data[i])
	}
	return hasher.Sum(nil)
}
