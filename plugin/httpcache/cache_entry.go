package httpcache

// cachedEntry is persisted in Application.Cache() (JSON-compatible).
type cachedEntry struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
}
