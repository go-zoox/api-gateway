package router

import (
	"errors"
	"regexp"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/core-utils/fmt"
	"github.com/go-zoox/core-utils/strings"
)

// ErrRouteNotFound is returned when no route matches the path and no default backend is configured.
var ErrRouteNotFound = errors.New("route not found")

// MatchPath returns the first route in the slice (config order) whose path rule matches the request path.
// For prefix routes, the first match wins; order matters in configuration.
func MatchPath(routes []route.Route, path string) (*route.Route, error) {
	for _, r := range routes {
		switch r.PathType {
		case "prefix", "":
			if strings.StartsWith(path, r.Path) {
				return &r, nil
			}
		case "regex":
			if ok, err := regexp.MatchString(r.Path, path); err != nil {
				return nil, err
			} else if ok {
				return &r, nil
			}
		default:
			return nil, fmt.Errorf("unsupport path type: %s", r.PathType)
		}
	}
	return nil, ErrRouteNotFound
}

// RouteForPath returns the service route for a request path, using the first matching route, or
// the default backend from cfg when no route matches and default backend is configured.
func RouteForPath(cfg *config.Config, path string) (*route.Route, error) {
	r, err := MatchPath(cfg.Routes, path)
	if err != nil {
		if !errors.Is(err, ErrRouteNotFound) {
			return nil, err
		}
	} else {
		return r, nil
	}

	if cfg.Backend.Service.Name != "" {
		return &route.Route{
			Name:    "default",
			Backend: cfg.Backend,
		}, nil
	}
	return nil, ErrRouteNotFound
}
