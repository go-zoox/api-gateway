package jsonaudit

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/gormx"
	"github.com/go-zoox/zoox"
)

func validateJSONAuditSink(where string, ja *route.JSONAudit) error {
	if ja == nil || !ja.Enable {
		return nil
	}
	switch route.EffectiveJSONAuditProvider(ja.Output) {
	case "file":
		if strings.TrimSpace(ja.Output.File.Path) == "" {
			return fmt.Errorf("json_audit %s: output.provider=file requires output.file.path", where)
		}
	case "http":
		if strings.TrimSpace(ja.Output.HTTP.URL) == "" {
			return fmt.Errorf("json_audit %s: output.provider=http requires output.http.url", where)
		}
	case "database":
		if _, _, err := resolveJSONAuditDatabaseConfig(ja.Output.Database); err != nil {
			return fmt.Errorf("json_audit %s: %w", where, err)
		}
	}
	return nil
}

func warnUnknownJSONAuditOutput(app *zoox.Application, raw string) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return
	}
	switch strings.ToLower(s) {
	case "console", "stdout", "file", "http", "https", "webhook", "endpoint", "api", "database", "db", "sql":
		return
	default:
		app.Logger().Warnf("[plugin:jsonaudit] unknown json_audit.output.provider %q, using console", raw)
	}
}

func (j *JSONAudit) emitAuditLine(ctx *zoox.Context, cfg *route.JSONAudit, line []byte) {
	switch route.EffectiveJSONAuditProvider(cfg.Output) {
	case "file":
		path := strings.TrimSpace(cfg.Output.File.Path)
		if path == "" {
			ctx.Logger.Warnf("[plugin:jsonaudit] output.provider=file but output.file.path empty, falling back to console")
			ctx.Logger.Infof("%s", string(line))
			return
		}
		if err := j.appendFileLine(path, line); err != nil {
			ctx.Logger.Warnf("[plugin:jsonaudit] write audit file %s: %v", path, err)
			ctx.Logger.Infof("%s", string(line))
		}
	case "http":
		if err := j.postAuditHTTP(cfg, line); err != nil {
			ctx.Logger.Warnf("[plugin:jsonaudit] http audit sink: %v", err)
			ctx.Logger.Infof("%s", string(line))
		}
	case "database":
		if err := j.insertAuditRecord(line); err != nil {
			ctx.Logger.Warnf("[plugin:jsonaudit] database audit sink: %v", err)
			ctx.Logger.Infof("%s", string(line))
		}
	default:
		ctx.Logger.Infof("%s", string(line))
	}
}

func (j *JSONAudit) prepareDatabaseSink(app *zoox.Application, cfg *config.Config) error {
	dbSinks := collectDatabaseSinks(cfg)
	if len(dbSinks) == 0 {
		return nil
	}
	first := dbSinks[0]
	engine, dsn, err := resolveJSONAuditDatabaseConfig(first.cfg)
	if err != nil {
		return fmt.Errorf("json_audit %s: %w", first.where, err)
	}
	for _, sink := range dbSinks[1:] {
		nextEngine, nextDSN, err := resolveJSONAuditDatabaseConfig(sink.cfg)
		if err != nil {
			return fmt.Errorf("json_audit %s: %w", sink.where, err)
		}
		if nextEngine != engine || nextDSN != dsn {
			return fmt.Errorf("json_audit %s: multiple database sinks detected; use one shared output.database config", sink.where)
		}
	}
	if err := gormx.LoadDB(engine, dsn); err != nil {
		return fmt.Errorf("json_audit %s: connect database failed: %w", first.where, err)
	}
	db := gormx.GetDB()
	if db == nil {
		return fmt.Errorf("json_audit %s: gormx returned nil db", first.where)
	}
	if err := db.AutoMigrate(&jsonAuditRecord{}); err != nil {
		return fmt.Errorf("json_audit %s: migrate failed: %w", first.where, err)
	}
	j.dbMu.Lock()
	j.db = db
	j.dbMu.Unlock()
	app.Logger().Infof("[plugin:jsonaudit] database sink ready (engine=%s, source=%s)", engine, first.where)
	return nil
}

func (j *JSONAudit) appendFileLine(path string, line []byte) error {
	j.fileMu.Lock()
	defer j.fileMu.Unlock()
	if j.fileHandles == nil {
		j.fileHandles = make(map[string]*os.File)
	}
	f, ok := j.fileHandles[path]
	var err error
	if !ok {
		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		j.fileHandles[path] = f
	}
	if !bytes.HasSuffix(line, []byte("\n")) {
		line = append(line, '\n')
	}
	_, err = f.Write(line)
	return err
}

