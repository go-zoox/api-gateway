package core

import "errors"

var ErrRouteNotFound = errors.New("route not found")
var ErrServiceNotFound = errors.New("service not found")
var ErrServiceUnavailable = errors.New("service unavailable")
