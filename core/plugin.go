package core

import (
	"github.com/go-zoox/api-gateway/plugin"
)

func (c *core) Plugin(plugin ...plugin.Plugin) Core {
	c.plugins = append(c.plugins, plugin...)
	return c
}
