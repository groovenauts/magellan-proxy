package main

import (
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/groovenauts/magellan-proxy/magellan"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
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

func watchChild(child *os.Process, sigchan chan os.Signal) {
	_, _ = child.Wait()
	sigchan <- os.Interrupt
}

func processSignal(sigchan chan os.Signal, child *os.Process, req_ch chan *magellan.RequestMessage, exitQueue chan bool) {
	sig := <-sigchan
	_ = child.Signal(sig)
	close(req_ch)
	exitQueue <- true
	close(exitQueue)
}

func processRequest(mq *magellan.MessageQueue, req_ch chan *magellan.RequestMessage) {
	for req := range req_ch {
		println(req.Request.Env.Method, req.Request.Env.Url)

		res := magellan.Response{
			Headers:      map[string]string{"Content-Type": "text/plain"},
			Status:       "200",
			Body:         "Hello World!\n",
			BodyEncoding: "plain",
		}
		mq.Publish(req, &res)
	}
}

func doRun(c *cli.Context) {
	mq, err := magellan.SetupMessageQueue()
	if err != nil {
		fmt.Println("fail to setup MQ:", err.Error())
		return
	}
	defer mq.Close()

	req_ch, err := mq.Consume()
	if err != nil {
		fmt.Println("fail to get message:", err.Error())
		return
	}

	child, err := spawn(c.Args())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	println("command started")

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go watchChild(child, sigchan)

	exitQueue := make(chan bool)

	go processSignal(sigchan, child, req_ch, exitQueue)

	go processRequest(mq, req_ch)

	for exit_p := range exitQueue {
		if exit_p {
			break
		}
	}
}

// vim:set noexpandtab ts=2:
