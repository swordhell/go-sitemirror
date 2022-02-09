package main

import (
	"os"
	"os/signal"

	"go-sitemirror/cacher"
	"go-sitemirror/engine"
)

func main() {
	config, err := engine.ParseConfig(os.Args[0], os.Args[1:], os.Stderr)
	if err != nil {
		os.Exit(1)
	}

	e := engine.FromConfig(cacher.NewFs(), config)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for sig := range c {
		switch sig {
		case os.Interrupt:
			e.Stop()
			os.Exit(0)
		}
	}
}
