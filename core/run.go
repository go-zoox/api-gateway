package core

import (
	"fmt"
)

func (c *core) Run() error {
	if err := c.build(); err != nil {
		return err
	}

	return c.app.Run(fmt.Sprintf(":%d", c.cfg.Port))
}
