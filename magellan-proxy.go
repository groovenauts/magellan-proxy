package main

import (
	"github.com/codegangsta/cli"
	"os"
	"os/exec"
)

func main() {
	app := cli.NewApp()
	app.Name = "magellan-proxy"
	app.Version = Version
	app.Usage = "bypass request to HTTP port"
	app.Author = "Groovenauts,Inc."
	app.Email = "tech@groovenauts.jp"
	app.Action = doMain
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "port, p",
			Value: 80,
			Usage: "Port to foward HTTP request",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "fork and exec command",
			Action: doRun,
		},
	}

	app.Run(os.Args)
}

func doMain(c *cli.Context) {
	doRun(c)
}

func doRun(c *cli.Context) {
	var cmd *exec.Cmd
	if len(c.Args()) > 0 {
		arg0 := c.Args()[0]
		args := c.Args()[1:len(c.Args())]
		cmd = exec.Command(arg0, args...)
	} else {
		println("Please specify command")
		return
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
  err := cmd.Run()
	if err != nil {
		println("cmd.Run() return fail")
	}
}

// vim:set noexpandtab ts=2:
