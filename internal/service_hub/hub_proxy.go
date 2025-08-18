package service_hub

import (
	context "context"
	"strings"
	"sync"
	"time"

	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/time/rate"
)

type IServiceHub interface {
	Regist(service string, endpoint string, leaseID etcdv3.LeaseID) (etcdv3.LeaseID, error)
	UnRegist(service string, endpoint string) error
	GetServiceEndpoints(service string) []string
	GetServiceEndpoint(service string) string
	Close()
}

// Proxy for ServiceHub, add local cache and rate limiter
type HubProxy struct {
	*ServiceHub
	endpointCache sync.Map
	limiter       *rate.Limiter
}

var (
	proxy     *HubProxy
	proxyOnce sync.Once
)

func GetServiceHubProxy(etcdServers []string, heartbeatFrequency int64, qps int) *HubProxy {
	proxyOnce.Do(func() {
		serviceHub := GetServiceHub(etcdServers, heartbeatFrequency)
		if serviceHub != nil {
			proxy = &HubProxy{
				ServiceHub:    serviceHub,
				endpointCache: sync.Map{},
				limiter:       rate.NewLimiter(rate.Every(time.Duration(1e9/qps)*time.Nanosecond), qps),
			}
		}
	})
	
	return proxy
}

func (proxy *HubProxy) watchEndpointsOfService(service string) {
	if _, exists := proxy.watched.LoadOrStore(service, true); exists {
		return
	}

	ctx := context.Background()
	prefix := strings.TrimRight(SERVICE_ROOT_PATH, "/") + "/" + service + "/"
	ch := proxy.client.Watch(ctx, prefix, etcdv3.WithPrefix())
	logger.Log.Printf("监听服务%s的节点变化", service)
	go func() {
		for response := range ch {
			for _, event := range response.Events {
				logger.Log.Printf("etcd event type %s", event.Type) // PUT/DELETE
				path := strings.Split(string(event.Kv.Key), "/")
				if len(path) > 2 {
					service := path[len(path)-2]
					// sync local cache
					endpoints := proxy.ServiceHub.GetServiceEndpoints(service)
					if len(endpoints) > 0 {
						proxy.endpointCache.Store(service, endpoints)
					} else {
						proxy.endpointCache.Delete(service)
					}
				}
			}
		}
	}()
}


func (proxy *HubProxy) GetServiceEndpoints(service string) []string {
	// ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	// defer cancel()
	// proxy.limiter.Wait(ctx)

	if !proxy.limiter.Allow() { // return nil if rate limit exceeded
		return nil
	}

	proxy.watchEndpointsOfService(service) // listen for changes, ensure cache is updated
	if endpoints, exists := proxy.endpointCache.Load(service); exists {
		return endpoints.([]string)
	} else {
		endpoints := proxy.ServiceHub.GetServiceEndpoints(service)
		if len(endpoints) > 0 {
			proxy.endpointCache.Store(service, endpoints) // query from etcd and cache it
		}
		return endpoints
	}
}
