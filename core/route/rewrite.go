package route

import (
	"github.com/go-zoox/core-utils/fmt"
)

func (r *Route) Rewrite(path string) string {
	if !r.Backend.Service.Request.Path.DisablePrefixRewrite {
		if r.PathType != "regex" {
			if r.PathType == "prefix" && r.Path == "/" {
				// home should not rewrite
			} else {
				r.Backend.Service.Request.Path.Rewrites = append(
					r.Backend.Service.Request.Path.Rewrites,
					fmt.Sprintf("^%s(.*)$:$1", r.Path),
				)
			}
		}
	}

	return r.Backend.Service.Rewrite(path)
}
