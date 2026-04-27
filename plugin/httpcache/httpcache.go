package httpcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/plugin"
	zc "github.com/go-zoox/cache"
	"github.com/go-zoox/zoox"
)

// HTTPCachePlugin caches GET (or configured methods) responses in Application.Cache().
type HTTPCachePlugin struct {
	plugin.Plugin

	global *route.HTTPCache
	routes map[string]*route.HTTPCache
	store  zc.Cache
}

// New creates the plugin instance; call Prepare from core.
func New() *HTTPCachePlugin {
	return &HTTPCachePlugin{
		routes: make(map[string]*route.HTTPCache),
	}
}

// Prepare loads config and attaches the gateway cache backend.
func (h *HTTPCachePlugin) Prepare(app *zoox.Application, cfg *config.Config) error {
	app.Logger().Infof("[plugin:httpcache] prepare ...")

	h.global = &route.HTTPCache{}
	*h.global = cfg.HTTPCache

	for _, rt := range cfg.Routes {
		if rt.HTTPCache.Enable {
			c := rt.HTTPCache
			h.routes[rt.Path] = &c
		}
	}

	h.store = app.Cache()
	if h.store == nil {
		app.Logger().Warnf("[plugin:httpcache] application cache is nil; HTTP response cache disabled at runtime")
	}

	app.Logger().Infof("[plugin:httpcache] initialized (route overrides=%d)", len(h.routes))
	return nil
}

// OnRequest serves cache HITs before the upstream is contacted.
func (h *HTTPCachePlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
	cfg := h.getConfigForPath(ctx.Path)
	if cfg == nil || !cfg.Enable || h.store == nil {
		return nil
	}

	if !h.methodAllowed(req.Method, cfg) {
		return nil
	}

	if !cfg.CacheAuthorizedRequests {
		if req.Header.Get("Authorization") != "" || req.Header.Get("Cookie") != "" {
			return nil
		}
	}

	key, err := h.buildCacheKey(cfg, req, ctx.Path)
	if err != nil {
		ctx.Logger.Warnf("[plugin:httpcache] cache key: %v", err)
		return nil
	}

	var entry cachedEntry
	if err := h.store.Get(key, &entry); err != nil {
		return nil
	}
	if entry.StatusCode < http.StatusOK {
		return nil
	}

	h.writeHit(ctx, &entry)
	return ErrTerminalResponse
}

// OnResponse stores cacheable responses after the upstream returns.
func (h *HTTPCachePlugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
	if res == nil {
		return nil
	}

	cfg := h.getConfigForPath(ctx.Path)
	if cfg == nil || !cfg.Enable || h.store == nil {
		return nil
	}

	req := ctx.Request
	if !h.methodAllowed(req.Method, cfg) {
		return nil
	}

	if !cfg.CacheAuthorizedRequests {
		if req.Header.Get("Authorization") != "" || req.Header.Get("Cookie") != "" {
			return nil
		}
	}

	if cfg.SkipResponsesWithSetCookie && res.Header.Get("Set-Cookie") != "" {
		return nil
	}

	if cfg.RespectUpstreamCacheControl && cacheControlBlocks(res.Header.Get("Cache-Control")) {
		return nil
	}

	statusMin, statusMax := cfg.EffectiveStatusRange()
	if res.StatusCode < statusMin || res.StatusCode > statusMax {
		return nil
	}

	maxBody := cfg.EffectiveMaxBodyBytes()
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxBody+1))
	if err != nil {
		return err
	}
	res.Body = io.NopCloser(bytes.NewReader(raw))
	if int64(len(raw)) > maxBody {
		return nil
	}

	ttl := cfg.EffectiveTTL()
	if ttl <= 0 {
		return nil
	}

	entry := cachedEntry{
		StatusCode: res.StatusCode,
		Headers:    cloneHeadersForCache(res.Header),
		Body:       raw,
	}

	key, err := h.buildCacheKey(cfg, req, ctx.Path)
	if err != nil {
		ctx.Logger.Warnf("[plugin:httpcache] cache key (store): %v", err)
		return nil
	}

	if err := h.store.Set(key, &entry, ttl); err != nil {
		ctx.Logger.Warnf("[plugin:httpcache] cache set failed: %v", err)
	}

	res.Header.Set("X-Cache", "MISS")

	return nil
}

