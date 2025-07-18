package registeration

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"time"
)

type ConsulRegister struct {
	client    *api.Client
	serviceID string
}

func NewConsulRegister(consulAddr string) (*ConsulRegister, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ConsulRegister{client: client}, nil
}

func (r *ConsulRegister) Register(serviceName string, port int) error {
	r.serviceID = fmt.Sprintf("%s-%d", serviceName, time.Now().Unix())

	reg := &api.AgentServiceRegistration{
		ID:      r.serviceID,
		Name:    serviceName,
		Port:    port,
		Address: getLocalIP(),
		Check: &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d", getLocalIP(), port),
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	return r.client.Agent().ServiceRegister(reg)
}

func (r *ConsulRegister) Deregister() error {
	if r.serviceID == "" {
		return nil
	}
	return r.client.Agent().ServiceDeregister(r.serviceID)
}

func getLocalIP() string {
	// 简化实现，生产环境需完善
	return "127.0.0.1"
}
