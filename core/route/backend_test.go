package route

import (
	"testing"

	"github.com/go-zoox/api-gateway/core/service"
)

func TestBackendNormalize(t *testing.T) {
	t.Run("SingleServerMode", func(t *testing.T) {
		backend := &Backend{
			Service: service.Service{
				Name:     "example.com",
				Port:     8080,
				Protocol: "https",
			},
		}

		normalized := backend.Normalize()
		if normalized == nil {
			t.Fatal("Normalize() 不应该返回 nil")
		}

		if normalized.Algorithm != "round-robin" {
			t.Errorf("默认算法应该是 round-robin, 实际: %s", normalized.Algorithm)
		}

		if len(normalized.Servers) != 1 {
			t.Fatalf("单服务模式应该有 1 个服务器, 实际: %d", len(normalized.Servers))
		}

		server := normalized.Servers[0]
		if server.Name != "example.com" {
			t.Errorf("服务器名称不匹配, 期望: example.com, 实际: %s", server.Name)
		}
		if server.Port != 8080 {
			t.Errorf("服务器端口不匹配, 期望: 8080, 实际: %d", server.Port)
		}
		if server.Protocol != "https" {
			t.Errorf("服务器协议不匹配, 期望: https, 实际: %s", server.Protocol)
		}
	})

	t.Run("MultiServerMode", func(t *testing.T) {
		backend := &Backend{
			Service: service.Service{
				Algorithm: "weighted",
				Servers: []service.Server{
					{Name: "server1.com", Port: 8080, Weight: 1},
					{Name: "server2.com", Port: 8080, Weight: 2},
					{Name: "server3.com", Port: 8080, Weight: 1},
				},
			},
		}

		normalized := backend.Normalize()
		if normalized == nil {
			t.Fatal("Normalize() 不应该返回 nil")
		}

		if normalized.Algorithm != "weighted" {
			t.Errorf("算法应该是 weighted, 实际: %s", normalized.Algorithm)
		}

		if len(normalized.Servers) != 3 {
			t.Fatalf("应该有 3 个服务器, 实际: %d", len(normalized.Servers))
		}
	})

	t.Run("EmptyBackend", func(t *testing.T) {
		backend := &Backend{}

		normalized := backend.Normalize()
		if normalized != nil {
			t.Error("空 Backend 的 Normalize() 应该返回 nil")
		}
	})
}

func TestBackendIsMultiServer(t *testing.T) {
	t.Run("SingleServer", func(t *testing.T) {
		backend := &Backend{
			Service: service.Service{
				Name: "example.com",
				Port: 8080,
			},
		}

		if backend.IsMultiServer() {
			t.Error("单服务模式应该返回 false")
		}
	})

	t.Run("MultiServer", func(t *testing.T) {
		backend := &Backend{
			Service: service.Service{
				Servers: []service.Server{
					{Name: "server1.com", Port: 8080},
				},
			},
		}

		if !backend.IsMultiServer() {
			t.Error("多服务模式应该返回 true")
		}
	})
}

func TestBackendGetService(t *testing.T) {
	t.Run("SingleServerMode", func(t *testing.T) {
		backend := &Backend{
			Service: service.Service{
				Name:     "example.com",
				Port:     8080,
				Protocol: "https",
			},
		}

		service := backend.GetService()
		if service == nil {
			t.Fatal("GetService() 不应该返回 nil")
		}

		if service.Name != "example.com" {
			t.Errorf("服务名称不匹配, 期望: example.com, 实际: %s", service.Name)
		}
	})

	t.Run("MultiServerMode", func(t *testing.T) {
		backend := &Backend{
			Service: service.Service{
				Servers: []service.Server{
					{Name: "server1.com", Port: 8080},
					{Name: "server2.com", Port: 8080},
				},
			},
		}

		service := backend.GetService()
		if service == nil {
			t.Fatal("GetService() 不应该返回 nil")
		}

		// 应该返回第一个服务器的有效配置
		if service.Name != "server1.com" {
			t.Errorf("应该返回第一个服务器, 期望: server1.com, 实际: %s", service.Name)
		}
	})
}
