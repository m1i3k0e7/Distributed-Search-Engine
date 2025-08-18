package indexing

import (
	context "context"
	"fmt"
	"sync"
	"sync/atomic"

	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	"github.com/m1i3k0e7/distributed-search-engine/api/proto/index"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	service_hub "github.com/m1i3k0e7/distributed-search-engine/internal/service_hub"
)

type Sentinel struct {
	hub      service_hub.IServiceHub // get the set of IndexServiceWorker endpoints from ServiceHub or ServiceHubProxy
	connPool sync.Map    // connection pool, key: endpoint, value: *grpc.ClientConn
}

func NewSentinel(etcdServers []string) *Sentinel {
	return &Sentinel{
		// hub: GetServiceHub(etcdServers, 10)
		hub:      service_hub.GetServiceHubProxy(etcdServers, 10, 100), // via Service Hub Proxy
		connPool: sync.Map{},
	}
}

func (sentinel *Sentinel) GetGrpcConn(endpoint string) *grpc.ClientConn {
	if v, exists := sentinel.connPool.Load(endpoint); exists {
		conn := v.(*grpc.ClientConn)
		// delete the connection if it is not in a connecting or ready state
		if !(conn.GetState() == connectivity.Connecting || conn.GetState() == connectivity.Ready) {
			logger.Log.Printf("connection status to endpoint %s is %s", endpoint, conn.GetState())
			conn.Close()
			sentinel.connPool.Delete(endpoint)
		} else {
			return conn // return the existing connection if it is still valid
		}
	}
	// ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	// defer cancel()
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithBlock(),
	)
	if err != nil {
		logger.Log.Printf("dial %s failed: %s", endpoint, err)
		return nil
	}

	logger.Log.Printf("connect to grpc server %s", endpoint)
	sentinel.connPool.Store(endpoint, conn)

	return conn
}

func (sentinel *Sentinel) AddDoc(doc search_proto.Document) (int, error) {
	endpoint := sentinel.hub.GetServiceEndpoint(INDEX_SERVICE) // select one IndexServiceWorker endpoint according to the load balancing policy
	if len(endpoint) == 0 {
		return 0, fmt.Errorf("there is no alive index worker")
	}

	conn := sentinel.GetGrpcConn(endpoint)
	if conn == nil {
		return 0, fmt.Errorf("connect to worker %s failed", endpoint)
	}

	client := index.NewIndexServiceClient(conn)
	affected, err := client.AddDoc(context.Background(), &doc)
	if err != nil {
		return 0, err
	}

	logger.Log.Printf("add %d doc to worker %s", affected.Count, endpoint)

	return int(affected.Count), nil
}

func (sentinel *Sentinel) UpdateDoc(doc search_proto.Document) (int, error) {
	sentinel.DeleteDoc(doc.Id)
	return sentinel.AddDoc(doc)
}

func (sentinel *Sentinel) DeleteDoc(docId string) int {
	endpoints := sentinel.hub.GetServiceEndpoints(INDEX_SERVICE)
	if len(endpoints) == 0 {
		return 0
	}

	var n int32
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	for _, endpoint := range endpoints {
		go func(endpoint string) { // parallel delete
			defer wg.Done()
			conn := sentinel.GetGrpcConn(endpoint)
			if conn != nil {
				client := index.NewIndexServiceClient(conn)
				affected, err := client.DeleteDoc(context.Background(), &index.DocId{DocId: docId})
				if err != nil {
					logger.Log.Printf("delete doc %s from worker %s failed: %s", docId, endpoint, err)
				} else {
					if affected.Count > 0 {
						atomic.AddInt32(&n, affected.Count)
						logger.Log.Printf("delete %d from worker %s", affected.Count, endpoint)
					}
				}
			}
		}(endpoint)
	}
	wg.Wait()
	
	return int(atomic.LoadInt32(&n))
}

func (sentinel *Sentinel) Search(query *search_proto.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*search_proto.Document {
	endpoints := sentinel.hub.GetServiceEndpoints(INDEX_SERVICE)
	if len(endpoints) == 0 {
		return nil
	}

	docs := make([]*search_proto.Document, 0, 1000)
	resultCh := make(chan *search_proto.Document, 1000)
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			defer wg.Done()
			conn := sentinel.GetGrpcConn(endpoint)
			if conn != nil {
				client := index.NewIndexServiceClient(conn)
				result, err := client.Search(context.Background(), &index.SearchRequest{Query: query, OnFlag: onFlag, OffFlag: offFlag, OrFlags: orFlags})
				if err != nil {
					logger.Log.Printf("search from cluster failed: %s", err)
				} else {
					if len(result.Results) > 0 {
						logger.Log.Printf("search %d doc from worker %s", len(result.Results), endpoint)
						for _, doc := range result.Results {
							resultCh <- doc
						}
					}
				}
			}
		}(endpoint)
	}

	receiveFinish := make(chan struct{})
	go func() {
		for {
			doc, ok := <-resultCh // synchronously receive documents from all workers
			if !ok {
				break // 2. if the channel is closed, exit the loop
			}
			docs = append(docs, doc)
		}
		receiveFinish <- struct{}{} // 3. signal that all documents have been received
	}()
	wg.Wait()
	close(resultCh) // 1. close the channel to signal that no more documents will be sent
	<-receiveFinish // 4. wait for the receive goroutine to finish

	return docs
}

func (sentinel *Sentinel) Count() int {
	var n int32
	endpoints := sentinel.hub.GetServiceEndpoints(INDEX_SERVICE)
	if len(endpoints) == 0 {
		return 0
	}

	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			defer wg.Done()
			conn := sentinel.GetGrpcConn(endpoint)
			if conn != nil {
				client := index.NewIndexServiceClient(conn)
				affected, err := client.Count(context.Background(), new(index.CountRequest))
				if err != nil {
					logger.Log.Printf("get doc count from worker %s failed: %s", endpoint, err)
				} else {
					if affected.Count > 0 {
						atomic.AddInt32(&n, affected.Count)
						logger.Log.Printf("worker %s have %d documents", endpoint, affected.Count)
					}
				}
			}
		}(endpoint)
	}
	wg.Wait()

	return int(n)
}

func (sentinel *Sentinel) Close() (err error) {
	sentinel.connPool.Range(func(key, value any) bool {
		conn := value.(*grpc.ClientConn)
		err = conn.Close()
		return true
	})
	sentinel.hub.Close()
	
	return
}
