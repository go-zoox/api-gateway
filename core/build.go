package core

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-zoox/logger"
	"github.com/go-zoox/proxy"
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func (c *core) build() error {
	// middlewares
	c.app.Use(func(ctx *zoox.Context) {
		if c.cfg.HealthCheck.Outer.Enable {
			if ctx.Path == c.cfg.HealthCheck.Outer.Path {
				if c.cfg.HealthCheck.Outer.Ok {
					ctx.String(http.StatusOK, "ok")
					return
				}
			}
		}

		ctx.Next()
	})

	// services (core plugin)
	c.app.Use(middleware.Proxy(func(ctx *zoox.Context, cfg *middleware.ProxyConfig) (next, stop bool, err error) {
		method := ctx.Method
		path := ctx.Path

		r, err := c.match(ctx, path)
		if err != nil {
			logger.Errorf("[build][path: %s] failed to match route: %s", ctx.Path, err)
			//
			return false, false, proxy.NewHTTPError(404, err.Error())
		}

		if r == nil {
			return true, false, nil
		}

		// Normalize backend (handles both single-server and multi-server modes)
		normalizedBackend := r.Backend.Normalize()
		if normalizedBackend == nil {
			logger.Errorf("[build][path: %s] no backend configured", ctx.Path)
			return false, false, proxy.NewHTTPError(503, "no backend configured")
		}

		// Get load balancer
		lb := c.lbManager.GetLoadBalancer(normalizedBackend.Algorithm)

		// Select server using load balancer
		server, err := lb.Select(ctx.Request, normalizedBackend)
		if err != nil {
			logger.Errorf("[build][path: %s] failed to select server: %s", ctx.Path, err)
			return false, false, proxy.NewHTTPError(503, err.Error())
		}

		// Get effective configuration (merges base config with server-specific overrides)
		effectiveService := server.GetEffectiveConfig(normalizedBackend.BaseConfig)

		// DNS check (optional, can be skipped based on health check)
		if _, err := effectiveService.CheckDNS(); err != nil {
			logger.Errorf("check dns error: %s", err)
			return false, false, proxy.NewHTTPError(503, ErrServiceUnavailable.Error())
		}

		// Track connection for least-connections algorithm
		var cleanupOnce *sync.Once
		if normalizedBackend.Algorithm == "least-connections" {
			lb.OnRequestStart(normalizedBackend, server)

			// Ensure OnRequestEnd is called even if proxy fails
			// Use sync.Once to ensure it's only called once
			cleanupOnce = &sync.Once{}
			cleanup := func() {
				cleanupOnce.Do(func() {
					lb.OnRequestEnd(normalizedBackend, server)
				})
			}

			// Monitor request context to ensure cleanup on failure
			go func() {
				<-ctx.Request.Context().Done()
				cleanup()
			}()
		}

		cfg.OnRequest = func(req, inReq *http.Request) error {
			req.URL.Scheme = effectiveService.Protocol
			req.URL.Host = server.Host()
			req.Host = req.URL.Host

			// apply path rewrite using effective service config
			req.URL.Path = effectiveService.Rewrite(req.URL.Path)

			// apply headers
			if effectiveService.Request.Headers != nil {
				for k, v := range effectiveService.Request.Headers {
					req.Header.Set(k, v)
				}
			}

			// apply query
			if effectiveService.Request.Query != nil {
				originQuery := req.URL.Query()
				for k, v := range effectiveService.Request.Query {
					originQuery.Set(k, v)
				}
				req.URL.RawQuery = originQuery.Encode()
			}

			for _, plugin := range c.plugins {
				if err := plugin.OnRequest(ctx, ctx.Request); err != nil {
					return err
				}
			}

			ctx.Logger.Infof("[route: %s] %s %s => %s (path: %s, algorithm: %s)", r.Name, method, path, server.Target(), req.URL.Path, normalizedBackend.Algorithm)

			return nil
		}

		cfg.OnResponse = func(res *http.Response, inReq *http.Request) error {
			// Apply response headers from effective service config
			if effectiveService.Response.Headers != nil {
				for k, v := range effectiveService.Response.Headers {
					res.Header.Set(k, v)
				}
			}

			for _, plugin := range c.plugins {
				if err := plugin.OnResponse(ctx, res); err != nil {
					return err
				}
			}

			res.Header.Set("X-Powered-By", fmt.Sprintf("gozoox-api-gateway/%s", c.version))

			// Decrement connection count for least-connections algorithm
			// This ensures cleanup happens when OnResponse is called (success case)
			if normalizedBackend.Algorithm == "least-connections" && cleanupOnce != nil {
				cleanupOnce.Do(func() {
					lb.OnRequestEnd(normalizedBackend, server)
				})
			}

			return nil
		}

		return
	}))

	return nil
}
