package discovery

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"math/rand"
)

type ConsulDiscoverer struct {
	client *api.Client
}

func NewConsulDiscoverer(consulAddr string) (*ConsulDiscoverer, error) {
	client, err := api.NewClient(&api.Config{
		Address: consulAddr,
	})
	if err != nil {
		return nil, err
	}
	return &ConsulDiscoverer{client: client}, nil
}

func (d *ConsulDiscoverer) Discover(serviceName string) (string, error) {
	entries, _, err := d.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("consul query failed: %v", err)
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("no healthy instances available")
	}

	// Go 1.20+ 推荐方式（无需显式Seed）
	selected := entries[rand.Intn(len(entries))] // 自动使用随机源
	return fmt.Sprintf("%s:%d", selected.Service.Address, selected.Service.Port), nil
}