func (j *JSONAudit) postAuditHTTP(cfg *route.JSONAudit, line []byte) error {
	u := strings.TrimSpace(cfg.Output.HTTP.URL)
	if u == "" {
		return fmt.Errorf("output.http.url is empty")
	}
	method := strings.ToUpper(strings.TrimSpace(cfg.Output.HTTP.Method))
	if method == "" {
		method = http.MethodPost
	}
	sec := cfg.Output.HTTP.TimeoutSeconds
	if sec <= 0 {
		sec = 5
	}
	reqCtx, cancel := context.WithTimeout(context.Background(), time.Duration(sec)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, method, u, bytes.NewReader(line))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Output.HTTP.Headers {
		if strings.TrimSpace(k) != "" {
			req.Header.Set(k, v)
		}
	}

	if j.httpClient == nil {
		j.httpClient = &http.Client{}
	}
	resp, err := j.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}
	return nil
}

func (j *JSONAudit) insertAuditRecord(line []byte) error {
	j.dbMu.Lock()
	db := j.db
	j.dbMu.Unlock()
	if db == nil {
		return fmt.Errorf("database sink not initialized")
	}

	var payload struct {
		Type              string          `json:"type"`
		Time              string          `json:"time"`
		Timestamp         int64           `json:"timestamp"`
		Method            string          `json:"method"`
		Path              string          `json:"path"`
		RemoteAddr        string          `json:"remote_addr"`
		RequestID         string          `json:"request_id"`
		UserAgent         string          `json:"user_agent"`
		ResponseStatus    int             `json:"response_status"`
		ContentType       string          `json:"content_type"`
		RequestTruncated  bool            `json:"request_truncated"`
		ResponseTruncated bool            `json:"response_truncated"`
		Request           json.RawMessage `json:"request"`
		Response          json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(line, &payload); err != nil {
		return fmt.Errorf("invalid json audit record: %w", err)
	}
	if strings.TrimSpace(payload.Type) == "" {
		payload.Type = "json_audit"
	}

	request := string(payload.Request)
	response := string(payload.Response)
	username, password, token, clientID, clientSecret, authorization, xAPIKey := extractAuditAuthFields(payload.Request)

	return db.Create(&jsonAuditRecord{
		Type:              payload.Type,
		Time:              payload.Time,
		Timestamp:         payload.Timestamp,
		Method:            payload.Method,
		Path:              payload.Path,
		RemoteAddr:        payload.RemoteAddr,
		RequestID:         payload.RequestID,
		UserAgent:         payload.UserAgent,
		ResponseStatus:    payload.ResponseStatus,
		ContentType:       payload.ContentType,
		RequestTruncated:  payload.RequestTruncated,
		ResponseTruncated: payload.ResponseTruncated,
		Username:          username,
		Password:          password,
		Token:             token,
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		Authorization:     authorization,
		XAPIKey:           xAPIKey,
		Request:           request,
		Response:          response,
	}).Error
}

func extractAuditAuthFields(requestRaw json.RawMessage) (username string, password string, token string, clientID string, clientSecret string, authorization string, xAPIKey string) {
	if len(requestRaw) == 0 {
		return "", "", "", "", "", "", ""
	}
	var req struct {
		Headers map[string][]string `json:"headers"`
		Query   map[string][]string `json:"query"`
	}
	if err := json.Unmarshal(requestRaw, &req); err != nil {
		return "", "", "", "", "", "", ""
	}

	auth := firstHeaderValue(req.Headers, "Authorization")
	authorization = auth
	if auth != "" {
		scheme, credential := parseAuthorizationHeader(auth)
		switch scheme {
		case "basic":
			if decoded, err := base64.StdEncoding.DecodeString(credential); err == nil {
				user, pass, found := strings.Cut(string(decoded), ":")
				if found {
					username = user
					password = pass
				} else {
					username = string(decoded)
				}
			}
		case "bearer":
			token = credential
		}
	}
	xAPIKey = firstHeaderValue(req.Headers, "X-API-Key")

	if v := firstHeaderValue(req.Headers, "X-Client-ID"); v != "" {
		clientID = v
	} else if v, ok := findFirstQueryValue(req.Query, "client_id"); ok {
		clientID = v
	}
	if v := firstHeaderValue(req.Headers, "X-Client-Secret"); v != "" {
		clientSecret = v
	} else if v, ok := findFirstQueryValue(req.Query, "client_secret"); ok {
		clientSecret = v
	}

	return username, password, token, clientID, clientSecret, authorization, xAPIKey
}

func parseAuthorizationHeader(raw string) (scheme string, credential string) {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) < 2 {
		return "", ""
	}
	return strings.ToLower(parts[0]), strings.Join(parts[1:], " ")
}

