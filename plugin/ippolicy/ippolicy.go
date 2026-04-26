package ippolicy

import (
	"fmt"
	"net/http"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/plugin"
	"github.com/go-zoox/api-gateway/router"
	"github.com/go-zoox/zoox"
)

// IPPolicy enforces per-route or global allow/deny CIDR lists at the edge (including OPTIONS).
type IPPolicy struct {
	plugin.Plugin

	cfg *config.Config
	// key: route name + \x00 + route path, only routes with ip_policy.enable
	perRoute map[string]*compiled
	global *compiled
}

// New creates the IP policy plugin.
func New() *IPPolicy {
	return &IPPolicy{perRoute: make(map[string]*compiled)}
}

// Prepare compiles CIDRs and registers middleware before the reverse proxy.
func (p *IPPolicy) Prepare(app *zoox.Application, cfg *config.Config) error {
	app.Logger().Infof("[plugin:ippolicy] prepare ...")
	p.cfg = cfg

	if err := p.compileAll(cfg); err != nil {
		return err
	}
	app.Use(p.handle)
	app.Logger().Infof("[plugin:ippolicy] initialized (global=%v, route rules=%d)", p.global != nil, len(p.perRoute))
	return nil
}

func (p *IPPolicy) compileAll(cfg *config.Config) error {
	p.perRoute = make(map[string]*compiled)
	p.global = nil

	if cfg.IPPolicy.Enable {
		c, err := newCompiled(cfg.IPPolicy.Allow, cfg.IPPolicy.Deny, cfg.IPPolicy.TrustedProxies, cfg.IPPolicy.Message)
		if err != nil {
			return fmt.Errorf("ip_policy: %w", err)
		}
		p.global = c
	}

	for i := range cfg.Routes {
		rt := &cfg.Routes[i]
		if !rt.IPPolicy.Enable {
			continue
		}
		eff := effectiveIPPolicy(&cfg.IPPolicy, rt)
		if eff == nil {
			continue
		}
		c, err := newCompiled(eff.Allow, eff.Deny, eff.TrustedProxies, eff.Message)
		if err != nil {
			return fmt.Errorf("ip_policy (route %s): %w", rt.Name, err)
		}
		key := routeKey(rt)
		p.perRoute[key] = c
	}
	return nil
}

func effectiveIPPolicy(global *config.IPPolicy, rt *route.Route) *route.IPPolicy {
	if rt == nil || !rt.IPPolicy.Enable {
		if global != nil && global.Enable {
			g := *global
			return &g
		}
		return nil
	}
	r := rt.IPPolicy
	out := route.IPPolicy{
		Enable:         true,
		Allow:          r.Allow,
		Deny:           r.Deny,
		TrustedProxies: r.TrustedProxies,
		Message:        r.Message,
	}
	if global != nil && global.Enable {
		if out.Message == "" {
			out.Message = global.Message
		}
		if len(out.Allow) == 0 {
			out.Allow = global.Allow
		}
		if len(out.Deny) == 0 {
			out.Deny = global.Deny
		} else if len(global.Deny) > 0 {
			out.Deny = append(append([]string{}, global.Deny...), out.Deny...)
		}
		if len(out.TrustedProxies) == 0 {
			out.TrustedProxies = global.TrustedProxies
		}
	}
	if out.Message == "" {
		out.Message = "Forbidden"
	}
	return &out
}

func routeKey(rt *route.Route) string {
	if rt == nil {
		return ""
	}
	return rt.Name + "\x00" + rt.Path
}

func (p *IPPolicy) compiledFor(rt *route.Route) *compiled {
	if rt == nil {
		if p != nil {
			return p.global
		}
		return nil
	}
	if rt.IPPolicy.Enable {
		if c := p.perRoute[routeKey(rt)]; c != nil {
			return c
		}
	}
	if p.cfg != nil && p.cfg.IPPolicy.Enable {
		return p.global
	}
	return nil
}

func (p *IPPolicy) handle(ctx *zoox.Context) {
	rt, err := router.RouteForPath(p.cfg, ctx.Path)
	if err != nil {
		ctx.Next()
		return
	}
	pol := p.compiledFor(rt)
	if pol == nil {
		ctx.Next()
		return
	}
	direct, err := DirectPeerIP(ctx.Request)
	if err != nil {
		ctx.Logger.Warnf("[plugin:ippolicy] direct peer: %v", err)
		ctx.String(http.StatusForbidden, "%s", pol.message)
		return
	}
	client := ClientIP(ctx.Request, direct, pol.trusted)
	if !pol.allows(client) {
		ctx.Logger.Warnf("[plugin:ippolicy] blocked client %s (direct %s) path %s", client, direct, ctx.Path)
		ctx.String(http.StatusForbidden, "%s", pol.message)
		return
	}
	ctx.Next()
}

// OnRequest is a no-op; policy runs in the HTTP middleware.
func (p *IPPolicy) OnRequest(_ *zoox.Context, _ *http.Request) error { return nil }

// OnResponse is a no-op.
func (p *IPPolicy) OnResponse(_ *zoox.Context, _ *http.Response) error { return nil }
