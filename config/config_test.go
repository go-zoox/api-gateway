package config

import (
	"encoding/json"
	"testing"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/core-utils/fmt"
	"gopkg.in/yaml.v3"
)

func TestConfigStructure(t *testing.T) {
	// 测试配置结构是否正确
	config := Config{
		Port:    8090,
		BaseURI: "/api",
		Backend: route.Backend{
			Service: route.Service{
				Protocol: "http",
				Name:     "example-service",
				Port:     8080,
			},
		},
		Routes: []route.Route{
			{
				Name: "test-route",
				Path: "/test",
				Backend: route.Backend{
					Service: route.Service{
						Protocol: "http",
						Name:     "test-service",
						Port:     8080,
					},
				},
			},
		},
		Cache: Cache{
			Host:     "localhost",
			Port:     6379,
			Username: "",
			Password: "",
			DB:       0,
			Prefix:   "api-gateway:",
		},
		HealthCheck: HealthCheck{
			Outer: HealthCheckOuter{
				Enable: true,
				Path:   "/health",
				Ok:     true,
			},
			Inner: HealthCheckInner{
				Enable:   true,
				Interval: 60,
				Timeout:  5,
			},
		},
	}

	// 测试 JSON 序列化
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	var jsonConfig Config
	if err := json.Unmarshal(jsonData, &jsonConfig); err != nil {
		t.Fatalf("JSON 反序列化失败: %v", err)
	}

	// 验证 JSON 序列化/反序列化后的字段值
	if jsonConfig.Port != config.Port {
		t.Errorf("Port 不匹配: 期望 %d, 得到 %d", config.Port, jsonConfig.Port)
	}

	if jsonConfig.BaseURI != config.BaseURI {
		t.Errorf("BaseURI 不匹配: 期望 %s, 得到 %s", config.BaseURI, jsonConfig.BaseURI)
	}

	// 测试 YAML 序列化
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("YAML 序列化失败: %v", err)
	}

	var yamlConfig Config
	if err := yaml.Unmarshal(yamlData, &yamlConfig); err != nil {
		t.Fatalf("YAML 反序列化失败: %v", err)
	}

	// 验证 YAML 序列化/反序列化后的字段值
	if yamlConfig.Port != config.Port {
		t.Errorf("Port 不匹配: 期望 %d, 得到 %d", config.Port, yamlConfig.Port)
	}

	if yamlConfig.BaseURI != config.BaseURI {
		t.Errorf("BaseURI 不匹配: 期望 %s, 得到 %s", config.BaseURI, yamlConfig.BaseURI)
	}

	t.Logf("配置结构测试通过: %s", fmt.PrettyJSON(config))
}

func TestSSLConfig(t *testing.T) {
	// 测试 SSL 配置
	ssl := SSL{
		Domain: "example.com",
		Cert: SSLCert{
			Certificate:    "/path/to/cert.pem",
			CertificateKey: "/path/to/key.pem",
		},
	}

	// 测试 JSON 序列化
	jsonData, err := json.Marshal(ssl)
	if err != nil {
		t.Fatalf("SSL JSON 序列化失败: %v", err)
	}

	var jsonSSL SSL
	if err := json.Unmarshal(jsonData, &jsonSSL); err != nil {
		t.Fatalf("SSL JSON 反序列化失败: %v", err)
	}

	// 验证 JSON 序列化/反序列化后的字段值
	if jsonSSL.Domain != ssl.Domain {
		t.Errorf("Domain 不匹配: 期望 %s, 得到 %s", ssl.Domain, jsonSSL.Domain)
	}

	if jsonSSL.Cert.Certificate != ssl.Cert.Certificate {
		t.Errorf("Certificate 不匹配: 期望 %s, 得到 %s", ssl.Cert.Certificate, jsonSSL.Cert.Certificate)
	}

	if jsonSSL.Cert.CertificateKey != ssl.Cert.CertificateKey {
		t.Errorf("CertificateKey 不匹配: 期望 %s, 得到 %s", ssl.Cert.CertificateKey, jsonSSL.Cert.CertificateKey)
	}

	// 测试 YAML 序列化
	yamlData, err := yaml.Marshal(ssl)
	if err != nil {
		t.Fatalf("SSL YAML 序列化失败: %v", err)
	}

	var yamlSSL SSL
	if err := yaml.Unmarshal(yamlData, &yamlSSL); err != nil {
		t.Fatalf("SSL YAML 反序列化失败: %v", err)
	}

	// 验证 YAML 序列化/反序列化后的字段值
	if yamlSSL.Domain != ssl.Domain {
		t.Errorf("Domain 不匹配: 期望 %s, 得到 %s", ssl.Domain, yamlSSL.Domain)
	}

	t.Logf("SSL 配置测试通过: %s", fmt.PrettyJSON(ssl))
}

