package client_test

import (
	"github.com/dashevo/aergo-trie-db/client"
	"testing"
)

func Test_main(t *testing.T) {
	c := client.CreateClient()

	client.CreateTrie(c)
}
