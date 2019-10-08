package main

import (
	"fmt"

	"github.com/aergoio/aergo/pkg/trie"
)

func main() {
	smt := trie.NewTrie(nil, func(data ...[]byte) []byte { return data[0] }, nil)
	fmt.Printf("smt: %+v\n", smt)
}