func TestHealthCheckConfig(t *testing.T) {
	// 测试健康检查配置
	healthCheck := HealthCheck{
		Outer: HealthCheckOuter{
			Enable: true,
			Path:   "/health",
			Ok:     true,
		},
		Inner: HealthCheckInner{
			Enable:   true,
			Interval: 60,
			Timeout:  5,
		},
	}

	// 测试外部健康检查配置
	if !healthCheck.Outer.Enable {
		t.Error("外部健康检查应该启用")
	}

	if healthCheck.Outer.Path != "/health" {
		t.Errorf("外部健康检查路径不匹配: 期望 /health, 得到 %s", healthCheck.Outer.Path)
	}

	// 测试内部健康检查配置
	if !healthCheck.Inner.Enable {
		t.Error("内部健康检查应该启用")
	}

	if healthCheck.Inner.Interval != 60 {
		t.Errorf("内部健康检查间隔不匹配: 期望 60, 得到 %d", healthCheck.Inner.Interval)
	}

	if healthCheck.Inner.Timeout != 5 {
		t.Errorf("内部健康检查超时不匹配: 期望 5, 得到 %d", healthCheck.Inner.Timeout)
	}

	t.Logf("健康检查配置测试通过: %s", fmt.PrettyJSON(healthCheck))
}

func TestCacheConfig(t *testing.T) {
	// 测试缓存配置
	cache := Cache{
		Host:     "redis.example.com",
		Port:     6379,
		Username: "user",
		Password: "password",
		DB:       1,
		Prefix:   "api-gateway:",
	}

	// 测试 JSON 序列化
	jsonData, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("缓存 JSON 序列化失败: %v", err)
	}

	var jsonCache Cache
	if err := json.Unmarshal(jsonData, &jsonCache); err != nil {
		t.Fatalf("缓存 JSON 反序列化失败: %v", err)
	}

	// 验证 JSON 序列化/反序列化后的字段值
	if jsonCache.Host != cache.Host {
		t.Errorf("Host 不匹配: 期望 %s, 得到 %s", cache.Host, jsonCache.Host)
	}

	if jsonCache.Port != cache.Port {
		t.Errorf("Port 不匹配: 期望 %d, 得到 %d", cache.Port, jsonCache.Port)
	}

	if jsonCache.Username != cache.Username {
		t.Errorf("Username 不匹配: 期望 %s, 得到 %s", cache.Username, jsonCache.Username)
	}

	if jsonCache.Password != cache.Password {
		t.Errorf("Password 不匹配: 期望 %s, 得到 %s", cache.Password, jsonCache.Password)
	}

	if jsonCache.DB != cache.DB {
		t.Errorf("DB 不匹配: 期望 %d, 得到 %d", cache.DB, jsonCache.DB)
	}

	if jsonCache.Prefix != cache.Prefix {
		t.Errorf("Prefix 不匹配: 期望 %s, 得到 %s", cache.Prefix, jsonCache.Prefix)
	}

	t.Logf("缓存配置测试通过: %s", fmt.PrettyJSON(cache))
}
