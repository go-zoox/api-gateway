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
				// Generate the rewrite rule
				rewriteRule := fmt.Sprintf("^%s(.*)$:$1", r.Path)

				// Check if the rewrite rule already exists to avoid duplicates
				exists := false
				for _, existingRule := range baseConfig.Request.Path.Rewrites {
					if existingRule == rewriteRule {
						exists = true
						break
					}
				}

				// Only append if it doesn't already exist
				if !exists {
					baseConfig.Request.Path.Rewrites = append(
						baseConfig.Request.Path.Rewrites,
						rewriteRule,
					)
				}
			}
		}
	}

	return baseConfig.Rewrite(path)
}
