package core

import (
	"errors"
	"time"

	"github.com/go-zoox/core-utils/fmt"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
	"github.com/go-zoox/api-gateway/router"
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

	s, err = router.RouteForPath(c.cfg, path)
	if err != nil {
		if errors.Is(err, ErrRouteNotFound) {
			return nil, err
		}
		return nil, err
	}

	ctx.Cache().Set(key, s, 60*time.Second)

	return s, nil
}

// MatchPath delegates to [router.MatchPath] for the same path matching rules as the gateway.
func MatchPath(routes []route.Route, path string) (*route.Route, error) {
	return router.MatchPath(routes, path)
}
