package discovery

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

const (
	schema = "etcd"
)

type Resolver struct {
	schema      string
	ETCDAddress []string
	DialTimeout int

	closeChan chan struct{}
	watchChan clientv3.WatchChan
	cli       clientv3.Client

	keyPrefix          string
	serviceAddressList []resolver.Address
	clientConnection   resolver.ClientConn
}

func NewResolver(ETCDAddress []string) *Resolver {
	return &Resolver{
		schema:      schema,
		ETCDAddress: ETCDAddress,
		DialTimeout: 3,
	}
}

func (r *Resolver) Scheme() string {
	return r.schema
}

func (r *Resolver) Build(target resolver.Target, clientConn resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.clientConnection = clientConn
	r.keyPrefix = BuildPrefix(Server{
		Name:    target.Endpoint(),
		Version: target.URL.Host,
	})
	if _, err := r.start(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Resolver) ResolveNow(options resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {
	r.closeChan <- struct{}{}
}

func (r *Resolver) start() (chan<- struct{}, error) {
	var err error
	// r.clientConnection, err = clientv3.New(clientv3.Config{
	// 	Endpoints:   r.ETCDAddress,
	// 	DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	resolver.Register(r)

	r.closeChan = make(chan struct{})

	if err = r.sync(); err != nil {
		return nil, err
	}

	go r.watch()

	return r.closeChan, nil
}

func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)
	r.watchChan = r.cli.Watch(context.Background(), r.keyPrefix, clientv3.WithPrefix())

	for {
		select {
		case <-r.closeChan:
			return
		case res, ok := <-r.watchChan:
			if ok {
				r.update(res.Events)
			}
		case <-ticker.C:
			if err := r.sync(); err != nil {
				fmt.Println("sync failed", err)
			}
		}
	}
}

func (r *Resolver) update(events []*clientv3.Event) {
	for _, ev := range events {
		var info Server
		var err error

		switch ev.Type {
		case clientv3.EventTypePut:
			info, err = ParseValue(ev.Kv.Value)
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
			if !Exist(r.serviceAddressList, addr) {
				r.serviceAddressList = append(r.serviceAddressList, addr)
				if err := r.clientConnection.UpdateState(resolver.State{Addresses: r.serviceAddressList}); err != nil {
					return
				}
			}
		case clientv3.EventTypeDelete:
			info, err = SplitPath(string(ev.Kv.Key))
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr}
			if s, ok := Remove(r.serviceAddressList, addr); ok {
				r.serviceAddressList = s
				if err := r.clientConnection.UpdateState(resolver.State{Addresses: r.serviceAddressList}); err != nil {
					return
				}
			}
		}
	}
}

func (r *Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := r.cli.Get(ctx, r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	r.serviceAddressList = []resolver.Address{}

	for _, v := range res.Kvs {
		info, err := ParseValue(v.Value)
		if err != nil {
			continue
		}
		addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
		r.serviceAddressList = append(r.serviceAddressList, addr)
	}
	if err := r.clientConnection.UpdateState(resolver.State{Addresses: r.serviceAddressList}); err != nil {
		return err
	}

	return nil
}