func firstHeaderValue(headers map[string][]string, name string) string {
	if len(headers) == 0 {
		return ""
	}
	lowerName := strings.ToLower(strings.TrimSpace(name))
	for k, values := range headers {
		if strings.ToLower(strings.TrimSpace(k)) != lowerName || len(values) == 0 {
			continue
		}
		for _, v := range values {
			v = strings.TrimSpace(v)
			if v != "" {
				return v
			}
		}
	}
	return ""
}

func findFirstQueryValue(query map[string][]string, key string) (string, bool) {
	if len(query) == 0 {
		return "", false
	}
	for k, values := range query {
		if !sameAuditKey(k, key) || len(values) == 0 {
			continue
		}
		for _, v := range values {
			v = strings.TrimSpace(v)
			if v != "" {
				return v, true
			}
		}
	}
	return "", false
}

func findStringByKey(v any, targetKey string) (string, bool) {
	switch x := v.(type) {
	case map[string]any:
		for k, item := range x {
			if sameAuditKey(k, targetKey) {
				if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
					return strings.TrimSpace(s), true
				}
			}
			if out, ok := findStringByKey(item, targetKey); ok {
				return out, true
			}
		}
	case []any:
		for _, item := range x {
			if out, ok := findStringByKey(item, targetKey); ok {
				return out, true
			}
		}
	}
	return "", false
}

func sameAuditKey(a string, b string) bool {
	normalize := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, "-", "_")
		return s
	}
	return normalize(a) == normalize(b)
}

type databaseSinkConfig struct {
	cfg   route.JSONAuditOutputDatabase
	where string
}

func collectDatabaseSinks(cfg *config.Config) []databaseSinkConfig {
	var sinks []databaseSinkConfig
	if cfg.JSONAudit.Enable && route.EffectiveJSONAuditProvider(cfg.JSONAudit.Output) == "database" {
		sinks = append(sinks, databaseSinkConfig{
			cfg:   cfg.JSONAudit.Output.Database,
			where: "global",
		})
	}
	for i, rt := range cfg.Routes {
		if !rt.JSONAudit.Enable || route.EffectiveJSONAuditProvider(rt.JSONAudit.Output) != "database" {
			continue
		}
		sinks = append(sinks, databaseSinkConfig{
			cfg:   rt.JSONAudit.Output.Database,
			where: fmt.Sprintf("routes[%d] path=%s", i, rt.Path),
		})
	}
	return sinks
}

func normalizeJSONAuditDatabaseEngine(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "postgres", "postgresql", "pg":
		return "postgres"
	case "mysql":
		return "mysql"
	case "sqlite", "sqlite3":
		return "sqlite"
	default:
		return ""
	}
}

func resolveJSONAuditDatabaseConfig(cfg route.JSONAuditOutputDatabase) (engine string, dsn string, err error) {
	if !hasAnyDatabaseConfig(cfg) {
		panic("json_audit output.provider=database requires dedicated output.database config")
	}

	if hasStructuredDatabaseConfig(cfg) {
		engine = normalizeJSONAuditDatabaseEngine(cfg.Engine)
		if engine == "" {
			return "", "", fmt.Errorf("output.provider=database requires output.database.engine (postgres/mysql/sqlite) when using host/port/db fields")
		}
		return buildDatabaseDSNFromFields(engine, cfg)
	}

	dsn = strings.TrimSpace(cfg.DSN)
	if dsn == "" {
		return "", "", fmt.Errorf("output.provider=database requires output.database.dsn or output.database host/port/db fields")
	}

	engine = normalizeJSONAuditDatabaseEngine(cfg.Engine)
	dsnEngine := inferJSONAuditDatabaseEngineFromDSN(dsn)
	if engine == "" {
		engine = dsnEngine
	}
	if engine == "" {
		return "", "", fmt.Errorf("output.provider=database requires output.database.engine (postgres/mysql/sqlite) or a URL dsn like postgres://... or mysql://...")
	}
	if dsnEngine != "" && dsnEngine != engine {
		return "", "", fmt.Errorf("output.database.engine %q does not match dsn scheme", cfg.Engine)
	}

	switch engine {
	case "mysql":
		normalizedDSN, convErr := normalizeMySQLDSN(dsn)
		if convErr != nil {
			return "", "", convErr
		}
		dsn = normalizedDSN
	case "sqlite":
		normalizedDSN, convErr := normalizeSQLiteDSN(dsn)
		if convErr != nil {
			return "", "", convErr
		}
		dsn = normalizedDSN
	}

	return engine, dsn, nil
}

func hasAnyDatabaseConfig(cfg route.JSONAuditOutputDatabase) bool {
	return strings.TrimSpace(cfg.Engine) != "" || strings.TrimSpace(cfg.DSN) != "" || hasStructuredDatabaseConfig(cfg)
}

