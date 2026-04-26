package cors

import (
	"fmt"
	"strings"

	"github.com/go-zoox/api-gateway/core/route"
)

func normalizeCORS(global *route.CORS, r *route.CORS) (*corsConfig, error) {
	var eff *route.CORS
	switch {
	case r != nil && r.Enable:
		eff = mergeCORSGlobalRoute(global, r)
	case global != nil && global.Enable:
		g := *global
		eff = &g
	default:
		return nil, nil
	}

	orig := eff.AllowOrigins
	if len(orig) == 0 {
		orig = []string{"*"}
	}
	if eff.AllowCredentials {
		for _, o := range orig {
			if o == "*" {
				return nil, fmt.Errorf("cors: allow_credentials is incompatible with allow_origins *")
			}
		}
	}
	methods := eff.AllowMethods
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	}
	headers := eff.AllowHeaders
	if len(headers) == 0 {
		headers = []string{"*"}
	}
	for i := range methods {
		methods[i] = strings.ToUpper(strings.TrimSpace(methods[i]))
	}
	return &corsConfig{
		allowOrigins:     orig,
		allowMethods:     methods,
		allowHeaders:     headers,
		exposeHeaders:    eff.ExposeHeaders,
		allowCredentials: eff.AllowCredentials,
		maxAge:           eff.MaxAge,
	}, nil
}

// mergeCORSGlobalRoute uses route for fields the route sets; empty slices fall back to global.
func mergeCORSGlobalRoute(global *route.CORS, r *route.CORS) *route.CORS {
	if global == nil {
		out := *r
		return &out
	}
	out := *r
	if len(out.AllowOrigins) == 0 {
		out.AllowOrigins = global.AllowOrigins
	}
	if len(out.AllowMethods) == 0 {
		out.AllowMethods = global.AllowMethods
	}
	if len(out.AllowHeaders) == 0 {
		out.AllowHeaders = global.AllowHeaders
	}
	if len(out.ExposeHeaders) == 0 {
		out.ExposeHeaders = global.ExposeHeaders
	}
	if out.MaxAge == 0 {
		out.MaxAge = global.MaxAge
	}
	return &out
}
