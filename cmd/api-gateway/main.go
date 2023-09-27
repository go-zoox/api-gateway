package main

import (
	"github.com/go-zoox/api-gateway"
	"github.com/go-zoox/api-gateway/core"
	"github.com/go-zoox/cli"
	"github.com/go-zoox/config"
	"github.com/go-zoox/core-utils/fmt"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
)

func main() {
	app := cli.NewSingleProgram(&cli.SingleProgramConfig{
		Name:    "api-gateway",
		Usage:   "An Easy Self Hosted API Gateway",
		Version: api.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "config",
				// Value:   "conf/api-gateway.yaml",
				Usage:   "The path to the configuration file",
				Aliases: []string{"c"},
				// Required: true,
			},
			&cli.StringFlag{
				Name:    "port",
				Usage:   "The port to listen on",
				Aliases: []string{"p"},
			},
		},
	})

	app.Command(func(c *cli.Context) error {
		configFilePath := c.String("config")
		if configFilePath == "" {
			configFilePath = "/etc/api-gateway/config.yaml"
		}

		var cfg core.Config

		if configFilePath != "" {
			if !fs.IsExist(configFilePath) {
				return fmt.Errorf("config file(%s) not found", configFilePath)
			}

			if err := config.Load(&cfg, &config.LoadOptions{
				FilePath: configFilePath,
			}); err != nil {
				return fmt.Errorf("failed to read config file: %s", err)
			}
		}

		if c.Int64("port") != 0 {
			cfg.Port = c.Int64("port")
		}

		if cfg.Port == 0 {
			cfg.Port = 8080
		}

		// @TODO
		if logger.IsDebugLevel() {
			// logger.Debug("config: %v", cfg)
			fmt.PrintJSON("config:", cfg)
		}

		app, err := core.New(api.Version, &cfg)
		if err != nil {
			return fmt.Errorf("failed to create core: %s", err)
		}

		return app.Run()
	})

	if err := app.RunWithError(); err != nil {
		logger.Fatal("%s", err.Error())
	}
}
