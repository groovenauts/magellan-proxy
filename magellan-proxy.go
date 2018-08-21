package main

import (
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func setTimezone(zonename string) {
	location, err := time.LoadLocation(zonename)
	if err != nil {
		location = time.FixedZone("UTC", 0)
	}
	time.Local = location
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix(fmt.Sprintf("magellan-proxy[%d]: ", os.Getpid()))

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
		cli.StringFlag{
			Name:  "publish",
			Value: "/publish",
			Usage: "Specify URL path to Post Publish message from MQTT",
		},
		cli.StringFlag{
			Name:  "timezone",
			Value: os.Getenv("TIMEZONE"),
			Usage: "Specify Timezone name in the IANA Time Zone Database",
		},
		cli.IntFlag{
			Name:  "timeout",
			Value: 60,
			Usage: "Specify Timeout seconds to wait for the port opened",
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
	_, err := child.Wait()
	if err != nil {
		log.Printf("failed to wait child process %v because of %v\n", child, err)
	}
	sigchan <- os.Interrupt
}

func processSignal(sigchan chan os.Signal, child *os.Process, req_ch chan *RequestMessage, exitQueue chan bool) {
	sig := <-sigchan
	close(req_ch)
	if err := child.Signal(sig); err != nil {
		log.Printf("failed to send signal %v to child process %v because of %v\n", sig, child, err)
	}
	_, err := child.Wait()
	if err != nil {
		log.Printf("failed to wait child process %v because of %v\n", child, err)
	}
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
		if res != nil {
			mq.Publish(req, res)
		}
	}
}

func doRun(c *cli.Context) {
	mq, err := SetupMessageQueue()
	if err != nil {
		log.Printf("fail to setup MQ: %s", err.Error())
		return
	}
	defer mq.Close()

	child, err := spawn(c.Args())
	if err != nil {
		return
	}

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go watchChild(child, sigchan)

	exitQueue := make(chan bool)

	req_ch := make(chan *RequestMessage)

	go processSignal(sigchan, child, req_ch, exitQueue)

	jobNum := c.Int("num")
	portNo := c.Int("port")
	InitHttpTransport(portNo, jobNum, c.String("publish"))

	setTimezone(c.String("timezone"))

	// wait until backend application server start to listen socket
	maxWait := c.Int("timeout")
	ready := false
	for i := 0; i < maxWait; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", portNo))
		if err != nil {
			if i == maxWait-1 {
				log.Printf("timeout waiting for application server start to listen socket port:%d, err=%s",
					portNo, err.Error())
				sigchan <- os.Interrupt
			}
			time.Sleep(1000000000) // 1 sec
		} else {
			log.Printf("Confirmed application server listen to 127.0.0.1:#{portNo}")
			conn.Close()
			ready = true
			break
		}
	}

	if ready {
		// start fetch request from TRMQ
		err := mq.Consume(req_ch)
		if err != nil {
			log.Printf("fail to get message from TRMQ: %s", err.Error())
			return
		}
		for i := 0; i < jobNum; i++ {
			go processRequest(mq, req_ch)
		}
	}

	for exit_p := range exitQueue {
		if exit_p {
			break
		}
	}
}

// vim:set noexpandtab ts=2:
