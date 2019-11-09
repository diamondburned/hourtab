package main

import (
	"os"
	"os/signal"

	"github.com/urfave/cli"
	"gitlab.com/diamondburned/hourtab/hourtab"
)

func server(c *cli.Context) error {
	s, err := hourtab.New(opts)
	if err != nil {
		return err
	}

	s.Start()
	defer s.Stop()

	// block until sigint
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig

	return nil
}
