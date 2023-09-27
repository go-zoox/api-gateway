package core

import (
	"fmt"
	"net/http"

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

	// plugins
	c.app.Use(middleware.Proxy(func(ctx *zoox.Context, cfg *middleware.ProxyConfig) (next bool, err error) {
		cfg.OnRequest = func(req, inReq *http.Request) error {
			for _, plugin := range c.plugins {
				if err := plugin.OnRequest(ctx, ctx.Request); err != nil {
					return err
				}
			}

			return nil
		}

		cfg.OnResponse = func(res *http.Response, inReq *http.Request) error {
			for _, plugin := range c.plugins {
				if err := plugin.OnResponse(ctx, ctx.Writer); err != nil {
					return err
				}
			}

			return nil
		}

		return true, nil
	}))

	// services (core plugin)
	c.app.Use(middleware.Proxy(func(ctx *zoox.Context, cfg *middleware.ProxyConfig) (next bool, err error) {
		method := ctx.Method
		path := ctx.Path

		r, err := c.match(ctx, path)
		if err != nil {
			logger.Errorf("failed to get config: %s", err)
			//
			return false, proxy.NewHTTPError(404, "Not Found")
		}

		if r == nil {
			return true, nil
		}

		cfg.OnRequest = func(req, inReq *http.Request) error {
			req.URL.Scheme = r.Backend.Service.Protocol
			req.URL.Host = r.Backend.Service.Host()
			req.Host = req.URL.Host

			// apply path
			req.URL.Path = r.Rewrite(req.URL.Path)

			// apply headers
			for k, v := range r.Backend.Service.Request.Headers {
				req.Header.Set(k, v)
			}

			// apply query
			if r.Backend.Service.Request.Query != nil {
				originQuery := req.URL.Query()
				for k, v := range r.Backend.Service.Request.Query {
					originQuery.Set(k, v)
				}
				req.URL.RawQuery = originQuery.Encode()
			}

			ctx.Logger.Infof("[route: %s] %s %s => %s (path: %s)", r.Name, method, path, r.Backend.Service.Target(), req.URL.Path)

			return nil
		}

		cfg.OnResponse = func(res *http.Response, inReq *http.Request) error {
			for k, v := range r.Backend.Service.Response.Headers {
				ctx.Writer.Header().Set(k, v)
			}

			ctx.Writer.Header().Set("X-Proxy-By", fmt.Sprintf("gozoox-api-gateway/%s", c.version))
			return nil
		}

		return
	}))

	return nil
}
