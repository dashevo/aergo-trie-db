package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aergoio/aergo/pkg/trie"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
	"log"
	"net"

	pb "github.com/dashevo/aergo-trie-db/proto"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	port     = flag.Int("port", 10000, "The server port")
)

type aergoTrieDBServer struct {
	pb.AergoTrieDBServer
	//savedFeatures []*pb.Feature // read-only after initialized

	//mu         sync.Mutex // protects routeNotes
	//routeNotes map[string][]*pb.RouteNote
}

func (s *aergoTrieDBServer) CreateTrie(ctx context.Context, trieDefinition *pb.CreateTrieRequest) (*pb.CreateTrieResponse, error) {
	smt := trie.NewTrie(nil, func(data ...[]byte) []byte { return data[0] }, nil)

	str := fmt.Sprintf("%#v", smt)

	return &pb.CreateTrieResponse{Trie: str}, nil
}

func newServer() *aergoTrieDBServer {
	s := &aergoTrieDBServer{}
	return s
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	if *tls {
		if *certFile == "" {
			*certFile = testdata.Path("server1.pem")
		}

		if *keyFile == "" {
			*keyFile = testdata.Path("server1.key")
		}

		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}

		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterAergoTrieDBServer(grpcServer, newServer())

	log.Printf("Server listens on 127.0.0.1:%d\n", *port)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
