package indexing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	index_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/index"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/net"
	service_hub "github.com/m1i3k0e7/distributed-search-engine/internal/service_hub"
	index "github.com/m1i3k0e7/distributed-search-engine/api/proto/index"
)

const (
	INDEX_SERVICE = "index_service"
)

// IndexWorker, a gRPC service worker for indexing.
type IndexServiceWorker struct {
	index.UnimplementedIndexServiceServer
	Indexer *Indexer // kvdb + inverted index
	hub      *service_hub.ServiceHub
	selfAddr string
}

func (service *IndexServiceWorker) Init(DocNumEstimate int, dbtype int, DataDir string) error {
	service.Indexer = new(Indexer)
	return service.Indexer.Init(DocNumEstimate, dbtype, DataDir)
}

func (service *IndexServiceWorker) Regist(etcdServers []string, servicePort int) error {
	// Register the service in etcd if etcdServers is not empty
	if len(etcdServers) > 0 {
		if servicePort <= 1024 {
			return fmt.Errorf("invalid listen port %d, should more than 1024", servicePort)
		}

		selfLocalIp, err := net.GetLocalIP()
		if err != nil {
			panic(err)
		}
		
		// selfLocalIp = "127.0.0.1" // when testing locally, use loopback address
		service.selfAddr = selfLocalIp + ":" + strconv.Itoa(servicePort)
		var heartBeat int64 = 3                      // heartbeat interval in seconds
		hub := service_hub.GetServiceHub(etcdServers, heartBeat)
		leaseId, err := hub.Regist(INDEX_SERVICE, service.selfAddr, 0)
		if err != nil {
			panic(err)
		}

		service.hub = hub
		// heartbeat goroutine
		go func() {
			for {
				hub.Regist(INDEX_SERVICE, service.selfAddr, leaseId)
				time.Sleep(time.Duration(heartBeat)*time.Second - 100*time.Millisecond)
			}
		}()
	}

	return nil
}

func (service *IndexServiceWorker) Close() error {
	if service.hub != nil {
		service.hub.UnRegist(INDEX_SERVICE, service.selfAddr)
	}
	return service.Indexer.Close()
}

func (service *IndexServiceWorker) DeleteDoc(ctx context.Context, docId *index_proto.DocId) (*index_proto.AffectedCount, error) {
	return &index_proto.AffectedCount{Count: int32(service.Indexer.DeleteDoc(docId.DocId))}, nil
}

func (service *IndexServiceWorker) AddDoc(ctx context.Context, doc *search_proto.Document) (*index_proto.AffectedCount, error) {
	n, err := service.Indexer.AddDoc(*doc)
	return &index_proto.AffectedCount{Count: int32(n)}, err
}

func (service *IndexServiceWorker) Search(ctx context.Context, request *index_proto.SearchRequest) (*index_proto.SearchResult, error) {
	result := service.Indexer.Search(request.Query, request.OnFlag, request.OffFlag, request.OrFlags)
	return &index_proto.SearchResult{Results: result}, nil
}

func (service *IndexServiceWorker) Count(ctx context.Context, request *index_proto.CountRequest) (*index_proto.AffectedCount, error) {
	return &index_proto.AffectedCount{Count: int32(service.Indexer.Count())}, nil
}
