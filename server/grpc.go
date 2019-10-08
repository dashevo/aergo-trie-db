package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aergoio/aergo/pkg/trie"
	"github.com/dashevo/universe-tree-db/universe"
)

func (s *universeTrieServer) ListTrees(ctx context.Context, req *universe.Void) (*universe.ListTreesReply, error) {
	s.Lock()
	defer s.Unlock()

	var resp universe.ListTreesReply

	listTreeInfo := make([]*universe.TreeInfo, len(s.trieInfo))
	i := 0
	for name, ti := range s.trieInfo {
		listTreeInfo[i] = &universe.TreeInfo{
			Name:             name,
			Root:             ti.trie.Root,
			TrieHeight:       uint32(ti.trie.TrieHeight),
			LoadDbCounter:    uint32(ti.trie.LoadDbCounter),
			LoadCacheCounter: uint32(ti.trie.LoadCacheCounter),
			CacheHeightLimit: uint32(ti.trie.CacheHeightLimit),
		}
		i++
	}

	resp.List = listTreeInfo
	return &resp, nil
}

func (s *universeTrieServer) CreateTree(ctx context.Context, req *universe.CreateTreeRequest) (*universe.CreateTreeReply, error) {
	var resp universe.CreateTreeReply

	treeName := req.GetName()

	s.Lock()
	defer s.Unlock()

	_, ok := s.trieInfo[treeName]
	if ok {
		log.Printf("CreateTree: tree [%v] already exists", treeName)
	} else {
		log.Printf("CreateTree: creating tree [%v]", treeName)
		t := trie.NewTrie(nil, Sha256, s.aergoDB)
		t.CacheHeightLimit = int(req.GetCacheHeightLimit())
		ti := TreeInfo{
			trie: t,
			TreeInfo: universe.TreeInfo{
				Name:             treeName,
				CacheHeightLimit: uint32(t.CacheHeightLimit),
			},
		}
		s.trieInfo[treeName] = ti
		resp.Created = true
		s.syncMeta()
	}

	return &resp, nil
}

func (s *universeTrieServer) DropTree(ctx context.Context, req *universe.DropTreeRequest) (*universe.DropTreeReply, error) {
	var resp universe.DropTreeReply

	treeName := req.GetName()

	s.Lock()
	defer s.Unlock()

	_, ok := s.trieInfo[treeName]
	if !ok {
		log.Printf("DropTree: Tree [%v] not found", treeName)
	} else {
		log.Printf("DropTree: Deleting Tree [%v]", treeName)
		delete(s.trieInfo, treeName)
		resp.Deleted = true
		s.syncMeta()
	}

	return &resp, nil
}

func (s *universeTrieServer) Update(ctx context.Context, req *universe.UpdateRequest) (*universe.UpdateReply, error) {
	var resp universe.UpdateReply

	treeName := req.GetTreeName()

	s.Lock()
	defer s.Unlock()

	val, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}

	pairs := req.KeyValuePairs
	keys := make([][]byte, len(pairs))
	values := make([][]byte, len(pairs))
	for i, pair := range pairs {
		keys[i] = pair.GetKey()
		values[i] = pair.GetValue()
	}

	trie := val.trie
	log.Printf("Update: trie.Root BEFORE update: [%x]", trie.Root)
	root, err := trie.Update(keys, values)
	if err != nil {
		return nil, err
	}
	log.Printf("Update: trie.Root AFTER  update: [%x]", trie.Root)
	resp.Root = root

	err = s.syncMeta()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *universeTrieServer) AtomicUpdate(ctx context.Context, req *universe.UpdateRequest) (*universe.UpdateReply, error) {
	var resp universe.UpdateReply

	treeName := req.GetTreeName()

	s.Lock()
	defer s.Unlock()

	val, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}

	pairs := req.KeyValuePairs
	keys := make([][]byte, len(pairs))
	values := make([][]byte, len(pairs))
	for i, pair := range pairs {
		keys[i] = pair.GetKey()
		values[i] = pair.GetValue()
	}

	trie := val.trie
	log.Printf("AtomicUpdate: trie.Root BEFORE update: [%x]", trie.Root)
	root, err := trie.AtomicUpdate(keys, values)
	if err != nil {
		return nil, err
	}
	log.Printf("AtomicUpdate: trie.Root AFTER  update: [%x]", trie.Root)
	resp.Root = root

	err = s.syncMeta()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *universeTrieServer) SyncMeta(ctx context.Context, in *universe.Void) (*universe.Void, error) {
	s.Lock()
	defer s.Unlock()

	err := s.syncMeta()
	return &universe.Void{}, err
}

