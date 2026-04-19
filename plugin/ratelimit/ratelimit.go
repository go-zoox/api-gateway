package ratelimit

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/plugin"
	"github.com/go-zoox/proxy"
	"github.com/go-zoox/zoox"
)

// RateLimit implements rate limiting plugin
type RateLimit struct {
	plugin.Plugin

	// Global configuration
	globalConfig route.RateLimit

	// Route-specific configurations (keyed by route path)
	routeConfigs map[string]*route.RateLimit

	// Components
	extractorFactory *ExtractorFactory
	algorithmFactory *AlgorithmFactory
	limitStore       Storage // zoox.Application.Cache()
}

// New creates a new rate limit plugin
func New() *RateLimit {
	return &RateLimit{
		routeConfigs:     make(map[string]*route.RateLimit),
		extractorFactory: &ExtractorFactory{},
		algorithmFactory: &AlgorithmFactory{},
	}
}

// Prepare initializes the rate limit plugin
func (r *RateLimit) Prepare(app *zoox.Application, cfg *config.Config) error {
	app.Logger().Infof("[plugin:ratelimit] prepare ...")

	// Store global configuration
	r.globalConfig = cfg.RateLimit

	// Store route-specific configurations
	for _, rt := range cfg.Routes {
		if rt.RateLimit.Enable {
			rateLimitConfig := rt.RateLimit
			r.routeConfigs[rt.Path] = &rateLimitConfig
		}
	}

	r.attachApplicationCache(app)

	app.Logger().Infof("[plugin:ratelimit] rate limit plugin initialized")
	return nil
}

func (r *RateLimit) attachApplicationCache(app *zoox.Application) {
	if app == nil {
		return
	}
	r.limitStore = newCacheStorage(app.Cache())
}

// OnRequest checks rate limit before forwarding request
func (r *RateLimit) OnRequest(ctx *zoox.Context, req *http.Request) error {
	rateLimitConfig := r.getRateLimitConfig(ctx.Path)
	if rateLimitConfig == nil || !rateLimitConfig.Enable {
		return nil
	}

	if rateLimitConfig.Limit <= 0 {
		ctx.Logger.Warnf("[plugin:ratelimit] invalid limit: %d", rateLimitConfig.Limit)
		return nil
	}

	if rateLimitConfig.Window <= 0 {
		ctx.Logger.Warnf("[plugin:ratelimit] invalid window: %d", rateLimitConfig.Window)
		return nil
	}

	extractor := r.extractorFactory.NewExtractor(rateLimitConfig.KeyType, rateLimitConfig.KeyHeader)
	key, err := extractor.Extract(ctx, req)
	if err != nil {
		ctx.Logger.Warnf("[plugin:ratelimit] failed to extract key: %s", err)
		return nil
	}

	if r.limitStore == nil {
		ctx.Logger.Warnf("[plugin:ratelimit] storage not initialized, skip rate limiting")
		return nil
	}

	algorithm := r.algorithmFactory.NewAlgorithm(rateLimitConfig.Algorithm)

	window := time.Duration(rateLimitConfig.Window) * time.Second
	allowed, remaining, resetTime, err := algorithm.Allow(
		context.Background(),
		r.limitStore,
		key,
		rateLimitConfig.Limit,
		window,
		rateLimitConfig.Burst,
	)

	if err != nil {
		ctx.Logger.Errorf("[plugin:ratelimit] rate limit check failed: %s", err)
		return nil
	}

	reqCtx := req.Context()
	reqCtx = context.WithValue(reqCtx, "ratelimit:limit", rateLimitConfig.Limit)
	reqCtx = context.WithValue(reqCtx, "ratelimit:remaining", remaining)
	reqCtx = context.WithValue(reqCtx, "ratelimit:reset", resetTime.Unix())
	*req = *req.WithContext(reqCtx)

	if !allowed {
		message := rateLimitConfig.Message
		if message == "" {
			message = "Too Many Requests"
		}

		retryAfter := int64(time.Until(resetTime).Seconds())
		if retryAfter < 0 {
			retryAfter = 0
		}

		reqCtx = context.WithValue(reqCtx, "ratelimit:retryAfter", retryAfter)
		*req = *req.WithContext(reqCtx)

		if ctx.Writer != nil {
			ctx.Writer.Header().Set("X-RateLimit-Limit", strconv.FormatInt(rateLimitConfig.Limit, 10))
			ctx.Writer.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
			ctx.Writer.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			ctx.Writer.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))

			for headerName, headerValue := range rateLimitConfig.Headers {
				ctx.Writer.Header().Set(headerName, headerValue)
			}
		}

		ctx.Logger.Warnf("[plugin:ratelimit] rate limit exceeded for key: %s, remaining: %d", key, remaining)
		return proxy.NewHTTPError(429, message)
	}

	return nil
}

// OnResponse is called after receiving response
func (r *RateLimit) OnResponse(ctx *zoox.Context, res *http.Response) error {
	reqCtx := ctx.Request.Context()
	if limit, ok := reqCtx.Value("ratelimit:limit").(int64); ok {
		res.Header.Set("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
	}
	if remaining, ok := reqCtx.Value("ratelimit:remaining").(int64); ok {
		res.Header.Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
	}
	if reset, ok := reqCtx.Value("ratelimit:reset").(int64); ok {
		res.Header.Set("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
	}
	if retryAfter, ok := reqCtx.Value("ratelimit:retryAfter").(int64); ok {
		res.Header.Set("Retry-After", strconv.FormatInt(retryAfter, 10))
	}
	return nil
}

func (r *RateLimit) getRateLimitConfig(path string) *route.RateLimit {
	routePaths := make([]string, 0, len(r.routeConfigs))
	for routePath := range r.routeConfigs {
		routePaths = append(routePaths, routePath)
	}
	sort.Slice(routePaths, func(i, j int) bool {
		return len(routePaths[i]) > len(routePaths[j])
	})

	for _, routePath := range routePaths {
		config := r.routeConfigs[routePath]
		if path == routePath {
			return config
		}
		if len(path) > len(routePath) && path[:len(routePath)] == routePath {
			if path[len(routePath)] == '/' {
				return config
			}
		}
	}

	if r.globalConfig.Enable {
		return &r.globalConfig
	}

	return nil
}
