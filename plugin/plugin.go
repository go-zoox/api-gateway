package plugin

import (
	"net/http"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/zoox"
)

type Plugin interface {
	// prepare
	Prepare(app *zoox.Application, cfg *config.Config) (err error)

	// request
	OnRequest(ctx *zoox.Context, req *http.Request) (err error)

	// response
	OnResponse(ctx *zoox.Context, res *http.Response) (err error)
}
