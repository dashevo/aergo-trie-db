package client

import (
	"context"
	"flag"
	pb "github.com/dashevo/aergo-trie-db/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
	"log"
	"time"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func CreateTrie(client pb.AergoTrieDBClient) {
	log.Println("Creating trie...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.CreateTrie(ctx, &pb.CreateTrieRequest{})
	if err != nil {
		log.Fatalf("%v.CreateTrie(_) = _, %v: ", client, err)
	}

	log.Println(response)
}

func CreateClient() pb.AergoTrieDBClient {
	flag.Parse()

	var opts []grpc.DialOption

	if *tls {
		if *caFile == "" {
			*caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	log.Printf("Connect to %s\n", *serverAddr)

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()

	return pb.NewAergoTrieDBClient(conn)
}
