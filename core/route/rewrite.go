package route

import (
	"github.com/go-zoox/core-utils/fmt"
)

func (r *Route) Rewrite(path string) string {
	// Normalize backend to get base config
	normalizedBackend := r.Backend.Normalize()
	if normalizedBackend == nil {
		return path
	}

	baseConfig := normalizedBackend.BaseConfig

	if !baseConfig.Request.Path.DisablePrefixRewrite {
		if r.PathType != "regex" {
			if r.PathType == "prefix" && r.Path == "/" {
				// home should not rewrite
			} else {
				baseConfig.Request.Path.Rewrites = append(
					baseConfig.Request.Path.Rewrites,
					fmt.Sprintf("^%s(.*)$:$1", r.Path),
				)
			}
		}
	}

	return baseConfig.Rewrite(path)
}
