package core

import (
	"github.com/go-zoox/api-gateway/core/loadbalancer"
	"github.com/go-zoox/api-gateway/plugin/baseuri"
	"github.com/go-zoox/api-gateway/plugin/cors"
	"github.com/go-zoox/api-gateway/plugin/ippolicy"
	"github.com/go-zoox/api-gateway/plugin/jsonaudit"
	"github.com/go-zoox/api-gateway/plugin/ratelimit"
	"github.com/go-zoox/kv"
	"github.com/go-zoox/kv/redis"
)

func (c *core) prepare() error {
	// prepare cache
	if err := c.prepareCache(); err != nil {
		return err
	}

	// prepare load balancer manager
	if err := c.prepareLoadBalancer(); err != nil {
		return err
	}

	// prepare plugins
	if err := c.preparePlugins(); err != nil {
		return err
	}

	return nil
}

func (c *core) prepareLoadBalancer() error {
	// Initialize load balancer manager
	c.lbManager = loadbalancer.NewManager()

	// Start health checks for all backends
	// Default backend
	if c.cfg.Backend.Service.Name != "" || len(c.cfg.Backend.Service.Servers) > 0 {
		normalizedBackend := c.cfg.Backend.Normalize()
		if normalizedBackend != nil {
			if err := c.lbManager.StartHealthCheck(normalizedBackend); err != nil {
				return err
			}
		}
	}

	// Route backends
	for _, route := range c.cfg.Routes {
		normalizedBackend := route.Backend.Normalize()
		if normalizedBackend != nil {
			if err := c.lbManager.StartHealthCheck(normalizedBackend); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *core) prepareCache() error {
	if c.cfg.Cache.Host != "" {
		prefix := c.cfg.Cache.Prefix
		if prefix == "" {
			prefix = "gozoox-api-gateway:"
		}

		c.app.Config.Cache = kv.Config{
			Engine: "redis",
			Config: &redis.Config{
				Host:     c.cfg.Cache.Host,
				Port:     int(c.cfg.Cache.Port),
				Username: c.cfg.Cache.Username,
				Password: c.cfg.Cache.Password,
				DB:       int(c.cfg.Cache.DB),
				Prefix:   prefix,
			},
		}
	}

	return nil
}

func (c *core) preparePlugins() error {
	// buildin plugins
	if err := c.preparePluginsBuildin(); err != nil {
		return err
	}

	for _, plugin := range c.plugins {
		if err := plugin.Prepare(c.app, c.cfg); err != nil {
			return err
		}
	}

	return nil
}

func (c *core) preparePluginsBuildin() error {
	c.app.Logger().Debugf("baseuri: %s", c.cfg.BaseURI)

	// baseuri
	if c.cfg.BaseURI != "" {
		c.plugins = append(c.plugins, &baseuri.BaseURI{
			BaseURI: c.cfg.BaseURI,
		})
	}

	// IP policy (middleware + no-op OnRequest/OnResponse)
	if c.shouldEnableIPPolicy() {
		c.plugins = append(c.plugins, ippolicy.New())
	}

	// CORS (middleware + OnResponse)
	if c.shouldEnableCORS() {
		c.plugins = append(c.plugins, cors.New())
	}

	// rate limit
	if c.shouldEnableRateLimit() {
		c.plugins = append(c.plugins, ratelimit.New())
	}

	// JSON audit (response JSON-like → log request + response)
	if c.shouldEnableJSONAudit() {
		c.plugins = append(c.plugins, jsonaudit.New())
	}

	return nil
}

// shouldEnableJSONAudit is true when global json_audit is enabled or any route enables json_audit.
func (c *core) shouldEnableJSONAudit() bool {
	if c.cfg.JSONAudit.Enable {
		return true
	}
	for _, rt := range c.cfg.Routes {
		if rt.JSONAudit.Enable {
			return true
		}
	}
	return false
}

func (c *core) shouldEnableIPPolicy() bool {
	if c.cfg.IPPolicy.Enable {
		return true
	}
	for _, r := range c.cfg.Routes {
		if r.IPPolicy.Enable {
			return true
		}
	}
	return false
}

func (c *core) shouldEnableCORS() bool {
	if c.cfg.CORS.Enable {
		return true
	}
	for _, r := range c.cfg.Routes {
		if r.CORS.Enable {
			return true
		}
	}
	return false
}

// shouldEnableRateLimit checks if rate limiting should be enabled
func (c *core) shouldEnableRateLimit() bool {
	// Check global rate limit
	if c.cfg.RateLimit.Enable {
		return true
	}

	// Check route-specific rate limits
	for _, route := range c.cfg.Routes {
		if route.RateLimit.Enable {
			return true
		}
	}

	return false
}
