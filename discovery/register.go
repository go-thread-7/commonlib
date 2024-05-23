package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Register struct {
	ETCDAddress []string
	DialTimeout int

	closeChan     chan struct{}
	leasesID      clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse

	serviceInfo Server
	serviceTTL  int64
	cli         *clientv3.Client
}

func NewRegister(ETCDAddress []string) *Register {
	return &Register{
		ETCDAddress: ETCDAddress,
		DialTimeout: 3,
	}
}

func (r *Register) Register(serviceInfo Server, ttl int64) (chan<- struct{}, error) {
	var err error
	if strings.Split(serviceInfo.Addr, ":")[0] == "" {
		return nil, errors.New("invalid ip address")
	}

	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.ETCDAddress,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	}); err != nil {
		return nil, err
	}

	r.serviceInfo = serviceInfo
	r.serviceTTL = ttl

	if err = r.register(); err != nil {
		return nil, err
	}

	r.closeChan = make(chan struct{})

	go r.keepAlive()

	return r.closeChan, nil
}

func (r *Register) register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()

	leaseResp, err := r.cli.Grant(ctx, r.serviceTTL)
	if err != nil {
		return err
	}

	r.leasesID = leaseResp.ID

	if r.keepAliveChan, err = r.cli.KeepAlive(context.Background(), r.leasesID); err != nil {
		return err
	}

	data, err := json.Marshal(r.serviceInfo)
	if err != nil {
		return err
	}

	_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.serviceInfo), string(data), clientv3.WithLease(r.leasesID))

	return err
}

func (r *Register) Stop() {
	//r.closeCh <- struct{}{}
	if err := r.unregister(); err != nil {
		fmt.Println("unregister failed, error: ", err)
	}

	if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil {
		fmt.Println("revoke failed, error: ", err)
	}
}

func (r *Register) unregister() error {
	_, err := r.cli.Delete(context.Background(), BuildRegisterPath(r.serviceInfo))
	return err
}

func (r *Register) keepAlive() {
	ticker := time.NewTicker(time.Duration(r.serviceTTL) * time.Second)

	for {
		select {
		case res := <-r.keepAliveChan:
			if res == nil {
				if err := r.register(); err != nil {
					fmt.Println("register failed, error: ", err)
				}
			}
		case <-ticker.C:
			if r.keepAliveChan == nil {
				if err := r.register(); err != nil {
					fmt.Println("register failed, error: ", err)
				}
			}
		}
	}
}

func (r *Register) UpdateHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		weightstr := req.URL.Query().Get("weight")
		weight, err := strconv.Atoi(weightstr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var update = func() error {
			r.serviceInfo.Weight = int64(weight)
			data, err := json.Marshal(r.serviceInfo)
			if err != nil {
				return err
			}

			_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.serviceInfo), string(data), clientv3.WithLease(r.leasesID))
			return err
		}

		if err := update(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write([]byte("update server weight success"))
	})
}

func (r *Register) GetServerInfo() (Server, error) {
	resp, err := r.cli.Get(context.Background(), BuildRegisterPath(r.serviceInfo))
	if err != nil {
		return r.serviceInfo, err
	}

	server := Server{}
	if resp.Count >= 1 {
		if err := json.Unmarshal(resp.Kvs[0].Value, &server); err != nil {
			return server, err
		}
	}

	return server, err
}
