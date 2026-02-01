package loadbalancer

import (
	"net/http"
	"testing"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

func TestIPHashSelect(t *testing.T) {
	ih := NewIPHash()

	backend := &route.NormalizedBackend{
		Algorithm: "ip-hash",
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

	// 测试相同 IP 总是选择相同的服务器
	req1, _ := http.NewRequest("GET", "http://example.com", nil)
	req1.RemoteAddr = "192.168.1.1:12345"

	selected1, err := ih.Select(req1, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	req2, _ := http.NewRequest("GET", "http://example.com", nil)
	req2.RemoteAddr = "192.168.1.1:54321"

	selected2, err := ih.Select(req2, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	// 相同 IP 应该选择相同的服务器
	if selected1.ID() != selected2.ID() {
		t.Errorf("相同 IP 应该选择相同的服务器, 实际: %s vs %s", selected1.ID(), selected2.ID())
	}
}

func TestIPHashDifferentIPs(t *testing.T) {
	ih := NewIPHash()

	backend := &route.NormalizedBackend{
		Algorithm: "ip-hash",
		Servers: []*service.Server{
			{Name: "server1.com", Port: 8080},
			{Name: "server2.com", Port: 8080},
		},
		BaseConfig: &service.Service{},
	}

	// 设置所有服务器为健康
	for _, server := range backend.Servers {
		server.SetHealthy(true)
		server.Disabled = false
	}

	// 测试不同 IP 可能选择不同的服务器
	req1, _ := http.NewRequest("GET", "http://example.com", nil)
	req1.RemoteAddr = "192.168.1.1:12345"

	req2, _ := http.NewRequest("GET", "http://example.com", nil)
	req2.RemoteAddr = "192.168.1.2:12345"

	selected1, _ := ih.Select(req1, backend)
	selected2, _ := ih.Select(req2, backend)

	// 不同 IP 可能选择不同的服务器（取决于哈希值）
	t.Logf("IP 192.168.1.1 选择: %s", selected1.ID())
	t.Logf("IP 192.168.1.2 选择: %s", selected2.ID())
}

func TestIPHashXForwardedFor(t *testing.T) {
	ih := NewIPHash()

	backend := &route.NormalizedBackend{
		Algorithm: "ip-hash",
		Servers: []*service.Server{
			{Name: "server1.com", Port: 8080},
			{Name: "server2.com", Port: 8080},
		},
		BaseConfig: &service.Service{},
	}

	// 设置所有服务器为健康
	for _, server := range backend.Servers {
		server.SetHealthy(true)
		server.Disabled = false
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.RemoteAddr = "192.168.1.1:12345"

	selected, err := ih.Select(req, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	// 应该基于 X-Forwarded-For 的 IP 选择
	t.Logf("使用 X-Forwarded-For 选择: %s", selected.ID())
}

func TestIPHashNoHealthyServers(t *testing.T) {
	ih := NewIPHash()

	backend := &route.NormalizedBackend{
		Algorithm: "ip-hash",
		Servers: []*service.Server{
			{Name: "server1.com", Port: 8080},
		},
		BaseConfig: &service.Service{},
	}

	// 设置服务器为不健康
	backend.Servers[0].SetHealthy(false)
	backend.Servers[0].Disabled = false

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	_, err := ih.Select(req, backend)
	if err == nil {
		t.Error("没有健康服务器时应该返回错误")
	}
}
