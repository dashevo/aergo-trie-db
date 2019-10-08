package main

import (
	"bytes"
	"encoding/gob"

	"github.com/aergoio/aergo/pkg/trie"
	"github.com/dashevo/universe-tree-db/universe"
)

// TreeInfo has a trie pointer and meta info about the trie.
type TreeInfo struct {
	trie *trie.Trie
	universe.TreeInfo
}

// Serialize returns the serialized bytes for a TreeInfo.
//
// Note that it's only serializing the underlying universe.TreeInfo and not the
// trie pointer.
func (ti TreeInfo) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(ti.TreeInfo)
	return buf.Bytes()
}

// TreeInfoFromBytes creates a TreeInfo from serialized bytes.
func TreeInfoFromBytes(data []byte) TreeInfo {
	var info TreeInfo
	dec := gob.NewDecoder(bytes.NewReader(data))
	dec.Decode(&info.TreeInfo)
	return info
}

// Equal checks if two TreeInfos are equal
func (ti TreeInfo) Equal(other TreeInfo) bool {
	return (ti.TreeInfo.Name == other.TreeInfo.Name &&
		bytes.Equal(ti.TreeInfo.Root, other.TreeInfo.Root))
}
