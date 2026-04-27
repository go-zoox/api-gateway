package httpcache

import (
	"net/http"

	"github.com/go-zoox/zoox/middleware"
)

// AttachTerminalAwareProxyOnError wraps cfg.OnError so cache HITs (ErrTerminalResponse) do not get a second write.
func AttachTerminalAwareProxyOnError(cfg *middleware.ProxyConfig) {
	inner := cfg.OnError
	cfg.OnError = func(err error, rw http.ResponseWriter, req *http.Request) {
		if IsTerminalResponse(err) {
			return
		}
		if inner != nil {
			inner(err, rw, req)
			return
		}
		DefaultProxyOnError(err, rw, req)
	}
}
