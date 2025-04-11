package baseuri

import (
	"net/http"
	"plugin"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/core-utils/fmt"
	"github.com/go-zoox/core-utils/strings"
	"github.com/go-zoox/zoox"
)

type BaseURI struct {
	plugin.Plugin
	//
	BaseURI string
}

func (b *BaseURI) Prepare(app *zoox.Application, cfg *config.Config) (err error) {
	app.Logger.Infof("[plugin:baseuri] prepare ...")

	if b.BaseURI == "" {
		return fmt.Errorf("BaseURI is required for baseuri plugin")
	}

	app.Logger.Infof("[plugin:baseuri] baseuri: %s", b.BaseURI)

	app.Use(func(ctx *zoox.Context) {
		if ok := strings.StartsWith(ctx.Request.URL.Path, b.BaseURI); !ok {
			ctx.Logger.Infof("[plugin:baseuri] baseuri is not match: %s", ctx.Request.URL.Path)
			ctx.Error(404, "Not Found")
			return
		}

		ctx.Request.URL.Path = ctx.Request.URL.Path[len(b.BaseURI):]
		ctx.Path = ctx.Request.URL.Path

		ctx.Next()
	})

	app.Logger.Infof("[plugin:baseuri] base uri: %s", b.BaseURI)
	return nil
}
func (b *BaseURI) OnRequest(ctx *zoox.Context, req *http.Request) (err error) {
	return nil
}
func (b *BaseURI) OnResponse(ctx *zoox.Context, res *http.Response) (err error) {
	return nil
}
