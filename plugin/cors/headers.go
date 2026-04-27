package cors

import (
	"net/http"
	"strconv"
	"strings"
)

func originOK(c *corsConfig, origin string) bool {
	return allowOriginValue(c, origin) != ""
}

func methodOK(c *corsConfig, m string) bool {
	if c == nil {
		return false
	}
	m = strings.ToUpper(strings.TrimSpace(m))
	if m == "" {
		return true
	}
	for _, x := range c.allowMethods {
		if x == m {
			return true
		}
	}
	return false
}

// allowOriginValue returns the Access-Control-Allow-Origin value to send, or "" if the response
// should not expose CORS for this case.
func allowOriginValue(c *corsConfig, origin string) string {
	if c == nil {
		return ""
	}
	for _, o := range c.allowOrigins {
		if o == "*" && !c.allowCredentials {
			return "*"
		}
	}
	if c.allowCredentials {
		if origin == "" {
			return ""
		}
		for _, o := range c.allowOrigins {
			if strings.EqualFold(o, origin) {
				return origin
			}
		}
		return ""
	}
	if origin == "" {
		return ""
	}
	for _, o := range c.allowOrigins {
		if strings.EqualFold(o, origin) {
			return o
		}
	}
	return ""
}

func applyCORSResponse(h http.Header, c *corsConfig, allowOrigin string, isPreflight bool) {
	if h == nil {
		return
	}
	if allowOrigin != "" {
		h.Set("Access-Control-Allow-Origin", allowOrigin)
	}
	if c.allowCredentials {
		h.Set("Access-Control-Allow-Credentials", "true")
	} else {
		h.Del("Access-Control-Allow-Credentials")
	}
	if isPreflight {
		if len(c.allowMethods) > 0 {
			h.Set("Access-Control-Allow-Methods", strings.Join(c.allowMethods, ", "))
		}
		if len(c.allowHeaders) > 0 {
			h.Set("Access-Control-Allow-Headers", strings.Join(c.allowHeaders, ", "))
		}
	} else {
		if len(c.exposeHeaders) > 0 {
			h.Set("Access-Control-Expose-Headers", strings.Join(c.exposeHeaders, ", "))
		}
	}
}

func itoa64(n int64) string { return strconv.FormatInt(n, 10) }
