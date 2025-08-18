package service_hub

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	SERVICE_ROOT_PATH = "/radic/index" // prefix path in etcd key-value store
)

type ServiceHub struct {
	client             *etcdv3.Client
	heartbeatFrequency int64 // heartbeat frequency in seconds
	watched            sync.Map
	loadBalancer       LoadBalancer
}

var (
	serviceHub *ServiceHub
	hubOnce    sync.Once // single instance only
)

// Get the singleton instance of ServiceHub
func GetServiceHub(etcdServers []string, heartbeatFrequency int64) *ServiceHub {
	hubOnce.Do(func() {
		if client, err := etcdv3.New(
			etcdv3.Config{
				Endpoints:   etcdServers,
				DialTimeout: 3 * time.Second,
			},
		); err != nil {
			logger.Log.Fatalf("Failed to connect to etcd server: %v", err)
		} else {
			serviceHub = &ServiceHub{
				client:             client,
				heartbeatFrequency: heartbeatFrequency,
				loadBalancer:       &RoundRobin{},
			}
		}
	})
	
	return serviceHub
}

func (hub *ServiceHub) Regist(service string, endpoint string, leaseID etcdv3.LeaseID) (etcdv3.LeaseID, error) {
	ctx := context.Background()
	if leaseID <= 0 {
		// create a lease, valid for heartbeatFrequency seconds
		if lease, err := hub.client.Grant(ctx, hub.heartbeatFrequency); err != nil {
			logger.Log.Printf("Failed to create lease：%v", err)
			return 0, err
		} else {
			key := strings.TrimRight(SERVICE_ROOT_PATH, "/") + "/" + service + "/" + endpoint
			// put key with lease, and keep the lease alive forever
			if _, err = hub.client.Put(ctx, key, "", etcdv3.WithLease(lease.ID)); err != nil {
				logger.Log.Printf("Failed to write endpoint %s to service %s：%v", endpoint, service, err)
				return lease.ID, err
			} else {
				return lease.ID, nil
			}
		}
	} else {
		// extend the lease
		if _, err := hub.client.KeepAliveOnce(ctx, leaseID); err == rpctypes.ErrLeaseNotFound {
			return hub.Regist(service, endpoint, 0) // lease not found, re-register
		} else if err != nil {
			logger.Log.Printf("Failed to extend lease:%v", err)
			return 0, err
		} else {
			return leaseID, nil
		}
	}
}

func (hub *ServiceHub) UnRegist(service string, endpoint string) error {
	ctx := context.Background()
	key := strings.TrimRight(SERVICE_ROOT_PATH, "/") + "/" + service + "/" + endpoint
	if _, err := hub.client.Delete(ctx, key); err != nil {
		logger.Log.Printf("Failed to delete endpoint %s of service %s: %v", endpoint, service, err)
		return err
	} else {
		logger.Log.Printf("Delete endpoint %s of service %s", endpoint, service)
		return nil
	}
}

func (hub *ServiceHub) GetServiceEndpoints(service string) []string {
	ctx := context.Background()
	prefix := strings.TrimRight(SERVICE_ROOT_PATH, "/") + "/" + service + "/"
	if resp, err := hub.client.Get(ctx, prefix, etcdv3.WithPrefix()); err != nil { // get all keys with the prefix
		logger.Log.Printf("Failed to get the endpoint of service %s: %v", service, err)
		return nil
	} else {
		endpoints := make([]string, 0, len(resp.Kvs))
		for _, kv := range resp.Kvs {
			path := strings.Split(string(kv.Key), "/")
			// fmt.Println(string(kv.Key), path[len(path)-1])
			endpoints = append(endpoints, path[len(path)-1])
		}
		logger.Log.Printf("Update server of service %s -- %v\n", service, endpoints)
		return endpoints
	}
}

func (hub *ServiceHub) GetServiceEndpoint(service string) string {
	return hub.loadBalancer.Take(hub.GetServiceEndpoints(service))
}

func (hub *ServiceHub) Close() {
	hub.client.Close()
}
