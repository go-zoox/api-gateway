package jsonaudit

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/plugin"
	"github.com/go-zoox/zoox"
)

type ctxKey int

const auditStateKey ctxKey = iota

// auditState is stored on the incoming request context between OnRequest and OnResponse.
type auditState struct {
	skipped bool

	method       string
	path         string
	remoteAddr   string
	requestID    string
	userAgent    string
	headers      map[string][]string
	query        map[string][]string
	reqBody      []byte
	reqTruncated bool

	// True when we should still emit audit if response is JSON-like (path + sample passed).
	eligible bool
}

var sensitiveHeaderNames = map[string]struct{}{
	"authorization": {}, "cookie": {}, "set-cookie": {},
	"x-api-key": {}, "proxy-authorization": {},
}

// JSONAudit logs request/response bodies for audit when the upstream response is JSON-like.
type JSONAudit struct {
	plugin.Plugin

	cfg config.JSONAudit

	defaultRedact map[string]struct{}
}

// New builds a JSON audit plugin (call Prepare after construction via core).
func New() *JSONAudit {
	return &JSONAudit{}
}

// Prepare validates options and caches defaults.
func (j *JSONAudit) Prepare(app *zoox.Application, cfg *config.Config) error {
	j.cfg = cfg.JSONAudit

	j.defaultRedact = map[string]struct{}{}
	keys := j.cfg.RedactKeys
	if len(keys) == 0 {
		keys = []string{
			"password", "passwd", "secret", "token", "authorization",
			"api_key", "apikey", "access_token", "refresh_token",
		}
	}
	for _, k := range keys {
		j.defaultRedact[strings.ToLower(strings.TrimSpace(k))] = struct{}{}
	}

	app.Logger().Infof("[plugin:jsonaudit] prepare (enable=%v, max_body_bytes=%d, sample_rate=%g)",
		j.cfg.Enable, j.maxBody(), j.effectiveSampleRate())

	return nil
}

func (j *JSONAudit) effectiveSampleRate() float64 {
	r := j.cfg.SampleRate
	if r <= 0 {
		return 1
	}
	if r > 1 {
		return 1
	}
	return r
}

func (j *JSONAudit) maxBody() int64 {
	max := j.cfg.MaxBodyBytes
	if max <= 0 {
		return 1048576
	}
	return max
}

// OnRequest snapshots the client request body (bounded) for pairing with JSON responses.
func (j *JSONAudit) OnRequest(ctx *zoox.Context, _ *http.Request) error {
	if !j.cfg.Enable {
		return nil
	}

	path := ctx.Path
	if !j.pathAllowed(path) || !j.sampleHit() {
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), auditStateKey, &auditState{skipped: true}))
		return nil
	}

	q := ctx.Request.URL.Query()
	st := &auditState{
		eligible:   true,
		method:     ctx.Request.Method,
		path:       path,
		remoteAddr: ctx.Request.RemoteAddr,
		requestID:  firstHeader(ctx.Request, "X-Request-ID", "X-Correlation-ID", "X-Trace-ID"),
		userAgent:  ctx.Request.UserAgent(),
		headers:    cloneHeadersRedacted(ctx.Request.Header),
		query:      redactQueryValues(q, j.defaultRedact),
	}

	raw, truncated, err := readBodyLimited(ctx.Request.Body, j.maxBody())
	if err != nil {
		return err
	}
	ctx.Request.Body = io.NopCloser(bytes.NewReader(raw))
	st.reqBody = raw
	st.reqTruncated = truncated

	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), auditStateKey, st))
	return nil
}

// OnResponse emits one audit log line when the upstream response looks like JSON.
func (j *JSONAudit) OnResponse(ctx *zoox.Context, res *http.Response) error {
	if !j.cfg.Enable || res == nil {
		return nil
	}

	st, _ := ctx.Request.Context().Value(auditStateKey).(*auditState)
	if st == nil || st.skipped || !st.eligible {
		return nil
	}

	raw, truncated, err := readBodyLimited(res.Body, j.maxBody())
	if err != nil {
		return err
	}
	res.Body = io.NopCloser(bytes.NewReader(raw))

	encoding := res.Header.Get("Content-Encoding")
	bodyForDetect := raw
	if j.cfg.DecompressGzip && strings.Contains(strings.ToLower(encoding), "gzip") {
		if dec, err := gunzipLimited(raw, j.maxBody()); err == nil {
			bodyForDetect = dec
		}
	}

	if !j.responseLooksJSON(res.Header.Get("Content-Type"), bodyForDetect) {
		return nil
	}

	reqBodyVal := j.requestBodyForLog(st.reqBody)
	respBodyVal := j.requestBodyForLog(bodyForDetect)

	reqObj := map[string]any{
		"method":  st.method,
		"path":    st.path,
		"headers": st.headers,
		"query":   st.query,
		"params":  snapshotParams(ctx),
		"body":    reqBodyVal,
	}

	respObj := map[string]any{
		"status": res.StatusCode,
		"body":   respBodyVal,
	}

	now := time.Now().UTC()
	rec := map[string]any{
		"type":               "json_audit",
		"time":               now.Format(time.RFC3339Nano),
		"timestamp":          now.UnixMilli(),
		"method":             st.method,
		"path":               st.path,
		"remote_addr":        st.remoteAddr,
		"request_id":         st.requestID,
		"user_agent":         st.userAgent,
		"response_status":    res.StatusCode,
		"content_type":       res.Header.Get("Content-Type"),
		"request_truncated":  st.reqTruncated,
		"response_truncated": truncated,
		"request":            reqObj,
		"response":           respObj,
	}

	line, err := json.Marshal(rec)
	if err != nil {
		ctx.Logger.Warnf("[plugin:jsonaudit] marshal audit record: %v", err)
		return nil
	}

	ctx.Logger.Infof("%s", string(line))
	return nil
}

func (j *JSONAudit) pathAllowed(path string) bool {
	if len(j.cfg.IncludePaths) > 0 {
		if !hasPrefixAny(path, j.cfg.IncludePaths) {
			return false
		}
	}
	if hasPrefixAny(path, j.cfg.ExcludePaths) {
		return false
	}
	return true
}

func (j *JSONAudit) sampleHit() bool {
	rate := j.effectiveSampleRate()
	if rate >= 1 {
		return true
	}
	if rate <= 0 {
		return false
	}
	return rand.Float64() < rate
}

func (j *JSONAudit) responseLooksJSON(ct string, body []byte) bool {
	if len(body) == 0 {
		return false
	}
	if jsonMIME(ct) {
		return true
	}
	if !j.cfg.SniffJSON {
		return false
	}
	return json.Valid(bytes.TrimSpace(body))
}

func (j *JSONAudit) redactJSONBytes(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return redactValue(v, j.defaultRedact)
}

func (j *JSONAudit) requestBodyForLog(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	if v := j.redactJSONBytes(raw); v != nil {
		return v
	}
	return string(raw)
}

func cloneHeadersRedacted(h http.Header) map[string][]string {
	if len(h) == 0 {
		return map[string][]string{}
	}
	out := make(map[string][]string)
	for k, vals := range h {
		lk := strings.ToLower(k)
		if _, ok := sensitiveHeaderNames[lk]; ok {
			out[k] = []string{"[REDACTED]"}
			continue
		}
		cp := make([]string, len(vals))
		copy(cp, vals)
		out[k] = cp
	}
	return out
}

func redactQueryValues(q url.Values, keys map[string]struct{}) map[string][]string {
	if len(q) == 0 {
		return map[string][]string{}
	}
	out := make(map[string][]string)
	for k, vals := range q {
		lk := strings.ToLower(k)
		if _, ok := keys[lk]; ok {
			out[k] = []string{"[REDACTED]"}
			continue
		}
		out[k] = append([]string(nil), vals...)
	}
	return out
}

func snapshotParams(ctx *zoox.Context) map[string]any {
	if ctx == nil {
		return map[string]any{}
	}
	pm := ctx.Params()
	if pm == nil {
		return map[string]any{}
	}
	m := pm.ToMap()
	if m == nil {
		return map[string]any{}
	}
	return m
}

func redactValue(v any, keys map[string]struct{}) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			lk := strings.ToLower(k)
			if _, ok := keys[lk]; ok {
				out[k] = "[REDACTED]"
				continue
			}
			out[k] = redactValue(val, keys)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = redactValue(val, keys)
		}
		return out
	default:
		return v
	}
}

func jsonMIME(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(ct))
	if ct == "" {
		return false
	}
	if strings.Contains(ct, "json") {
		return true
	}
	// e.g. application/vnd.api+json
	return strings.HasSuffix(ct, "+json")
}

func readBodyLimited(body io.ReadCloser, max int64) ([]byte, bool, error) {
	if body == nil {
		return nil, false, nil
	}
	defer body.Close()

	lr := io.LimitReader(body, max+1)
	b, err := io.ReadAll(lr)
	if err != nil {
		return nil, false, err
	}
	truncated := int64(len(b)) > max
	if truncated && max > 0 {
		b = b[:max]
	}
	return b, truncated, nil
}

func gunzipLimited(raw []byte, max int64) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	lr := io.LimitReader(gr, max+1)
	out, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(out)) > max {
		return out[:max], nil
	}
	return out, nil
}

func firstHeader(req *http.Request, names ...string) string {
	for _, n := range names {
		if v := req.Header.Get(n); v != "" {
			return v
		}
	}
	return ""
}

func hasPrefixAny(path string, prefixes []string) bool {
	for _, p := range prefixes {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
