package main_test

import (
	"testing"

	main "github.com/dashevo/universe-tree-db/server"
	"github.com/dashevo/universe-tree-db/universe"
)

func TestTreeInfo(t *testing.T) {
	ti := makeTreeInfo()

	b := ti.Serialize()
	ti2 := main.TreeInfoFromBytes(b)
	if !ti.Equal(ti2) {
		t.Errorf("got %v, expected %v", ti, ti2)
	}
}

func makeTreeInfo() main.TreeInfo {
	return main.TreeInfo{
		TreeInfo: universe.TreeInfo{
			Root: []byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			Name:             "documents",
			TrieHeight:       7,
			CacheHeightLimit: 0,
			LoadDbCounter:    14,
			LoadCacheCounter: 33,
		},
	}
}
