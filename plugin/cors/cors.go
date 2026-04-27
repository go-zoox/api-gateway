package cors

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/plugin"
	"github.com/go-zoox/api-gateway/router"
	"github.com/go-zoox/zoox"
)

type ctxKey int

const stateKey ctxKey = 1

// state is attached to the request for [Plugin.OnResponse].
type state struct {
	cfg *corsConfig
	// origin header from the client (may be empty for same-site requests)
	origin string
}

// Plugin adds CORS: OPTIONS preflight is answered in the HTTP middleware; other responses
// are decorated in [Plugin.OnResponse].
type Plugin struct {
	plugin.Plugin
	cfg   *config.Config
	perM  map[string]*corsConfig
	globl *corsConfig
	ready bool
}

// New creates a CORS plugin.
func New() *Plugin { return &Plugin{perM: make(map[string]*corsConfig)} }

// Prepare validates configuration and registers middleware.
func (p *Plugin) Prepare(app *zoox.Application, cfg *config.Config) error {
	app.Logger().Infof("[plugin:cors] prepare ...")
	p.cfg = cfg
	if err := p.compileAll(cfg); err != nil {
		return err
	}
	if p.ready {
		app.Use(p.handle)
	}
	app.Logger().Infof("[plugin:cors] initialized (enabled=%v, route overrides=%d)", p.ready, len(p.perM))
	return nil
}

// corsConfig is a normalized, non-nil effective policy.
type corsConfig struct {
	allowOrigins     []string
	allowMethods     []string
	allowHeaders     []string
	exposeHeaders    []string
	allowCredentials bool
	maxAge           int64
}

func (p *Plugin) compileAll(cfg *config.Config) error {
	p.perM = make(map[string]*corsConfig)
	p.globl = nil
	p.ready = false

	if cfg.CORS.Enable {
		c, err := normalizeCORS(&cfg.CORS, nil)
		if err != nil {
			return err
		}
		if c != nil {
			p.globl = c
			p.ready = true
		}
	}

	for i := range cfg.Routes {
		rt := &cfg.Routes[i]
		if !rt.CORS.Enable {
			continue
		}
		c, err := normalizeCORS(&cfg.CORS, &rt.CORS)
		if err != nil {
			return err
		}
		if c == nil {
			continue
		}
		p.perM[routeKey(rt)] = c
		p.ready = true
	}
	return nil
}

func routeKey(rt *route.Route) string { return rt.Name + "\x00" + rt.Path }

func (p *Plugin) forRoute(rt *route.Route) *corsConfig {
	if rt == nil {
		return p.globl
	}
	if rt.CORS.Enable {
		if c := p.perM[routeKey(rt)]; c != nil {
			return c
		}
	}
	if p.cfg != nil && p.cfg.CORS.Enable {
		return p.globl
	}
	return nil
}

func (p *Plugin) handle(ctx *zoox.Context) {
	if p.cfg == nil {
		ctx.Next()
		return
	}
	rt, err := router.RouteForPath(p.cfg, ctx.Path)
	if err != nil {
		ctx.Next()
		return
	}
	cfgX := p.forRoute(rt)
	if cfgX == nil {
		ctx.Next()
		return
	}
	origin := ctx.Request.Header.Get("Origin")
	reqMethod := ctx.Request.Method

	// preflight: respond without proxying
	if reqMethod == http.MethodOptions && isPreflight(ctx.Request) {
		if !originOK(cfgX, origin) {
			ctx.String(http.StatusForbidden, "CORS origin not allowed")
			return
		}
		reqMethodHdr := ctx.Request.Header.Get("Access-Control-Request-Method")
		if !methodOK(cfgX, reqMethodHdr) {
			ctx.String(http.StatusForbidden, "CORS method not allowed")
			return
		}
		applyCORSResponse(ctx.Writer.Header(), cfgX, allowOriginValue(cfgX, origin), true)
		if cfgX.maxAge > 0 {
			ctx.Writer.Header().Set("Access-Control-Max-Age", itoa64(cfgX.maxAge))
		}
		ctx.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	// Defer CORS for actual response
	st := &state{cfg: cfgX, origin: origin}
	nctx := context.WithValue(ctx.Request.Context(), stateKey, st)
	ctx.Request = ctx.Request.WithContext(nctx)
	ctx.Next()
}

// OnResponse sets CORS headers on the upstream response.
func (p *Plugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
	v := ctx.Request.Context().Value(stateKey)
	if v == nil {
		return nil
	}
	st, ok := v.(*state)
	if !ok || st == nil || st.cfg == nil {
		return nil
	}
	origin := st.origin
	if res == nil {
		return nil
	}
	ao := allowOriginValue(st.cfg, origin)
	if ao == "" {
		return nil
	}
	applyCORSResponse(res.Header, st.cfg, ao, false)
	if origin != "" {
		if res.Header.Get("Vary") == "" {
			res.Header.Set("Vary", "Origin")
		} else if !strings.Contains(res.Header.Get("Vary"), "Origin") {
			res.Header.Set("Vary", res.Header.Get("Vary")+", Origin")
		}
	}
	return nil
}

func (p *Plugin) OnRequest(_ *zoox.Context, _ *http.Request) error { return nil }

func isPreflight(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.Method != http.MethodOptions {
		return false
	}
	if r.Header.Get("Origin") == "" {
		return false
	}
	if r.Header.Get("Access-Control-Request-Method") == "" {
		return false
	}
	return true
}
