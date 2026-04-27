package httpcache

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-zoox/proxy"
)

// DefaultProxyOnError mirrors github.com/go-zoox/proxy defaultOnError so we can wrap
// cfg.OnError and skip writing when ErrTerminalResponse is returned from OnRequest.
func DefaultProxyOnError(err error, rw http.ResponseWriter, req *http.Request) {
	status := http.StatusBadGateway
	message := err.Error()

	if errX, ok := err.(*proxy.HTTPError); ok {
		if errX.Status() != 0 {
			status = errX.Status()
		}
	}

	log.Printf("error: %s (%s %s %d)\n", err, req.Method, req.URL.String(), status)

	if strings.Contains(message, "connection refused") {
		status = http.StatusServiceUnavailable
		message = "Service Unavailable"
	}

	rw.WriteHeader(status)
	_, _ = rw.Write([]byte(message))
}
