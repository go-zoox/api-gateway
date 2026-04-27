package httpcache

import "errors"

// ErrTerminalResponse indicates the proxy must not write again: the handler already
// wrote the full HTTP response (e.g. cache HIT served from plugin.OnRequest).
var ErrTerminalResponse = errors.New("httpcache: terminal response written")

// IsTerminalResponse reports whether err signals that the response body was already sent.
func IsTerminalResponse(err error) bool {
	return errors.Is(err, ErrTerminalResponse)
}
