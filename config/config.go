package config

import (
	"github.com/go-zoox/api-gateway/core/route"
)

type Config struct {
	Port int64 `config:"port"`
	// BaseURI is the base uri of api gateway, which is used to generate the proxy request url
	BaseURI string `config:"baseuri"`
	//
	Backend route.Backend `config:"backend"`
	//
	Routes []route.Route `config:"routes"`
	//
	Cache Cache `config:"cache"`
	//
	HealthCheck HealthCheck `config:"healthcheck"`
	//
	RateLimit RateLimit `config:"rate_limit"`
	//
	// Match func(path string) (r *route.Route, err error)
}

type HealthCheck struct {
	Outer HealthCheckOuter `config:"outer"`
	Inner HealthCheckInner `config:"inner"`
}

type HealthCheckOuter struct {
	Enable bool `config:"enable"`
	// Path is the health check request path
	Path string `config:"path"`
	// Ok means all health check request returns ok
	Ok bool `config:"ok"`
}

type HealthCheckInner struct {
	Enable bool `config:"enable"`
	//
	Interval int64 `config:"interval"`
	Timeout  int64 `config:"timeout"`
}

type Cache struct {
	Host     string `config:"host"`
	Port     int64  `config:"port"`
	Username string `config:"username"`
	Password string `config:"password"`
	DB       int64  `config:"db"`
	Prefix   string `config:"prefix"`
}

type SSL struct {
	Domain string  `config:"domain"`
	Cert   SSLCert `config:"cert"`
}

type SSLCert struct {
	Certificate    string `config:"certificate"`
	CertificateKey string `config:"certificate_key"`
}

// RateLimit uses route.RateLimit to avoid circular dependency
type RateLimit = route.RateLimit
