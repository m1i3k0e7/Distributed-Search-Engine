package main

import (
	// "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/m1i3k0e7/distributed-search-engine/internal/handler"
	"github.com/m1i3k0e7/distributed-search-engine/internal/indexing"
	storage "github.com/m1i3k0e7/distributed-search-engine/internal/indexing/trie"
	// "github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
	// "github.com/m1i3k0e7/distributed-search-engine/internal/indexing/trie"
)

func WebServerInit(mode int) {
	switch mode {
	case 1:
		standaloneIndexer := new(indexing.Indexer)                        // Standalone indexer
		if err := standaloneIndexer.Init(50000, dbType, *dbPath); err != nil { // initialize the indexer
			panic(err)
		}

		if *rebuildIndex {
			indexing.BuildIndexFromDir(csvFilesDir, standaloneIndexer, *totalWorkers, *workerIndex) // rebuild index from csv files in the directory
		} else {
			standaloneIndexer.LoadFromIndexFile() // load index from file
		}
		handler.Indexer = standaloneIndexer
		
		standaloneTrieDB, err := storage.NewTrieDB(trieDBPath)
		if err != nil {
			panic(err)
		}
		
		handler.TrieDB = standaloneTrieDB // Set the trie database for the handler
	case 3:
		handler.Indexer = indexing.NewSentinel(etcdServers) // Distributed indexer using sentinel
	default:
		panic("invalid mode")
	}

}

func WebServerTeardown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	handler.Indexer.Close() // close the indexer after receiving the signal
	os.Exit(0)              // exit the program
}

func WebServerMain(mode int) {
	go WebServerTeardown()
	WebServerInit(mode)
}
