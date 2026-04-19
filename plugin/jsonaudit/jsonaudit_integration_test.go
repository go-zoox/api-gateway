package jsonaudit

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/defaults"
)

func testZCtx(req *http.Request) *zoox.Context {
	return &zoox.Context{
		Request: req,
		Path:    req.URL.Path,
		Logger:  logger.New(),
	}
}

func TestPrepare_DefaultRedactKeys(t *testing.T) {
	j := New()
	app := defaults.Default()
	cfg := &config.Config{
		JSONAudit: config.JSONAudit{
			Enable:       true,
			MaxBodyBytes: 512,
			SampleRate:   1,
		},
	}
	if err := j.Prepare(app, cfg); err != nil {
		t.Fatal(err)
	}
	keys := buildRedactKeySet(&j.globalConfig)
	if _, ok := keys["password"]; !ok {
		t.Fatal("expected built-in password key")
	}
	if maxBodyFor(&j.globalConfig) != 512 {
		t.Fatalf("maxBody=%d", maxBodyFor(&j.globalConfig))
	}
	if effectiveSampleRateFor(&j.globalConfig) != 1 {
		t.Fatal()
	}
}

func TestPrepare_CustomRedactKeys(t *testing.T) {
	j := New()
	app := defaults.Default()
	cfg := &config.Config{
		JSONAudit: config.JSONAudit{
			Enable:     true,
			RedactKeys: []string{"  CUSTOM  "},
		},
	}
	if err := j.Prepare(app, cfg); err != nil {
		t.Fatal(err)
	}
	keys := buildRedactKeySet(&j.globalConfig)
	if _, ok := keys["custom"]; !ok {
		t.Fatal()
	}
}

func TestEffectiveSampleRate_Clamp(t *testing.T) {
	cfg := &route.JSONAudit{SampleRate: 2}
	if g := effectiveSampleRateFor(cfg); g != 1 {
		t.Fatalf("want 1, got %v", g)
	}
	cfg.SampleRate = -5
	if g := effectiveSampleRateFor(cfg); g != 1 {
		t.Fatalf("want 1 for <=0, got %v", g)
	}
	cfg.SampleRate = 0.25
	if g := effectiveSampleRateFor(cfg); g != 0.25 {
		t.Fatal()
	}
}

func TestMaxBody_DefaultWhenZero(t *testing.T) {
	cfg := &route.JSONAudit{MaxBodyBytes: 0}
	if maxBodyFor(cfg) != 1048576 {
		t.Fatal()
	}
}

func TestPathAllowed(t *testing.T) {
	j := New()
	cfg := &route.JSONAudit{
		IncludePaths: []string{"/api"},
		ExcludePaths: []string{"/health"},
	}
	if !j.pathAllowed("/api/v1", cfg) {
		t.Fatal()
	}
	if j.pathAllowed("/web", cfg) {
		t.Fatal()
	}
	cfg.IncludePaths = nil
	if j.pathAllowed("/health/live", cfg) {
		t.Fatal()
	}
}

func TestSampleHit_AlwaysOne(t *testing.T) {
	j := New()
	cfg := &route.JSONAudit{SampleRate: 1}
	for i := 0; i < 50; i++ {
		if !j.sampleHit(cfg) {
			t.Fatal()
		}
	}
}

func TestOnRequest_Disabled(t *testing.T) {
	j := New()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	ctx := testZCtx(req)
	if err := j.OnRequest(ctx, req); err != nil {
		t.Fatal(err)
	}
	if ctx.Request.Context().Value(auditStateKey) != nil {
		t.Fatal("unexpected audit state")
	}
}

func TestOnRequest_PathExcluded_Skipped(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable:       true,
			ExcludePaths: []string{"/nobody"},
			SampleRate:   1,
		},
	}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/nobody/here", nil)
	ctx := testZCtx(req)
	if err := j.OnRequest(ctx, req); err != nil {
		t.Fatal(err)
	}
	st, _ := ctx.Request.Context().Value(auditStateKey).(*auditState)
	if st == nil || !st.skipped {
		t.Fatal("expected skipped state")
	}
}

func TestOnRequest_IncludeMismatch_Skipped(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable:       true,
			IncludePaths: []string{"/only-this"},
			SampleRate:   1,
		},
	}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/other", nil)
	ctx := testZCtx(req)
	if err := j.OnRequest(ctx, req); err != nil {
		t.Fatal(err)
	}
	st, _ := ctx.Request.Context().Value(auditStateKey).(*auditState)
	if st == nil || !st.skipped {
		t.Fatal()
	}
}

