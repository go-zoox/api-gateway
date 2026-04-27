package httpcache

import (
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// normalizedQueryString parses the raw query, sorts parameter names and each name's
// values, then re-encodes so semantically equivalent query strings map to the same bytes.
// Empty input yields empty output.
func normalizedQueryString(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	v, err := url.ParseQuery(rawQuery)
	if err != nil {
		// If parsing fails, fall back to raw (still deterministic for identical strings).
		return rawQuery
	}
	if len(v) == 0 {
		return ""
	}
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		vals := append([]string(nil), v[k]...)
		sort.Strings(vals)
		for _, val := range vals {
			if b.Len() > 0 {
				b.WriteByte('&')
			}
			b.WriteString(url.QueryEscape(k))
			b.WriteByte('=')
			b.WriteString(url.QueryEscape(val))
		}
	}
	return b.String()
}

// sortedHeaderField joins all values for a header field in sorted order, independent of
// the order those header lines were received (multi-value / duplicate headers).
func sortedHeaderField(h http.Header, name string) string {
	canon := http.CanonicalHeaderKey(name)
	vals, ok := h[canon]
	if !ok && canon != name {
		vals = h[name]
	}
	if len(vals) == 0 {
		return ""
	}
	cpy := append([]string(nil), vals...)
	sort.Strings(cpy)
	return strings.Join(cpy, "\x1e") // record separator between sorted lines
}
