package main

import (
	"context"
	"log/slog"
	"os"

	wolproxy "github.com/benfiola/homelab-wol-proxy/pkg"
	"github.com/urfave/cli/v3"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	err := (&cli.Command{
		Action: func(ctx context.Context, c *cli.Command) error {
			address := c.String("address")
			backend := c.String("backend")
			wolHostname := c.String("wol-hostname")
			wolMacAddress := c.String("wol-mac-address")

			proxy, err := wolproxy.New(wolproxy.Opts{
				Address:     address,
				Backend:     backend,
				Logger:      logger,
				WolHostname: wolHostname,
				WolMacAddress: wolMacAddress,
			})
			if err != nil {
				return err
			}

			return proxy.Run()
		},
		Description: "run the proxy",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "backend",
				Required: true,
				Sources:  cli.EnvVars("WOLPROXY_BACKEND"),
				Usage:    "proxy backend",
			},
			&cli.StringFlag{
				Name: "wol-hostname",
				Usage: "hostname to send wake-on-lan packet to",
				Sources: cli.EnvVars("WOLPROXY_WOL_HOSTNAME"),
			},
			&cli.StringFlag{
				Name:    "wol-mac-address",
				Required: true,
				Usage:   "mac-address to send wake-on-lan packet to",
				Sources: cli.EnvVars("WOLPROXY_WOL_MAC_ADDRESS"),
			},
			&cli.StringFlag{
				Name:    "address",
				Sources: cli.EnvVars("WOLPROXY_ADDRESS"),
				Usage:   "address to listen on",
				Value:   "0.0.0.0:8080",
			},
		},
	}).Run(context.Background(), os.Args)
	code := 0
	if err != nil {
		logger.Error("command failed", "error", err.Error())
		code = 1
	}
	os.Exit(code)
}
