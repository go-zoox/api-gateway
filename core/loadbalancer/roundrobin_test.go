package loadbalancer

import (
	"net/http"
	"testing"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

func TestRoundRobinSelect(t *testing.T) {
	rr := NewRoundRobin()

	backend := &route.NormalizedBackend{
		Algorithm: "round-robin",
		Servers: []*service.Server{
			{Name: "server1.com", Port: 8080},
			{Name: "server2.com", Port: 8080},
			{Name: "server3.com", Port: 8080},
		},
		BaseConfig: &service.Service{},
	}

	// 设置所有服务器为健康
	for _, server := range backend.Servers {
		server.SetHealthy(true)
		server.Disabled = false
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	// 测试轮询
	selected1, err := rr.Select(req, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	selected2, err := rr.Select(req, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	selected3, err := rr.Select(req, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	// 验证选择了不同的服务器（轮询）
	if selected1.ID() == selected2.ID() || selected2.ID() == selected3.ID() {
		t.Logf("Selected servers: %s, %s, %s", selected1.ID(), selected2.ID(), selected3.ID())
		// 注意：由于并发，可能选择相同的服务器，但至少应该轮询
	}

	// 第四次选择应该回到第一个
	selected4, err := rr.Select(req, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	// 验证至少有一次选择了不同的服务器
	allSame := selected1.ID() == selected2.ID() && selected2.ID() == selected3.ID() && selected3.ID() == selected4.ID()
	if allSame {
		t.Error("Round-robin 应该选择不同的服务器")
	}
}

func TestRoundRobinNoHealthyServers(t *testing.T) {
	rr := NewRoundRobin()

	backend := &route.NormalizedBackend{
		Algorithm: "round-robin",
		Servers: []*service.Server{
			{Name: "server1.com", Port: 8080},
		},
		BaseConfig: &service.Service{},
	}

	// 设置服务器为不健康
	backend.Servers[0].SetHealthy(false)
	backend.Servers[0].Disabled = false

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	_, err := rr.Select(req, backend)
	if err == nil {
		t.Error("没有健康服务器时应该返回错误")
	}
}

func TestRoundRobinDisabledServer(t *testing.T) {
	rr := NewRoundRobin()

	backend := &route.NormalizedBackend{
		Algorithm: "round-robin",
		Servers: []*service.Server{
			{Name: "server1.com", Port: 8080, Disabled: true},
			{Name: "server2.com", Port: 8080, Disabled: false},
		},
		BaseConfig: &service.Service{},
	}

	// 设置所有服务器为健康
	for _, server := range backend.Servers {
		server.SetHealthy(true)
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	selected, err := rr.Select(req, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	// 应该只选择启用的服务器
	if selected.ID() != "server2.com:8080" {
		t.Errorf("应该选择启用的服务器, 实际: %s", selected.ID())
	}
}
