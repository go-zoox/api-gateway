package service

import (
	"testing"
)

func TestServiceHost(t *testing.T) {
	// 测试场景1: 指定端口
	t.Run("SpecifiedPort", func(t *testing.T) {
		service := &Service{
			Name: "example.com",
			Port: 8080,
		}

		host := service.Host()
		expected := "example.com:8080"

		if host != expected {
			t.Errorf("生成的主机地址不匹配，期望: %s, 实际: %s", expected, host)
		}
	})

	// 测试场景2: 未指定端口（应默认为80）
	t.Run("DefaultPort", func(t *testing.T) {
		service := &Service{
			Name: "example.com",
			Port: 0, // 未指定端口
		}

		host := service.Host()
		expected := "example.com:80"

		if host != expected {
			t.Errorf("生成的主机地址不匹配，期望: %s, 实际: %s", expected, host)
		}

		// 验证服务对象的端口是否已更新为默认值
		if service.Port != 80 {
			t.Errorf("服务对象的端口未更新为默认值，期望: 80, 实际: %d", service.Port)
		}
	})

	// 测试场景3: 特殊端口号
	t.Run("SpecialPorts", func(t *testing.T) {
		// HTTPS 默认端口
		service := &Service{
			Name: "secure.example.com",
			Port: 443,
		}

		host := service.Host()
		expected := "secure.example.com:443"

		if host != expected {
			t.Errorf("生成的主机地址不匹配，期望: %s, 实际: %s", expected, host)
		}
	})

	// 测试场景4: 带有IP地址的服务名
	t.Run("IPAddressAsName", func(t *testing.T) {
		service := &Service{
			Name: "192.168.1.1",
			Port: 8080,
		}

		host := service.Host()
		expected := "192.168.1.1:8080"

		if host != expected {
			t.Errorf("生成的主机地址不匹配，期望: %s, 实际: %s", expected, host)
		}
	})

	// 测试场景5: 本地主机
	t.Run("LocalhostAsName", func(t *testing.T) {
		service := &Service{
			Name: "localhost",
			Port: 3000,
		}

		host := service.Host()
		expected := "localhost:3000"

		if host != expected {
			t.Errorf("生成的主机地址不匹配，期望: %s, 实际: %s", expected, host)
		}
	})
}