func (s *universeTrieServer) Commit(ctx context.Context, req *universe.CommitRequest) (*universe.Void, error) {
	treeName := req.GetTreeName()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	trie := ti.trie
	err := trie.Commit()
	if err != nil {
		return nil, err
	}

	log.Printf("Commit: trie [%v] committed", treeName)
	return &universe.Void{}, nil
}

func (s *universeTrieServer) Get(ctx context.Context, req *universe.GetRequest) (*universe.GetReply, error) {
	var resp universe.GetReply

	treeName := req.GetTreeName()
	key := req.GetKey()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	log.Printf("Get: key [%x]", key)
	trie := ti.trie
	val, err := trie.Get(key)
	if err != nil {
		return nil, err
	}

	resp.Value = val
	log.Printf("Get: got value [%x]", val)
	return &resp, nil
}

func (s *universeTrieServer) Stash(ctx context.Context, req *universe.StashRequest) (*universe.Void, error) {
	treeName := req.GetTreeName()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	trie := ti.trie
	err := trie.Stash(req.GetRollbackCache())
	if err != nil {
		return nil, err
	}

	log.Printf("Stash: trie [%v] stashed", treeName)
	return &universe.Void{}, nil
}

func (s *universeTrieServer) Revert(ctx context.Context, req *universe.RevertRequest) (*universe.Void, error) {
	treeName := req.GetTreeName()
	toOldRoot := req.GetToOldRoot()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	trie := ti.trie
	err := trie.Revert(toOldRoot)
	if err != nil {
		return nil, err
	}

	log.Printf("Revert: trie [%v] reverted to old root [%x]", treeName, toOldRoot)
	return &universe.Void{}, nil
}

func (s *universeTrieServer) MerkleProof(ctx context.Context, req *universe.GetRequest) (*universe.MerkleProofReply, error) {
	treeName := req.GetTreeName()
	key := req.GetKey()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	trie := ti.trie
	auditPath, included, proofKey, proofValue, err := trie.MerkleProof(key)
	if err != nil {
		return nil, err
	}

	log.Printf("MerkleProof: trie [%v] key [%x] auditPath: %v, included: %v, proofKey [%x], proofValue [%x]\n", treeName, key, auditPath, included, proofKey, proofValue)
	return &universe.MerkleProofReply{
		MerkleProof: &universe.MerkleProof{
			AuditPath:  auditPath,
			Included:   included,
			ProofKey:   proofKey,
			ProofValue: proofValue,
		},
	}, nil
}

func (s *universeTrieServer) MerkleProofCompressed(ctx context.Context, req *universe.GetRequest) (*universe.MerkleProofCompressedReply, error) {
	treeName := req.GetTreeName()
	key := req.GetKey()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	trie := ti.trie
	bitmap, auditPath, height, included, proofKey, proofValue, err := trie.MerkleProofCompressed(key)
	if err != nil {
		return nil, err
	}

	log.Printf("MerkleProofCompressed: trie [%v] key [%x] bitmap: [%x], auditPath: %v, height: %d, included: %v, proofKey [%x], proofValue [%x]\n", treeName, key, bitmap, auditPath, height, included, proofKey, proofValue)
	return &universe.MerkleProofCompressedReply{
		MerkleProof: &universe.MerkleProofCompressed{
			Bitmap:     bitmap,
			AuditPath:  auditPath,
			Height:     uint32(height),
			Included:   included,
			ProofKey:   proofKey,
			ProofValue: proofValue,
		},
	}, nil
}

func (s *universeTrieServer) MerkleProofR(ctx context.Context, req *universe.MerkleProofRRequest) (*universe.MerkleProofReply, error) {
	treeName := req.GetTreeName()
	key := req.GetKey()
	root := req.GetRoot()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	trie := ti.trie
	auditPath, included, proofKey, proofValue, err := trie.MerkleProofR(key, root)
	if err != nil {
		return nil, err
	}

	log.Printf("MerkleProofR: trie [%v] key [%x] auditPath: %v, included: %v, proofKey [%x], proofValue [%x]\n", treeName, key, auditPath, included, proofKey, proofValue)
	return &universe.MerkleProofReply{
		MerkleProof: &universe.MerkleProof{
			AuditPath:  auditPath,
			Included:   included,
			ProofKey:   proofKey,
			ProofValue: proofValue,
		},
	}, nil
}

func (s *universeTrieServer) MerkleProofCompressedR(ctx context.Context, req *universe.MerkleProofRRequest) (*universe.MerkleProofCompressedReply, error) {
	treeName := req.GetTreeName()
	key := req.GetKey()
	root := req.GetRoot()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	trie := ti.trie
	bitmap, auditPath, height, included, proofKey, proofValue, err := trie.MerkleProofCompressedR(key, root)
	if err != nil {
		return nil, err
	}

	log.Printf("MerkleProofCompressedR: trie [%v] key [%x] bitmap: [%x], auditPath: %v, height: %d, included: %v, proofKey [%x], proofValue [%x]\n", treeName, key, bitmap, auditPath, height, included, proofKey, proofValue)
	return &universe.MerkleProofCompressedReply{
		MerkleProof: &universe.MerkleProofCompressed{
			Bitmap:     bitmap,
			AuditPath:  auditPath,
			Height:     uint32(height),
			Included:   included,
			ProofKey:   proofKey,
			ProofValue: proofValue,
		},
	}, nil
}

func (s *universeTrieServer) VerifyInclusion(ctx context.Context, req *universe.VerifyInclusionRequest) (*universe.VerifyInclusionReply, error) {
	treeName := req.GetTreeName()
	mp := req.GetMerkleProof()

	auditPath := mp.GetAuditPath()
	key := mp.GetProofKey()
	value := mp.GetProofValue()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	included := ti.trie.VerifyInclusion(auditPath, key, value)

	log.Printf("VerifyInclusion: trie [%v] auditPath: %v, key [%x], value: [%x], included: %v\n", treeName, auditPath, key, value, included)
	return &universe.VerifyInclusionReply{
		Included: included,
	}, nil
}

func (s *universeTrieServer) VerifyNonInclusion(ctx context.Context, req *universe.VerifyNonInclusionRequest) (*universe.VerifyInclusionReply, error) {
	treeName := req.GetTreeName()
	mp := req.GetMerkleProof()
	proofKey := req.GetProofKey()

	auditPath := mp.GetAuditPath()
	key := mp.GetProofKey()
	value := mp.GetProofValue()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	included := ti.trie.VerifyNonInclusion(auditPath, key, value, proofKey)

	log.Printf("VerifyNonInclusion: trie [%v] auditPath: %v, key [%x], value: [%x], proofKey: [%x], included: %v\n", treeName, auditPath, key, value, proofKey, included)
	return &universe.VerifyInclusionReply{
		Included: included,
	}, nil
}

func (s *universeTrieServer) VerifyInclusionC(ctx context.Context, req *universe.VerifyInclusionCRequest) (*universe.VerifyInclusionReply, error) {
	treeName := req.GetTreeName()
	mp := req.GetMerkleProof()

	bitmap := mp.GetBitmap()
	key := mp.GetProofKey()
	value := mp.GetProofValue()
	auditPath := mp.GetAuditPath()
	length := mp.GetHeight()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	included := ti.trie.VerifyInclusionC(bitmap, key, value, auditPath, int(length))

	log.Printf("VerifyInclusionC: trie [%v] bitmap: [%x], key: [%x], value: [%x], auditPath: %v, length: %d, included: %v\n", treeName, bitmap, key, value, auditPath, length, included)
	return &universe.VerifyInclusionReply{
		Included: included,
	}, nil
}

func (s *universeTrieServer) VerifyNonInclusionC(ctx context.Context, req *universe.VerifyNonInclusionCRequest) (*universe.VerifyInclusionReply, error) {
	treeName := req.GetTreeName()
	mp := req.GetMerkleProof()
	proofKey := req.GetProofKey()

	bitmap := mp.GetBitmap()
	key := mp.GetProofKey()
	value := mp.GetProofValue()
	auditPath := mp.GetAuditPath()
	length := mp.GetHeight()

	s.Lock()
	defer s.Unlock()

	ti, ok := s.trieInfo[treeName]
	if !ok {
		return nil, fmt.Errorf("tree [%v] not found", treeName)
	}
	included := ti.trie.VerifyNonInclusionC(auditPath, int(length), bitmap, key, value, proofKey)

	log.Printf("VerifyNonInclusionC: trie [%v] auditPath: %v, length: %d, bitmap: [%x], key: [%x], value: [%x], proofKey: [%x], included: %v\n", treeName, auditPath, length, bitmap, key, value, proofKey, included)
	return &universe.VerifyInclusionReply{
		Included: included,
	}, nil
}
