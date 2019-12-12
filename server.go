package main

import (
	"log"
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

	log.Println("Started server")

	// block until sigint
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig

	return nil
}
