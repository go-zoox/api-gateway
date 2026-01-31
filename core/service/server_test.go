package service

import (
	"testing"
)

func TestServerID(t *testing.T) {
	server := &Server{
		Name: "example.com",
		Port: 8080,
	}

	expected := "example.com:8080"
	actual := server.ID()

	if actual != expected {
		t.Errorf("Server ID 不匹配，期望: %s, 实际: %s", expected, actual)
	}
}

func TestServerHost(t *testing.T) {
	t.Run("SpecifiedPort", func(t *testing.T) {
		server := &Server{
			Name: "example.com",
			Port: 8080,
		}

		expected := "example.com:8080"
		actual := server.Host()

		if actual != expected {
			t.Errorf("Server Host 不匹配，期望: %s, 实际: %s", expected, actual)
		}
	})

	t.Run("DefaultPort", func(t *testing.T) {
		server := &Server{
			Name: "example.com",
			Port: 0,
		}

		expected := "example.com:80"
		actual := server.Host()

		if actual != expected {
			t.Errorf("Server Host 不匹配，期望: %s, 实际: %s", expected, actual)
		}
	})
}

func TestServerTarget(t *testing.T) {
	t.Run("WithProtocol", func(t *testing.T) {
		server := &Server{
			Name:     "example.com",
			Port:     8080,
			Protocol: "https",
		}

		expected := "https://example.com:8080"
		actual := server.Target()

		if actual != expected {
			t.Errorf("Server Target 不匹配，期望: %s, 实际: %s", expected, actual)
		}
	})

	t.Run("DefaultProtocol", func(t *testing.T) {
		server := &Server{
			Name: "example.com",
			Port: 8080,
		}

		expected := "http://example.com:8080"
		actual := server.Target()

		if actual != expected {
			t.Errorf("Server Target 不匹配，期望: %s, 实际: %s", expected, actual)
		}
	})
}

func TestServerHealthStatus(t *testing.T) {
	server := &Server{}

	// 默认不健康（需要显式设置或通过健康检查设置）
	// 但在 Normalize() 中会被设置为健康
	server.SetHealthy(true)
	if !server.IsHealthy() {
		t.Error("设置后应该是健康的")
	}

	// 设置为不健康
	server.SetHealthy(false)
	if server.IsHealthy() {
		t.Error("设置后应该是不健康的")
	}

	// 重新设置为健康
	server.SetHealthy(true)
	if !server.IsHealthy() {
		t.Error("重新设置后应该是健康的")
	}
}

func TestServerGetEffectiveConfig(t *testing.T) {
	base := &Service{
		Protocol: "http",
		Request: Request{
			Headers: map[string]string{
				"X-Global": "global-value",
			},
		},
		Response: Response{
			Headers: map[string]string{
				"X-Global-Response": "global-response",
			},
		},
	}

	server := &Server{
		Protocol: "https",
		Request: &Request{
			Headers: map[string]string{
				"X-Server": "server-value",
			},
		},
	}

	effective := server.GetEffectiveConfig(base)

	// 检查协议覆盖
	if effective.Protocol != "https" {
		t.Errorf("协议应该被覆盖，期望: https, 实际: %s", effective.Protocol)
	}

	// 检查请求头合并
	if effective.Request.Headers["X-Global"] != "global-value" {
		t.Error("全局请求头应该保留")
	}
	if effective.Request.Headers["X-Server"] != "server-value" {
		t.Error("服务器请求头应该存在")
	}
}

func TestServiceIsMultiServer(t *testing.T) {
	t.Run("SingleServer", func(t *testing.T) {
		service := &Service{
			Name: "example.com",
			Port: 8080,
		}

		if service.IsMultiServer() {
			t.Error("单服务模式应该返回 false")
		}
	})

	t.Run("MultiServer", func(t *testing.T) {
		service := &Service{
			Algorithm: "round-robin",
			Servers: []Server{
				{Name: "server1.com", Port: 8080},
				{Name: "server2.com", Port: 8080},
			},
		}

		if !service.IsMultiServer() {
			t.Error("多服务模式应该返回 true")
		}
	})
}
