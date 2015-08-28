package main

import (
	"github.com/codegangsta/cli"
	//"github.com/ugorji/go/codec"
	//"github.com/streadway/amqp"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"fmt"
	"errors"
	"path/filepath"
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

func spawn(args []string) (*os.Process, error) {
	if len(args) == 0 {
		return nil, errors.New("Please specify command")
	}

	arg0 := args[0]
	if arg0 == filepath.Base(arg0) {
		if lp, err := exec.LookPath(arg0); err != nil {
			return nil, errors.New(fmt.Sprintln("command", arg0, "not found"))
		} else {
			arg0 = lp
		}
	}
	var attr os.ProcAttr
	attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	child, err := os.StartProcess(arg0, args, &attr)
	if err != nil {
		return nil, err
	}
	return child, nil
}

func killChild(child *os.Process, sigchan chan os.Signal) {
	sig := <-sigchan
	_ = child.Signal(sig)
}

func doRun(c *cli.Context) {
	child, err := spawn(c.Args())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	sigchan := make(chan os.Signal)
	go killChild(child, sigchan)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer child.Wait()
	println("command started")
}

// vim:set noexpandtab ts=2:
