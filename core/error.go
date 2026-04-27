package core

import (
	"errors"

	"github.com/go-zoox/api-gateway/router"
)

// ErrRouteNotFound is re-exported from router for stable errors.Is usage.
var ErrRouteNotFound = router.ErrRouteNotFound
var ErrServiceNotFound = errors.New("service not found")
var ErrServiceUnavailable = errors.New("service unavailable")
