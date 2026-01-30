package core

import (
	"github.com/go-zoox/api-gateway/plugin/baseuri"
	"github.com/go-zoox/kv"
	"github.com/go-zoox/kv/redis"
)

func (c *core) prepare() error {
	// prepare cache
	if err := c.prepareCache(); err != nil {
		return err
	}

	// prepare plugins
	if err := c.preparePlugins(); err != nil {
		return err
	}

	return nil
}

func (c *core) prepareCache() error {
	if c.cfg.Cache.Host != "" {
		prefix := c.cfg.Cache.Prefix
		if prefix == "" {
			prefix = "gozoox-api-gateway:"
		}

		c.app.Config.Cache = kv.Config{
			Engine: "redis",
			Config: &redis.Config{
				Host:     c.cfg.Cache.Host,
				Port:     int(c.cfg.Cache.Port),
				Username: c.cfg.Cache.Username,
				Password: c.cfg.Cache.Password,
				DB:       int(c.cfg.Cache.DB),
				Prefix:   prefix,
			},
		}
	}

	return nil
}

func (c *core) preparePlugins() error {
	// buildin plugins
	if err := c.preparePluginsBuildin(); err != nil {
		return err
	}

	for _, plugin := range c.plugins {
		if err := plugin.Prepare(c.app, c.cfg); err != nil {
			return err
		}
	}

	return nil
}

func (c *core) preparePluginsBuildin() error {
	c.app.Logger().Debugf("baseuri: %s", c.cfg.BaseURI)

	// baseuri
	if c.cfg.BaseURI != "" {
		c.plugins = append(c.plugins, &baseuri.BaseURI{
			BaseURI: c.cfg.BaseURI,
		})
	}

	return nil
}
