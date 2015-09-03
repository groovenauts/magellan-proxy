package main

import (
	"errors"
	"github.com/codegangsta/cli"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("magellan-proxy: ")

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
		cli.IntFlag{
			Name:  "num, n",
			Value: 1,
			Usage: "Maximum number concurrent HTTP request",
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
		log.Println("Please specify command")
		return nil, errors.New("Please specify command")
	}

	arg0 := args[0]
	if arg0 == filepath.Base(arg0) {
		if lp, err := exec.LookPath(arg0); err != nil {
			msg := "command " + arg0 + " not found"
			log.Println(msg)
			return nil, errors.New(msg)
		} else {
			arg0 = lp
		}
	}
	var attr os.ProcAttr
	attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	child, err := os.StartProcess(arg0, args, &attr)
	if err != nil {
		log.Printf("fail to os.StartProcess: %s", err.Error())
		return nil, err
	}
	return child, nil
}

func watchChild(child *os.Process, sigchan chan os.Signal) {
	_, _ = child.Wait()
	sigchan <- os.Interrupt
}

func processSignal(sigchan chan os.Signal, child *os.Process, req_ch chan *RequestMessage, exitQueue chan bool) {
	sig := <-sigchan
	_ = child.Signal(sig)
	close(req_ch)
	exitQueue <- true
	close(exitQueue)
}

func processRequest(mq *MessageQueue, req_ch chan *RequestMessage) {
	for req := range req_ch {
		log.Println(req.Request.Env.Method + " " + req.Request.Env.Url)

		res, err := ProcessHttpRequest(&req.Request)
		if err != nil {
			log.Printf("ProcessHttpRequest fail: %s", err.Error())
			res = &Response{
				Headers:      map[string]string{"Content-Type": "text/plain"},
				Status:       "200",
				Body:         []byte("magellan-proxy: ProcessHttpRequest fail: " + err.Error()),
				BodyEncoding: "plain",
			}
		}
		mq.Publish(req, res)
	}
}

func doRun(c *cli.Context) {
	mq, err := SetupMessageQueue()
	if err != nil {
		log.Printf("fail to setup MQ: %s", err.Error())
		return
	}
	defer mq.Close()

	req_ch, err := mq.Consume()
	if err != nil {
		log.Printf("fail to get message: %s", err.Error())
		return
	}

	child, err := spawn(c.Args())
	if err != nil {
		return
	}

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go watchChild(child, sigchan)

	exitQueue := make(chan bool)

	go processSignal(sigchan, child, req_ch, exitQueue)

	jobNum := c.Int("num")
	InitHttpTransport(c.Int("port"), jobNum)

	for i := 0; i < jobNum; i++ {
		go processRequest(mq, req_ch)
	}

	for exit_p := range exitQueue {
		if exit_p {
			break
		}
	}
}

// vim:set noexpandtab ts=2:
