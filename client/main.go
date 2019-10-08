package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/dashevo/universe-tree-db/universe"
	grpc "google.golang.org/grpc"
)

// Note: This gRPC client is for example purposes and is only intended to
// demonstrate example usage of the server. It is not meant to be used for
// production use.

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "missing subcommand: list, create, drop, sync, update, commit, get, stash, revert, merkleproof, merkleproofcompressed, merkleproofr, merkleproofcompressedr")
		os.Exit(1)
	}

	// Get connection addr from env or use default
	strConnect := srvConnAddr()

	conn, err := grpc.Dial(strConnect, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to backend: %v\n", err)
		os.Exit(1)
	}
	client := universe.NewUniTreeDBClient(conn)

	switch cmd := flag.Arg(0); cmd {
	case "list":
		err = listTrees(context.Background(), client)
	case "create":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: create <name>")
			os.Exit(1)
		}
		err = createTree(context.Background(), client, flag.Arg(1))
	case "drop":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: drop <name>")
			os.Exit(1)
		}
		err = dropTree(context.Background(), client, flag.Arg(1))
	case "sync":
		err = syncMeta(context.Background(), client)
	case "update":
		if flag.NArg() < 4 {
			fmt.Fprintln(os.Stderr, "usage: update <treename> <key-str> <val-str> [1=atomic]")
			os.Exit(1)
		}
		var atomicUpdate bool
		if flag.NArg() >= 4 && flag.Arg(4) == "1" {
			atomicUpdate = true
		}
		err = update(context.Background(), client, flag.Arg(1), flag.Arg(2), flag.Arg(3), atomicUpdate)
	case "commit":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: commit <treename>")
			os.Exit(1)
		}
		err = commit(context.Background(), client, flag.Arg(1))
	case "get":
		if flag.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "usage: get <treename> <key-str>")
			os.Exit(1)
		}
		err = get(context.Background(), client, flag.Arg(1), flag.Arg(2))
	case "stash":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: stash <treename> [1=rollbackCache]")
			os.Exit(1)
		}
		var rollbackCache bool
		if flag.NArg() >= 3 && flag.Arg(2) == "1" {
			rollbackCache = true
		}
		err = stash(context.Background(), client, flag.Arg(1), rollbackCache)
	case "revert":
		if flag.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "usage: revert <treename> <oldroot-hex>")
			os.Exit(1)
		}
		err = revert(context.Background(), client, flag.Arg(1), flag.Arg(2))
	case "merkleproof":
		if flag.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "usage: merkleproof <treename> <key-str>")
			os.Exit(1)
		}
		err = merkleproof(context.Background(), client, flag.Arg(1), flag.Arg(2))
	case "merkleproofcompressed":
		if flag.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "usage: merkleproofcompressed <treename> <key-str>")
			os.Exit(1)
		}
		err = merkleproofcompressed(context.Background(), client, flag.Arg(1), flag.Arg(2))
	case "merkleproofr":
		if flag.NArg() < 4 {
			fmt.Fprintln(os.Stderr, "usage: merkleproofr <treename> <key-str> <root-hex>")
			os.Exit(1)
		}
		err = merkleproofr(context.Background(), client, flag.Arg(1), flag.Arg(2), flag.Arg(3))
	case "merkleproofcompressedr":
		if flag.NArg() < 4 {
			fmt.Fprintln(os.Stderr, "usage: merkleproofcompressedr <treename> <key-str> <root-hex>")
			os.Exit(1)
		}
		err = merkleproofcompressedr(context.Background(), client, flag.Arg(1), flag.Arg(2), flag.Arg(3))
	default:
		err = fmt.Errorf("unknown subcommand %s", cmd)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func listTrees(ctx context.Context, client universe.UniTreeDBClient) error {
	resp, err := client.ListTrees(ctx, &universe.Void{})
	if err != nil {
		return err
	}

	fmt.Printf("Got %d trees\n", len(resp.GetList()))
	for _, t := range resp.GetList() {
		fmt.Printf("\tname: %s, root: %x, height: %d, loadDbCounter: %d, loadCacheCounter: %d, cacheHeightLimit: %d\n", t.Name, t.Root, t.TrieHeight, t.LoadDbCounter, t.LoadCacheCounter, t.CacheHeightLimit)
	}

	return nil
}

func createTree(ctx context.Context, client universe.UniTreeDBClient, name string) error {
	resp, err := client.CreateTree(ctx, &universe.CreateTreeRequest{
		Name:             name,
		CacheHeightLimit: 0,
	})
	if err != nil {
		return err
	}

	wasCreated := resp.GetCreated()
	fmt.Println("tree was created:", wasCreated)

	return nil
}

func dropTree(ctx context.Context, client universe.UniTreeDBClient, name string) error {
	resp, err := client.DropTree(ctx, &universe.DropTreeRequest{Name: name})
	if err != nil {
		return err
	}

	wasDeleted := resp.GetDeleted()
	fmt.Println("tree was deleted:", wasDeleted)

	return nil
}

func syncMeta(ctx context.Context, client universe.UniTreeDBClient) error {
	_, err := client.SyncMeta(ctx, &universe.Void{})
	if err != nil {
		return err
	}

	fmt.Println("meta synced!")
	return nil
}

func update(ctx context.Context, client universe.UniTreeDBClient, treeName, key, value string, atomic bool) error {
	hashK := hash256([]byte(key))
	hashV := hash256([]byte(value))

	req := &universe.UpdateRequest{
		TreeName:      treeName,
		KeyValuePairs: []*universe.KeyValuePair{{Key: hashK, Value: hashV}},
	}

	type UpdateFunc func(context.Context, *universe.UpdateRequest, ...grpc.CallOption) (*universe.UpdateReply, error)
	var updateFunc UpdateFunc

	updateFunc = client.Update
	if atomic {
		updateFunc = client.AtomicUpdate
	}
	resp, err := updateFunc(ctx, req)
	if err != nil {
		return err
	}

	root := resp.GetRoot()
	fmt.Printf("new root: [%x]\n", root)

	return nil
}

func get(ctx context.Context, client universe.UniTreeDBClient, treeName string, key string) error {
	hashK := hash256([]byte(key))
	resp, err := client.Get(ctx, &universe.GetRequest{
		TreeName: treeName,
		Key:      hashK,
	})
	if err != nil {
		return err
	}

	val := resp.GetValue()
	fmt.Printf("val: %x\n", val)

	return nil
}

func commit(ctx context.Context, client universe.UniTreeDBClient, treeName string) error {
	_, err := client.Commit(ctx, &universe.CommitRequest{TreeName: treeName})
	if err != nil {
		return err
	}

	fmt.Printf("trie %v committed\n", treeName)
	return nil
}

func stash(ctx context.Context, client universe.UniTreeDBClient, treeName string, rollbackCache bool) error {
	_, err := client.Stash(ctx, &universe.StashRequest{
		TreeName:      treeName,
		RollbackCache: rollbackCache,
	})
	if err != nil {
		return err
	}

	fmt.Printf("trie %v stashed w/rollbackCache=%v\n", treeName, rollbackCache)
	return nil
}

func revert(ctx context.Context, client universe.UniTreeDBClient, treeName string, toOldRoot string) error {
	b, err := hex.DecodeString(toOldRoot)
	if err != nil {
		return err
	}

	_, err = client.Revert(ctx, &universe.RevertRequest{
		TreeName:  treeName,
		ToOldRoot: b,
	})
	if err != nil {
		return err
	}

	fmt.Printf("trie %v reverted to old root [%x]\n", treeName, b)
	return nil
}

func merkleproof(ctx context.Context, client universe.UniTreeDBClient, treeName string, key string) error {
	hashK := hash256([]byte(key))
	resp, err := client.MerkleProof(ctx, &universe.GetRequest{
		TreeName: treeName,
		Key:      hashK,
	})
	if err != nil {
		return err
	}

	mp := resp.MerkleProof

	fmt.Printf("trie %v got merkle proof, included: %v, proofKey: [%x], proofVal: [%x]\n", treeName, mp.Included, mp.ProofKey, mp.ProofValue)
	for i, b := range mp.AuditPath {
		fmt.Printf("audit path[%d] = [%x]\n", i, b)
	}

	return nil
}

func merkleproofcompressed(ctx context.Context, client universe.UniTreeDBClient, treeName string, key string) error {
	hashK := hash256([]byte(key))
	resp, err := client.MerkleProofCompressed(ctx, &universe.GetRequest{
		TreeName: treeName,
		Key:      hashK,
	})
	if err != nil {
		return err
	}

	mp := resp.MerkleProof

	fmt.Printf("trie %v got merkle compressed proof, bitmap [%x], height: %d, included: %v, proofKey: [%x], proofVal: [%x]\n", treeName, mp.Bitmap, mp.Height, mp.Included, mp.ProofKey, mp.ProofValue)
	for i, b := range mp.AuditPath {
		fmt.Printf("audit path[%d] = [%x]\n", i, b)
	}

	return nil
}

func merkleproofr(ctx context.Context, client universe.UniTreeDBClient, treeName string, key string, root string) error {
	hashK := hash256([]byte(key))
	r, err := hex.DecodeString(root)
	if err != nil {
		return err
	}
	resp, err := client.MerkleProofR(ctx, &universe.MerkleProofRRequest{
		TreeName: treeName,
		Key:      hashK,
		Root:     r,
	})
	if err != nil {
		return err
	}

	mp := resp.MerkleProof

	fmt.Printf("trie %v got merkle proof, included: %v, proofKey: [%x], proofVal: [%x]\n", treeName, mp.Included, mp.ProofKey, mp.ProofValue)
	for i, b := range mp.AuditPath {
		fmt.Printf("audit path[%d] = [%x]\n", i, b)
	}

	return nil
}

func merkleproofcompressedr(ctx context.Context, client universe.UniTreeDBClient, treeName string, key string, root string) error {
	hashK := hash256([]byte(key))
	r, err := hex.DecodeString(root)
	if err != nil {
		return err
	}
	resp, err := client.MerkleProofCompressedR(ctx, &universe.MerkleProofRRequest{
		TreeName: treeName,
		Key:      hashK,
		Root:     r,
	})
	if err != nil {
		return err
	}

	mp := resp.MerkleProof

	fmt.Printf("trie %v got merkle compressed proof, bitmap [%x], height: %d, included: %v, proofKey: [%x], proofVal: [%x]\n", treeName, mp.Bitmap, mp.Height, mp.Included, mp.ProofKey, mp.ProofValue)
	for i, b := range mp.AuditPath {
		fmt.Printf("audit path[%d] = [%x]\n", i, b)
	}

	return nil
}

// srvConnAddr returns the IP / port to connect to
func srvConnAddr() string {
	addr := os.Getenv("UNIDB_CONNECT")
	if len(addr) == 0 {
		addr = "127.0.0.1:9002"
	}
	return addr
}
