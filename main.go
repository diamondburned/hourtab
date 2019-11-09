package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
	"gitlab.com/diamondburned/hourtab/hourtab"
)

var opts *hourtab.SessionOptions

func main() {
	app := cli.NewApp()
	app.Name = "hourtab"
	app.Usage = "Tracks hour"

	def, err := hourtab.DefaultOptions()
	if err != nil {
		log.Fatalln("Unable to get default options:", err)
	}

	app.Before = func(c *cli.Context) error {
		opts = def
		return nil
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "dbpath, db",
			Value:       def.DBPath,
			Destination: &def.DBPath,
		},
		cli.DurationFlag{
			Name:        "sync-frequency, s",
			Value:       def.SyncFrequency,
			Destination: &def.SyncFrequency,
		},
		cli.UintFlag{
			Name:        "timeout-after, t",
			Value:       def.TimeoutAfter,
			Destination: &def.TimeoutAfter,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "serve",
			Action:      server,
			Description: "Runs the server",
		},
	}

	app.Run(os.Args)
}
