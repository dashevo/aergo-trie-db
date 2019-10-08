package main

import (
	"bytes"
	"encoding/gob"

	"github.com/dgraph-io/badger"
)

// keys for metadb
const (
	KeyTrees      = "tries"
	KeyInfoPrefix = "info:"
)

// SerializeStringSlice serializes a list of strings to bytes.
func SerializeStringSlice(slc []string) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(slc)
	return buf.Bytes()
}

// DeserializeStringSlice constructs a list of strings from bytes.
func DeserializeStringSlice(data []byte) []string {
	var slc []string
	dec := gob.NewDecoder(bytes.NewReader(data))
	dec.Decode(&slc)
	return slc
}

// MetaListTrees retrieves the trees list from the meta DB.
func (s *universeTrieServer) MetaListTrees() ([]string, error) {
	var list []string
	err := s.metaDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(KeyTrees))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			return nil
		}
		err = item.Value(func(val []byte) error {
			list = DeserializeStringSlice(val)
			return nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

// MetaGetTreeInfo retrieves a tree info object from the meta DB.
func (s *universeTrieServer) MetaGetTreeInfo(treeName string) (TreeInfo, error) {
	var info TreeInfo
	err := s.metaDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(KeyInfoPrefix + treeName))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			return nil
		}
		err = item.Value(func(val []byte) error {
			info = TreeInfoFromBytes(val)
			return nil
		})
		return err
	})
	if err != nil {
		return TreeInfo{}, err
	}
	return info, nil
}

// MetaSaveTrees updates the list of trees in the meta DB.
func (s *universeTrieServer) MetaSaveTrees(trees []string) error {
	err := s.metaDB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(KeyTrees), SerializeStringSlice(trees))
		return err
	})
	return err
}

// MetaSetTreeInfo saves a tree info object to the meta DB.
func (s *universeTrieServer) MetaSetTreeInfo(treeName string, ti TreeInfo) error {
	err := s.metaDB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(KeyInfoPrefix+treeName), ti.Serialize())
		return err
	})
	return err
}
