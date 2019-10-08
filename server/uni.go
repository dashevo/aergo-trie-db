package main

import (
	"log"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/pkg/trie"
	"github.com/dgraph-io/badger"
)

// grpc server -- hang DB handle off this
type universeTrieServer struct {
	trieInfo map[string]TreeInfo
	aergoDB  db.DB
	metaDB   *badger.DB
	sync.Mutex
	shutdown bool
}

// newUniverseTrieServer constructs a new *universeTrieServer.
func newUniverseTrieServer() *universeTrieServer {
	return &universeTrieServer{
		trieInfo: make(map[string]TreeInfo),
	}
}

// syncMeta synchronizes the in-memory metadata to the on-disk meta DB.
// Expected to be called w/lock.
func (s *universeTrieServer) syncMeta() error {
	// s.Lock()
	// defer s.Unlock()

	var err error
	trees := make([]string, len(s.trieInfo))
	i := 0
	for treeName, ti := range s.trieInfo {
		// sync metadata from trie before serializing to disk
		ti.TreeInfo.Root = ti.trie.Root
		ti.TreeInfo.TrieHeight = uint32(ti.trie.TrieHeight)
		ti.TreeInfo.LoadDbCounter = uint32(ti.trie.LoadDbCounter)
		ti.TreeInfo.LoadCacheCounter = uint32(ti.trie.LoadCacheCounter)
		ti.TreeInfo.CacheHeightLimit = uint32(ti.trie.CacheHeightLimit)

		err = s.MetaSetTreeInfo(treeName, ti)
		if err != nil {
			log.Printf("SyncMeta: error setting meta for tree [%v]", treeName)
			return err
		}
		trees[i] = treeName
		i++
	}

	err = s.MetaSaveTrees(trees)
	if err != nil {
		log.Printf("SyncMeta: error setting meta tree list")
		return err
	}

	return nil
}

// listTrees returns the names of active/open tries from the trie map
func (s *universeTrieServer) listTrees() []string {
	s.Lock()
	defer s.Unlock()

	trees := make([]string, len(s.trieInfo))
	i := 0
	for name := range s.trieInfo {
		trees[i] = name
		i++
	}
	return trees
}

// loadTries gets a list of tries from the metadb and initializes them as Aergo
// tries in the trie map.
func (s *universeTrieServer) loadTries() error {
	trees, err := s.MetaListTrees()
	if err != nil {
		return err
	}
	log.Printf("loadTries: Got %d tries from meta DB", len(trees))

	s.Lock()
	defer s.Unlock()

	for _, treeName := range trees {
		log.Printf("loadTries: Loading tree %v", treeName)
		ti, err := s.MetaGetTreeInfo(treeName)
		if err != nil {
			return err
		}

		log.Printf("\tRoot=%x, TrieHeight=%d, LoadDbCounter=%d, LoadCacheCounter=%d, CacheHeightLimit=%d", ti.Root, ti.TrieHeight, ti.LoadDbCounter, ti.LoadCacheCounter, ti.CacheHeightLimit)
		t := trie.NewTrie(ti.Root, Sha256, s.aergoDB)
		t.TrieHeight = int(ti.TrieHeight)
		t.LoadDbCounter = int(ti.LoadDbCounter)
		t.LoadCacheCounter = int(ti.LoadCacheCounter)
		t.CacheHeightLimit = int(ti.CacheHeightLimit)

		ti.trie = t
		s.trieInfo[treeName] = ti
	}
	return nil
}

// commitAllTries iterates all active tree names and commits each to the Aergo
// Trie DB. It is intended to be called upon shutdown. The lock should be held
// before calling this.
func (s *universeTrieServer) commitAllTries() {
	for treeName, ti := range s.trieInfo {
		err := ti.trie.Commit()
		if err != nil {
			log.Printf("could not commit trie %v: %v", treeName, err)
		}
	}

	return
}

// init performs any necessary initialization
func (s *universeTrieServer) init() error {
	// Load tries from s.metaDB
	log.Print("Initializing trie server")
	if err := s.loadTries(); err != nil {
		return err
	}
	return nil
}

// GracefulStop shuts down gracefully
func (s *universeTrieServer) GracefulStop() {
	s.Lock()
	defer s.Unlock()

	// already requested, do not run again
	if s.shutdown {
		return
	}

	log.Print("Shutting down gracefully")
	s.commitAllTries()
	log.Print("AergoDB tries committed")

	err := s.syncMeta()
	if err != nil {
		log.Print("Could not sync meta: ", err)
	} else {
		log.Print("Meta synced")
	}

	s.aergoDB.Close()
	log.Print("AergoDB Closed")
	s.shutdown = true
}