func (h *HTTPCachePlugin) writeHit(ctx *zoox.Context, e *cachedEntry) {
	hdrp := ctx.Writer.Header()
	for k, vals := range e.Headers {
		k = http.CanonicalHeaderKey(k)
		for _, v := range vals {
			hdrp.Add(k, v)
		}
	}
	hdrp.Set("X-Cache", "HIT")
	ctx.Writer.WriteHeader(e.StatusCode)
	_, _ = ctx.Writer.Write(e.Body)
}

func (h *HTTPCachePlugin) methodAllowed(method string, cfg *route.HTTPCache) bool {
	methods := cfg.Methods
	if len(methods) == 0 {
		return method == http.MethodGet
	}
	for _, m := range methods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

func (h *HTTPCachePlugin) buildCacheKey(cfg *route.HTTPCache, req *http.Request, gatewayPath string) (string, error) {
	hh := sha256.New()
	_, _ = hh.Write([]byte(req.Method))
	_, _ = hh.Write([]byte{0})
	_, _ = hh.Write([]byte(gatewayPath))
	_, _ = hh.Write([]byte{0})
	_, _ = hh.Write([]byte(req.URL.Path))
	_, _ = hh.Write([]byte{0})
	if cfg.IncludeQueryInKey() {
		_, _ = hh.Write([]byte(normalizedQueryString(req.URL.RawQuery)))
	}
	_, _ = hh.Write([]byte{0})
	for _, name := range cfg.NormalizedVaryHeaders() {
		v := strings.TrimSpace(sortedHeaderField(req.Header, name))
		_, _ = hh.Write([]byte(strings.ToLower(name)))
		_, _ = hh.Write([]byte{0})
		_, _ = hh.Write([]byte(v))
		_, _ = hh.Write([]byte{0})
	}
	sum := hex.EncodeToString(hh.Sum(nil))

	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "httpcache"
	}
	return prefix + ":" + sum, nil
}

func cacheControlBlocks(directive string) bool {
	d := strings.ToLower(directive)
	return strings.Contains(d, "no-store") || strings.Contains(d, "private")
}

func cloneHeadersForCache(h http.Header) map[string][]string {
	out := make(map[string][]string)
	skip := hopByHopLookup()
	for k, vals := range h {
		if skip[strings.ToLower(k)] {
			continue
		}
		cp := make([]string, len(vals))
		copy(cp, vals)
		out[k] = cp
	}
	return out
}

func hopByHopLookup() map[string]bool {
	return map[string]bool{
		"connection":          true,
		"keep-alive":          true,
		"proxy-authenticate":  true,
		"proxy-authorization": true,
		"te":                  true,
		"trailer":             true,
		"transfer-encoding":   true,
		"upgrade":             true,
	}
}

func (h *HTTPCachePlugin) getConfigForPath(path string) *route.HTTPCache {
	routePaths := make([]string, 0, len(h.routes))
	for p := range h.routes {
		routePaths = append(routePaths, p)
	}
	sort.Slice(routePaths, func(i, j int) bool {
		return len(routePaths[i]) > len(routePaths[j])
	})

	for _, routePath := range routePaths {
		cfg := h.routes[routePath]
		if path == routePath {
			return cfg
		}
		if len(path) > len(routePath) && path[:len(routePath)] == routePath {
			if path[len(routePath)] == '/' {
				return cfg
			}
		}
	}

	if h.global != nil && h.global.Enable {
		return h.global
	}
	return nil
}
