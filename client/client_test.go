package client_test

import (
	"github.com/dashevo/aergo-trie-db/client"
	"testing"
)

func Test_main(t *testing.T) {
	c := client.CreateClient("127.0.0.1:10000")

	client.CreateTrie(c)
}