func hasStructuredDatabaseConfig(cfg route.JSONAuditOutputDatabase) bool {
	return strings.TrimSpace(cfg.Host) != "" || cfg.Port > 0 || strings.TrimSpace(cfg.Username) != "" || strings.TrimSpace(cfg.Password) != "" || strings.TrimSpace(cfg.DB) != ""
}

func buildDatabaseDSNFromFields(engine string, cfg route.JSONAuditOutputDatabase) (string, string, error) {
	host := strings.TrimSpace(cfg.Host)
	port := cfg.Port
	if port <= 0 {
		switch engine {
		case "postgres":
			port = 5432
		case "mysql":
			port = 3306
		}
	}
	db := strings.TrimSpace(cfg.DB)

	switch engine {
	case "postgres":
		if host == "" || db == "" {
			return "", "", fmt.Errorf("postgres host/db is required when using output.database host/port/db fields")
		}
		return "postgres", fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, cfg.Username, cfg.Password, db, port), nil
	case "mysql":
		if host == "" || db == "" {
			return "", "", fmt.Errorf("mysql host/db is required when using output.database host/port/db fields")
		}
		return "mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Username, cfg.Password, host, port, db), nil
	case "sqlite":
		if db == "" {
			return "", "", fmt.Errorf("sqlite db is required when using output.database host/port/db fields")
		}
		return "sqlite", db, nil
	default:
		return "", "", fmt.Errorf("unsupported database engine %q", engine)
	}
}

func inferJSONAuditDatabaseEngineFromDSN(rawDSN string) string {
	if !strings.Contains(rawDSN, "://") {
		return ""
	}
	u, err := url.Parse(rawDSN)
	if err != nil {
		return ""
	}
	return normalizeJSONAuditDatabaseEngine(u.Scheme)
}

func normalizeMySQLDSN(rawDSN string) (string, error) {
	if !strings.HasPrefix(strings.ToLower(rawDSN), "mysql://") {
		return rawDSN, nil
	}
	u, err := url.Parse(rawDSN)
	if err != nil {
		return "", fmt.Errorf("invalid mysql dsn: %w", err)
	}
	host := u.Host
	if strings.TrimSpace(host) == "" {
		return "", fmt.Errorf("mysql dsn missing host")
	}
	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		return "", fmt.Errorf("mysql dsn missing database name in path")
	}
	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	var auth string
	switch {
	case username == "" && password == "":
		auth = ""
	case password == "":
		auth = fmt.Sprintf("%s@", username)
	default:
		auth = fmt.Sprintf("%s:%s@", username, password)
	}
	out := fmt.Sprintf("%stcp(%s)/%s", auth, host, database)
	if u.RawQuery != "" {
		out = out + "?" + u.RawQuery
	}
	return out, nil
}

func normalizeSQLiteDSN(rawDSN string) (string, error) {
	if !strings.Contains(rawDSN, "://") {
		return rawDSN, nil
	}
	u, err := url.Parse(rawDSN)
	if err != nil {
		return "", fmt.Errorf("invalid sqlite dsn: %w", err)
	}
	if normalizeJSONAuditDatabaseEngine(u.Scheme) != "sqlite" {
		return rawDSN, nil
	}
	path := strings.TrimSpace(u.Path)
	if path == "" || path == "/" {
		return "", fmt.Errorf("sqlite dsn missing file path")
	}
	if u.Host != "" && u.Host != "localhost" {
		path = strings.TrimSpace(u.Host + path)
	}
	if u.RawQuery != "" {
		path = path + "?" + u.RawQuery
	}
	return path, nil
}

type jsonAuditRecord struct {
	ID                uint      `gorm:"primaryKey"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	Type              string    `gorm:"type:varchar(64);not null;index"`
	Time              string    `gorm:"type:varchar(64)"`
	Timestamp         int64     `gorm:"index"`
	Method            string    `gorm:"type:varchar(16);index"`
	Path              string    `gorm:"type:text"`
	RemoteAddr        string    `gorm:"type:varchar(255)"`
	RequestID         string    `gorm:"type:varchar(255);index"`
	UserAgent         string    `gorm:"type:text"`
	ResponseStatus    int       `gorm:"index"`
	ContentType       string    `gorm:"type:varchar(255)"`
	RequestTruncated  bool
	ResponseTruncated bool
	Username          string `gorm:"type:varchar(255);index"`
	Password          string `gorm:"type:text"`
	Token             string `gorm:"type:text"`
	ClientID          string `gorm:"type:varchar(255);index"`
	ClientSecret      string `gorm:"type:text"`
	Authorization     string `gorm:"type:text"`
	XAPIKey           string `gorm:"type:text"`
	Request           string `gorm:"type:text"`
	Response          string `gorm:"type:text"`
}

func (jsonAuditRecord) TableName() string {
	return "json_audit_records"
}
