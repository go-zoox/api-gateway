package core

import (
	"fmt"
)

func (c *core) Run() error {
	if err := c.build(); err != nil {
		return err
	}

	// Stop all health checks when shutting down
	defer func() {
		if c.lbManager != nil {
			c.lbManager.StopAllHealthChecks()
		}
	}()

	return c.app.Run(fmt.Sprintf(":%d", c.cfg.Port))
}
