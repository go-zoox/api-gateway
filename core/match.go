package core

import (
	"errors"
	"regexp"
	"time"

	"github.com/go-zoox/core-utils/fmt"

	"github.com/go-zoox/core-utils/strings"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
	"github.com/go-zoox/zoox"
)

type Matcher struct {
	Service service.Service
}

func (c *core) match(ctx *zoox.Context, path string) (s *route.Route, err error) {
	key := fmt.Sprintf("match.path:%s", path)
	matcher := &route.Route{}
	if err := ctx.Cache().Get(key, matcher); err == nil {
		return matcher, nil
	}

	matcher, err = MatchPath(c.cfg.Routes, path)
	if err != nil {
		if !errors.Is(err, ErrRouteNotFound) {
			return nil, err
		}
	}

	// main service
	s = matcher

	// // match func
	// if s == nil {
	// 	if c.cfg.Match != nil {
	// 		sm, err := c.cfg.Match(path)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		s = sm
	// 	}
	// }

	// use default service if not found
	if s == nil && c.cfg.Backend.Service.Name != "" {
		s = &route.Route{
			Name:    "default",
			Backend: c.cfg.Backend,
		}
	}

	// route not found or not match, return error
	if s == nil {
		return nil, ErrRouteNotFound
	}

	ctx.Cache().Set(key, s, 60*time.Second)

	return s, nil
}

func MatchPath(routes []route.Route, path string) (r *route.Route, err error) {
	// fmt.PrintJSON(routes)
	for _, route := range routes {
		// fmt.Println("starts woth =>", len(routes), path, route.Path, strings.StartsWith(path, route.Path))
		switch route.PathType {
		case "prefix", "":
			if isMatched := strings.StartsWith(path, route.Path); isMatched {
				return &route, nil
			}
		case "regex":
			if isMatched, err := regexp.MatchString(route.Path, path); err != nil {
				return nil, err
			} else if isMatched {
				return &route, nil
			}
		default:
			return nil, fmt.Errorf("unsupport path type: %s", route.PathType)
		}
	}

	return nil, ErrRouteNotFound
}