func TestOnRequest_ReadBodyError(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{Enable: true, SampleRate: 1},
	}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api", io.NopCloser(&errReader{err: errors.New("boom")}))
	ctx := testZCtx(req)
	err := j.OnRequest(ctx, req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOnRequest_Success(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{JSONAudit: config.JSONAudit{Enable: true, SampleRate: 1}}); err != nil {
		t.Fatal(err)
	}

	body := bytes.NewBufferString(`{"x":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/a?token=secret&k=1", body)
	req.URL.RawQuery = "token=secret&k=1"
	req.Header.Set("Cookie", "sid=abc")
	req.Header.Set("X-Request-ID", "rid-1")
	req.RemoteAddr = "10.0.0.1:99"
	ctx := testZCtx(req)

	if err := j.OnRequest(ctx, req); err != nil {
		t.Fatal(err)
	}
	st, _ := ctx.Request.Context().Value(auditStateKey).(*auditState)
	if st == nil || st.skipped || !st.eligible {
		t.Fatal("expected eligible snapshot")
	}
	if st.requestID != "rid-1" || st.query["token"][0] != "[REDACTED]" {
		t.Fatalf("state: %+v", st.query)
	}
}

func TestOnResponse_EarlyReturns(t *testing.T) {
	j := New()
	ctx := testZCtx(httptest.NewRequest(http.MethodGet, "/", nil))
	if err := j.OnResponse(ctx, &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{}`))}); err != nil {
		t.Fatal(err)
	}
	if err := j.OnResponse(ctx, nil); err != nil {
		t.Fatal(err)
	}

	if err := j.Prepare(defaults.Default(), &config.Config{JSONAudit: config.JSONAudit{Enable: true}}); err != nil {
		t.Fatal(err)
	}
	if err := j.OnResponse(ctx, jsonResponse(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx2 := testZCtx(req)
	if err := j.OnResponse(ctx2, jsonResponse(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
}

func TestOnResponse_NonJSON_NoLog(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{Enable: true, SniffJSON: false},
	}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	ctx := testZCtx(req)
	ctx.Request = ctx.Request.WithContext(contextWithAudit(req.Context(), &auditState{
		eligible: true,
		method:   "GET",
		path:     "/api",
		headers:  map[string][]string{},
		query:    map[string][]string{},
	}))
	res := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       io.NopCloser(strings.NewReader("not json")),
	}
	if err := j.OnResponse(ctx, res); err != nil {
		t.Fatal(err)
	}
}

func TestOnResponse_EmptyBody_NoAudit(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{Enable: true, SniffJSON: true},
	}); err != nil {
		t.Fatal(err)
	}
	ctx := auditContext(t, &auditState{eligible: true, method: "GET", path: "/p"})
	res := jsonResponse(``)
	if err := j.OnResponse(ctx, res); err != nil {
		t.Fatal(err)
	}
}

func TestOnResponse_GzipJSON(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{Enable: true, DecompressGzip: true, SniffJSON: true},
	}); err != nil {
		t.Fatal(err)
	}

	plain := []byte(`{"gz":true}`)
	gz := gzipCompress(t, plain)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(plain))
	ctx := testZCtx(req)
	st := &auditState{
		eligible: true,
		method:   "POST",
		path:     "/",
		headers:  map[string][]string{},
		query:    map[string][]string{},
		reqBody:  []byte(`{}`),
	}
	ctx.Request = ctx.Request.WithContext(contextWithAudit(ctx.Request.Context(), st))

	res := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type":     []string{"application/json"},
			"Content-Encoding": []string{"gzip"},
		},
		Body: io.NopCloser(bytes.NewReader(gz)),
	}
	if err := j.OnResponse(ctx, res); err != nil {
		t.Fatal(err)
	}
}

func TestOnResponse_InvalidGzip_FallbackRaw(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{Enable: true, DecompressGzip: true, SniffJSON: true},
	}); err != nil {
		t.Fatal(err)
	}

	ctx := auditContext(t, &auditState{eligible: true, method: "GET", path: "/", reqBody: []byte(`{}`)})
	res := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type":     []string{"application/json"},
			"Content-Encoding": []string{"gzip"},
		},
		Body: io.NopCloser(strings.NewReader("not-gzip")),
	}
	if err := j.OnResponse(ctx, res); err != nil {
		t.Fatal(err)
	}
}

