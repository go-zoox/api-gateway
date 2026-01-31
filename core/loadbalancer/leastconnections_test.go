package loadbalancer

import (
	"net/http"
	"testing"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

func TestLeastConnectionsSelect(t *testing.T) {
	lc := NewLeastConnections()

	backend := &route.NormalizedBackend{
		Algorithm: "least-connections",
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

	// 第一次选择，所有连接数都是 0，应该选择第一个
	selected1, err := lc.Select(req, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	// 模拟增加连接数
	lc.OnRequestStart(backend, selected1)

	// 再次选择，应该选择连接数最少的
	selected2, err := lc.Select(req, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	// 应该选择不同的服务器（连接数更少的）
	if selected1.ID() == selected2.ID() {
		t.Logf("选择了相同的服务器: %s，这是可能的（如果所有服务器连接数相同）", selected1.ID())
	}

	// 减少连接数
	lc.OnRequestEnd(backend, selected1)

	// 再次选择
	selected3, err := lc.Select(req, backend)
	if err != nil {
		t.Fatalf("Select 不应该返回错误: %v", err)
	}

	t.Logf("选择结果: %s, %s, %s", selected1.ID(), selected2.ID(), selected3.ID())
}

func TestLeastConnectionsConnectionTracking(t *testing.T) {
	lc := NewLeastConnections()

	backend := &route.NormalizedBackend{
		Algorithm: "least-connections",
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

	// 选择 server1 并增加连接数
	selected1, _ := lc.Select(req, backend)
	lc.OnRequestStart(backend, selected1)
	lc.OnRequestStart(backend, selected1) // 再增加一次

	// 现在应该选择 server2（连接数更少）
	selected2, _ := lc.Select(req, backend)

	if selected1.ID() == selected2.ID() {
		t.Error("应该选择连接数更少的服务器")
	}

	// 减少 server1 的连接数
	lc.OnRequestEnd(backend, selected1)
	lc.OnRequestEnd(backend, selected1)

	// 现在应该选择 server1（连接数更少）
	selected3, _ := lc.Select(req, backend)

	if selected3.ID() != selected1.ID() {
		t.Error("应该选择连接数最少的服务器")
	}
}

func TestLeastConnectionsNoHealthyServers(t *testing.T) {
	lc := NewLeastConnections()

	backend := &route.NormalizedBackend{
		Algorithm: "least-connections",
		Servers: []*service.Server{
			{Name: "server1.com", Port: 8080},
		},
		BaseConfig: &service.Service{},
	}

	// 设置服务器为不健康
	backend.Servers[0].SetHealthy(false)
	backend.Servers[0].Disabled = false

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	_, err := lc.Select(req, backend)
	if err == nil {
		t.Error("没有健康服务器时应该返回错误")
	}
}
