package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	
	"github.com/m1i3k0e7/distributed-search-engine/internal/indexing"
	"github.com/m1i3k0e7/distributed-search-engine/api/proto/index"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
	"google.golang.org/grpc"
)

var service *indexing.IndexServiceWorker //IndexWorker is a gRPC server that provides indexing services

func GrpcIndexerInit() {
	// listen on a specific port
	lis, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(*port))
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	service = new(indexing.IndexServiceWorker)
	// Initialize the indexer service
	service.Init(50000, dbType, *dbPath+"_part"+strconv.Itoa(*workerIndex))
	if *rebuildIndex {
		logger.Log.Printf("totalWorkers=%d, workerIndex=%d", *totalWorkers, *workerIndex)
		indexing.BuildIndexFromFile(csvFile, service.Indexer, *totalWorkers, *workerIndex) // rebuild index from csv file
	} else {
		service.Indexer.LoadFromIndexFile() // load index from file
	}

	// Register the service with the gRPC server
	index.RegisterIndexServiceServer(server, service)
	// Start the gRPC server
	fmt.Printf("start grpc server on port %d\n", *port)
	// Register service to etcd for service discovery
	service.Regist(etcdServers, *port)
	err = server.Serve(lis)
	if err != nil {
		service.Close()
		fmt.Printf("start grpc server on port %d failed: %s\n", *port, err)
	}
}

func GrpcIndexerTeardown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	service.Close() // close the service after receiving the signal
	os.Exit(0)
}

func GrpcIndexerMain() {
	go GrpcIndexerTeardown()
	GrpcIndexerInit()
}
