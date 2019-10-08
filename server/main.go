package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aergoio/aergo-lib/db"
	"github.com/dashevo/universe-tree-db/universe"
	"github.com/dgraph-io/badger"
	grpc "google.golang.org/grpc"
)

func main() {
	if !dbDirValid(baseDBDir()) {
		log.Fatal("invalid dir for db")
	}

	uniTreeSrv := newUniverseTrieServer()

	// aergo db
	aergoDB := db.NewDB(db.BadgerImpl, aergoDBSubDir())
	defer aergoDB.Close()
	uniTreeSrv.aergoDB = aergoDB

	// meta db
	metaDB, err := badger.Open(badger.DefaultOptions(metaDBSubDir()))
	if err != nil {
		log.Fatal(err)
	}
	defer metaDB.Close()
	uniTreeSrv.metaDB = metaDB

	// handle interrupts gracefully
	var handler CloseHandler
	handler.RegisterShutdownHandler(uniTreeSrv)

	// Load tries from meta
	err = uniTreeSrv.init()
	if err != nil {
		log.Fatal(err)
	}
	defer uniTreeSrv.GracefulStop()

	strListen := srvListenAddr()
	srv := grpc.NewServer()
	universe.RegisterUniTreeDBServer(srv, uniTreeSrv)
	handler.RegisterShutdownHandler(srv)
	handler.Init()

	l, err := net.Listen("tcp", strListen)
	if err != nil {
		log.Fatalf("could not listen on %s: %v", strListen, err)
	}
	err = srv.Serve(l)
	if err != nil {
		log.Fatal(err)
	}
}

func dbDirValid(dir string) bool {
	stat, err := os.Stat(dir)
	if err != nil {
		return false
	}
	if !stat.IsDir() {
		return false
	}
	return true
}

// srvListenAddr returns the IP / port to listen on
func srvListenAddr() string {
	listen := os.Getenv("UNIDB_LISTEN")
	if len(listen) == 0 {
		listen = "127.0.0.1:9002"
	}
	return listen
}

func baseDBDir() string {
	dbDirVar := "UNIDB_DIR"
	dbDir := os.Getenv(dbDirVar)
	if len(dbDir) == 0 {
		log.Printf("db dir %s not set in environment", dbDirVar)
	}
	return dbDir
}
func aergoDBSubDir() string {
	return baseDBDir() + "/aergo"
}
func metaDBSubDir() string {
	return baseDBDir() + "/meta"
}

// Graceful is an interface services implement to shutdown gracefully.
type Graceful interface {
	GracefulStop()
}

// CloseHandler creates a 'listener' on a new goroutine which
// will notify the program if it receives an interrupt from the OS. We then
// handle this by calling our cleanup procedure and exiting the program.
type CloseHandler struct {
	services []Graceful
	sync.Mutex
}

// Init starts listening for an interrupt event.
func (ch *CloseHandler) Init() {
	ch.Lock()
	defer ch.Unlock()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Print("Got interrupt, closing!")
		for _, s := range ch.services {
			s.GracefulStop()
		}
	}()
}

// RegisterShutdownHandler registers the service to stop when interrupt
// received. Services are shutdown in the order called.
func (ch *CloseHandler) RegisterShutdownHandler(s Graceful) {
	ch.Lock()
	defer ch.Unlock()
	// add service to list
	ch.services = append(ch.services, s)
}
