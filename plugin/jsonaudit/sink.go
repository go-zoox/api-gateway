package jsonaudit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-zoox/api-gateway/core/route"
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
	}
	return nil
}

func warnUnknownJSONAuditOutput(app *zoox.Application, raw string) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return
	}
	switch strings.ToLower(s) {
	case "console", "stdout", "file", "http", "https", "webhook", "endpoint", "api":
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
	default:
		ctx.Logger.Infof("%s", string(line))
	}
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
