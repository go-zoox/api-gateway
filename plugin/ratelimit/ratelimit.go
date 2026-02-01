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
	"github.com/go-zoox/core-utils/fmt"
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
	storageFactory   *StorageFactory

	// Storage instances (keyed by storage type)
	storages map[string]Storage
}

// New creates a new rate limit plugin
func New() *RateLimit {
	return &RateLimit{
		routeConfigs:     make(map[string]*route.RateLimit),
		extractorFactory: &ExtractorFactory{},
		algorithmFactory: &AlgorithmFactory{},
		storageFactory:   &StorageFactory{},
		storages:         make(map[string]Storage),
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

	// Initialize storages
	if err := r.initializeStorages(app, cfg); err != nil {
		app.Logger().Warnf("[plugin:ratelimit] failed to initialize storages: %s, falling back to memory", err)
		// Fallback to memory storage
		memoryStorage, _ := r.storageFactory.NewStorage("memory", nil)
		if memoryStorage != nil {
			r.storages["memory"] = memoryStorage
		}
	}

	app.Logger().Infof("[plugin:ratelimit] rate limit plugin initialized")
	return nil
}

// initializeStorages initializes storage instances
func (r *RateLimit) initializeStorages(app *zoox.Application, cfg *config.Config) error {
	// Initialize memory storage (always available)
	memoryStorage, err := r.storageFactory.NewStorage("memory", nil)
	if err != nil {
		return fmt.Errorf("failed to create memory storage: %w", err)
	}
	r.storages["memory"] = memoryStorage

	// Initialize Redis storage if cache is configured
	if cfg.Cache.Host != "" {
		// Try to get cache from app config
		// Note: This is a simplified approach - in practice, you might need
		// to access the cache differently based on the zoox framework
		redisStorage, err := r.storageFactory.NewStorage("redis", app.Config.Cache)
		if err != nil {
			app.Logger().Warnf("[plugin:ratelimit] failed to create redis storage: %s", err)
		} else {
			r.storages["redis"] = redisStorage
		}
	}

	return nil
}

// OnRequest checks rate limit before forwarding request
func (r *RateLimit) OnRequest(ctx *zoox.Context, req *http.Request) error {
	// Get route-specific or global rate limit configuration
	rateLimitConfig := r.getRateLimitConfig(ctx.Path)
	if rateLimitConfig == nil || !rateLimitConfig.Enable {
		return nil // Rate limiting disabled for this route
	}

	// Validate configuration
	if rateLimitConfig.Limit <= 0 {
		ctx.Logger.Warnf("[plugin:ratelimit] invalid limit: %d", rateLimitConfig.Limit)
		return nil // Invalid config, skip rate limiting
	}

	if rateLimitConfig.Window <= 0 {
		ctx.Logger.Warnf("[plugin:ratelimit] invalid window: %d", rateLimitConfig.Window)
		return nil // Invalid config, skip rate limiting
	}

	// Extract key
	extractor := r.extractorFactory.NewExtractor(rateLimitConfig.KeyType, rateLimitConfig.KeyHeader)
	key, err := extractor.Extract(ctx, req)
	if err != nil {
		ctx.Logger.Warnf("[plugin:ratelimit] failed to extract key: %s", err)
		return nil // Skip rate limiting on extraction error
	}

	// Get storage
	storageType := rateLimitConfig.Storage
	if storageType == "" {
		storageType = "memory"
	}
	storage, ok := r.storages[storageType]
	if !ok {
		ctx.Logger.Warnf("[plugin:ratelimit] storage not found: %s, falling back to memory", storageType)
		storage = r.storages["memory"]
		if storage == nil {
			return nil // No storage available, skip rate limiting
		}
	}

	// Get algorithm
	algorithm := r.algorithmFactory.NewAlgorithm(rateLimitConfig.Algorithm)

	// Check rate limit
	window := time.Duration(rateLimitConfig.Window) * time.Second
	allowed, remaining, resetTime, err := algorithm.Allow(
		context.Background(),
		storage,
		key,
		rateLimitConfig.Limit,
		window,
		rateLimitConfig.Burst,
	)

	if err != nil {
		ctx.Logger.Errorf("[plugin:ratelimit] rate limit check failed: %s", err)
		return nil // On error, allow request (fail open)
	}

	// Store rate limit info in request context for OnResponse to set headers
	reqCtx := req.Context()
	reqCtx = context.WithValue(reqCtx, "ratelimit:limit", rateLimitConfig.Limit)
	reqCtx = context.WithValue(reqCtx, "ratelimit:remaining", remaining)
	reqCtx = context.WithValue(reqCtx, "ratelimit:reset", resetTime.Unix())
	*req = *req.WithContext(reqCtx)

	if !allowed {
		// Rate limit exceeded
		message := rateLimitConfig.Message
		if message == "" {
			message = "Too Many Requests"
		}

		// Calculate Retry-After
		retryAfter := int64(time.Until(resetTime).Seconds())
		if retryAfter < 0 {
			retryAfter = 0
		}

		// Set custom headers if configured (will be applied in error response)
		// For now, we'll rely on the error response to include these

		ctx.Logger.Warnf("[plugin:ratelimit] rate limit exceeded for key: %s, remaining: %d", key, remaining)
		return proxy.NewHTTPError(429, message)
	}

	return nil
}

// OnResponse is called after receiving response
func (r *RateLimit) OnResponse(ctx *zoox.Context, res *http.Response) error {
	// Set rate limit headers from request context
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
	return nil
}

// getRateLimitConfig gets rate limit configuration for a route
// Route-specific config takes precedence over global config
func (r *RateLimit) getRateLimitConfig(path string) *route.RateLimit {
	// Collect route paths and sort by length (longest first) for deterministic matching
	routePaths := make([]string, 0, len(r.routeConfigs))
	for routePath := range r.routeConfigs {
		routePaths = append(routePaths, routePath)
	}
	sort.Slice(routePaths, func(i, j int) bool {
		return len(routePaths[i]) > len(routePaths[j])
	})

	// Check route-specific config in order of path length (longest first)
	for _, routePath := range routePaths {
		config := r.routeConfigs[routePath]
		// Exact match
		if path == routePath {
			return config
		}
		// Prefix match: path must start with routePath AND the next character (if any) must be '/'
		if len(path) > len(routePath) && path[:len(routePath)] == routePath {
			// Check if the character after the prefix is a path separator
			if path[len(routePath)] == '/' {
				return config
			}
		}
	}

	// Fall back to global config
	if r.globalConfig.Enable {
		return &r.globalConfig
	}

	return nil
}
