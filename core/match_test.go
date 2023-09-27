package core

import (
	"testing"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

func TestMatchPath(t *testing.T) {
	routes := []route.Route{
		{
			Name:     "task deploy service",
			Path:     "^/ip1/\\d+/deploy",
			PathType: "regex",
			Backend: route.Backend{
				Service: service.Service{
					Protocol: "https",
					Name:     "task.httpbin.zcorky.com",
					Port:     443,
				},
			},
		},
		{
			Name: "ip1",
			Path: "/ip1",
			Backend: route.Backend{
				Service: service.Service{
					Protocol: "http",
					Name:     "ip3.httpbin.zcorky.com",
					Port:     443,
				},
			},
		},
		{
			Name: "ip2",
			Path: "/ip2",
			Backend: route.Backend{
				Service: service.Service{
					Protocol: "https",
					Name:     "ip2.httpbin.zcorky.com",
					Port:     443,
				},
			},
		},
	}

	s, err := MatchPath(routes, "/ip")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if s != nil {
		t.Fatalf("expected nil, got %v", s)
	}

	s, err = MatchPath(routes, "/ip1")
	if err != nil {
		t.Fatal(err)
	}
	if s.Backend.Service.Name != "ip3.httpbin.zcorky.com" {
		t.Fatalf("expected ip3.httpbin.zcorky.com, got %s", s.Backend.Service.Name)
	}
	if s.Backend.Service.Port != 443 {
		t.Fatalf("expected 443, got %d", s.Backend.Service.Port)
	}
	if s.Backend.Service.Protocol != "http" {
		t.Fatalf("expected http, got %s", s.Backend.Service.Protocol)
	}

	s, err = MatchPath(routes, "/ip2")
	if err != nil {
		t.Fatal(err)
	}
	if s.Backend.Service.Name != "ip2.httpbin.zcorky.com" {
		t.Fatalf("expected ip2.httpbin.zcorky.com, got %s", s.Backend.Service.Name)
	}
	if s.Backend.Service.Port != 443 {
		t.Fatalf("expected 443, got %d", s.Backend.Service.Port)
	}
	if s.Backend.Service.Protocol != "https" {
		t.Fatalf("expected https, got %s", s.Backend.Service.Protocol)
	}

	// regex
	s, err = MatchPath(routes, "/ip1/123/deploy")
	if err != nil {
		t.Fatal(err)
	}
	if s.Backend.Service.Name != "task.httpbin.zcorky.com" {
		t.Fatalf("expected task.httpbin.zcorky.com, got %s", s.Backend.Service.Name)
	}
	if s.Backend.Service.Port != 443 {
		t.Fatalf("expected 443, got %d", s.Backend.Service.Port)
	}
	if s.Backend.Service.Protocol != "https" {
		t.Fatalf("expected https, got %s", s.Backend.Service.Protocol)
	}

	s, err = MatchPath(routes, "/ip1/123")
	if err != nil {
		t.Fatal(err)
	}
	if s.Backend.Service.Name != "ip3.httpbin.zcorky.com" {
		t.Fatalf("expected ip3.httpbin.zcorky.com, got %s", s.Backend.Service.Name)
	}
	if s.Backend.Service.Port != 443 {
		t.Fatalf("expected 443, got %d", s.Backend.Service.Port)
	}
	if s.Backend.Service.Protocol != "http" {
		t.Fatalf("expected http, got %s", s.Backend.Service.Protocol)
	}
}
