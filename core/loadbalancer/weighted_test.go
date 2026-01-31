package loadbalancer

import (
	"net/http"
	"testing"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

func TestWeightedRoundRobinSelect(t *testing.T) {
	wrr := NewWeightedRoundRobin()

	backend := &route.NormalizedBackend{
		Algorithm: "weighted",
		Servers: []*service.Server{
			{Name: "server1.com", Port: 8080, Weight: 1},
			{Name: "server2.com", Port: 8080, Weight: 2},
			{Name: "server3.com", Port: 8080, Weight: 1},
		},
		BaseConfig: &service.Service{},
	}

	// 设置所有服务器为健康
	for _, server := range backend.Servers {
		server.SetHealthy(true)
		server.Disabled = false
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	// 多次选择，验证权重分配
	selections := make(map[string]int)
	for i := 0; i < 1000; i++ {
		selected, err := wrr.Select(req, backend)
		if err != nil {
			t.Fatalf("Select 不应该返回错误: %v", err)
		}
		selections[selected.ID()]++
	}

	// 验证权重分配（平滑加权轮询在长期运行中会接近权重比例）
	server2Count := selections["server2.com:8080"]
	server1Count := selections["server1.com:8080"]
	server3Count := selections["server3.com:8080"]

	// 计算期望比例：server1=25%, server2=50%, server3=25%
	// 允许一定的误差范围（±10%）
	total := server1Count + server2Count + server3Count
	expectedServer2Ratio := float64(server2Count) / float64(total)
	
	// server2 的权重是 server1 和 server3 的 2 倍，所以应该被选择更多次
	// 在长期运行中，server2 应该被选择大约 50% 的时间
	if expectedServer2Ratio < 0.35 {
		t.Errorf("权重为 2 的服务器选择比例太低: server2=%.2f%%, 期望 >= 35%%", expectedServer2Ratio*100)
	}
	
	// 验证所有服务器都被选择了
	if server1Count == 0 || server2Count == 0 || server3Count == 0 {
		t.Error("所有服务器都应该被选择")
	}

	// 验证所有服务器都被选择了
	if server1Count == 0 || server2Count == 0 || server3Count == 0 {
		t.Error("所有服务器都应该被选择")
	}

	t.Logf("选择分布: server1=%d, server2=%d, server3=%d", server1Count, server2Count, server3Count)
}

func TestWeightedRoundRobinNoHealthyServers(t *testing.T) {
	wrr := NewWeightedRoundRobin()

	backend := &route.NormalizedBackend{
		Algorithm: "weighted",
		Servers: []*service.Server{
			{Name: "server1.com", Port: 8080, Weight: 1},
		},
		BaseConfig: &service.Service{},
	}

	// 设置服务器为不健康
	backend.Servers[0].SetHealthy(false)
	backend.Servers[0].Disabled = false

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	_, err := wrr.Select(req, backend)
	if err == nil {
		t.Error("没有健康服务器时应该返回错误")
	}
}