func TestOnResponse_FullAuditLine(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable:       true,
			MaxBodyBytes: 4096,
			SampleRate:   1,
			SniffJSON:    true,
		},
	}); err != nil {
		t.Fatal(err)
	}

	reqBody := `{"user":"a","password":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ua-test")
	req.Header.Set("Authorization", "Bearer z")
	req.RemoteAddr = "192.0.2.1:4444"

	ctx := testZCtx(req)
	if err := j.OnRequest(ctx, req); err != nil {
		t.Fatal(err)
	}

	respJSON := `{"ok":true,"token":"secret"}`
	res := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(respJSON)),
	}
	if err := j.OnResponse(ctx, res); err != nil {
		t.Fatal(err)
	}

	// ensure client can still read response body
	out, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(out), "secret") {
		t.Fatal("response body must be restored for downstream")
	}
}

func TestGunzipLimited_Truncate(t *testing.T) {
	large := bytes.Repeat([]byte("a"), 5000)
	gz := gzipCompress(t, large)
	out, err := gunzipLimited(gz, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 100 {
		t.Fatalf("len=%d", len(out))
	}
}

func TestGunzipLimited_Invalid(t *testing.T) {
	_, err := gunzipLimited([]byte{1, 2, 3}, 100)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadBodyLimited_NilBody(t *testing.T) {
	b, trunc, err := readBodyLimited(nil, 10)
	if err != nil || trunc || b != nil {
		t.Fatal()
	}
}

func TestReadBodyLimited_Error(t *testing.T) {
	_, _, err := readBodyLimited(io.NopCloser(&errReader{err: io.ErrUnexpectedEOF}), 10)
	if err == nil {
		t.Fatal()
	}
}

func TestCloneHeadersRedacted_Empty(t *testing.T) {
	out := cloneHeadersRedacted(http.Header{})
	if len(out) != 0 {
		t.Fatal()
	}
}

func TestRedactQueryValues_Empty(t *testing.T) {
	out := redactQueryValues(url.Values{}, nil)
	if len(out) != 0 {
		t.Fatal()
	}
}

func TestSnapshotParams_NilContext(t *testing.T) {
	if len(snapshotParams(nil)) != 0 {
		t.Fatal()
	}
}

func TestSnapshotParams_Zoox(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := testZCtx(req)
	m := snapshotParams(ctx)
	if m == nil {
		t.Fatal()
	}
}

func TestRedactJSONBytes_Invalid(t *testing.T) {
	cfg := &route.JSONAudit{Enable: true}
	keys := buildRedactKeySet(cfg)
	if redactJSONBytes([]byte(`not json`), keys) != nil {
		t.Fatal()
	}
}

func TestRequestBodyForLog_RawFallback(t *testing.T) {
	j := New()
	if err := j.Prepare(defaults.Default(), &config.Config{JSONAudit: config.JSONAudit{Enable: true}}); err != nil {
		t.Fatal(err)
	}
	keys := buildRedactKeySet(&j.globalConfig)
	v := j.requestBodyForLog([]byte(`<<<not-json>>>`), keys)
	if s, ok := v.(string); !ok || s != `<<<not-json>>>` {
		t.Fatalf("%#v", v)
	}
}

func TestRedactValue_ArrayPrimitives(t *testing.T) {
	keys := map[string]struct{}{"x": {}}
	v := []any{
		map[string]any{"x": "hide"},
		"plain",
	}
	out := redactValue(v, keys)
	arr, ok := out.([]any)
	if !ok || arr[1] != "plain" {
		t.Fatal()
	}
}

func TestHasPrefixAny_EmptyPrefixSkipped(t *testing.T) {
	if !hasPrefixAny("/x", []string{"", " /", "/x"}) {
		t.Fatal()
	}
}

func TestFirstHeader_SecondKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "c1")
	if firstHeader(req, "X-Request-ID", "X-Correlation-ID") != "c1" {
		t.Fatal()
	}
}

func TestResponseLooksJSON_EmptyBody(t *testing.T) {
	j := New()
	cfg := &route.JSONAudit{}
	if j.responseLooksJSON(cfg, "application/json", nil) {
		t.Fatal()
	}
}

func TestJSONMIME_PlusJSONSuffix(t *testing.T) {
	if !jsonMIME("application/vnd.api+json") {
		t.Fatal()
	}
}

// Helpers

type errReader struct {
	err error
}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, e.err
}

func gzipCompress(t *testing.T, plain []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func contextWithAudit(parent context.Context, st *auditState) context.Context {
	return context.WithValue(parent, auditStateKey, st)
}

func auditContext(t *testing.T, st *auditState) *zoox.Context {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, st.path, nil)
	ctx := testZCtx(req)
	ctx.Request = ctx.Request.WithContext(contextWithAudit(ctx.Request.Context(), st))
	return ctx
}
