package main

import (
	"github.com/streadway/amqp"
	"log"
	"os"
	"strings"
	"syscall"
)

type MessageQueue struct {
	Host             string
	Port             string
	Vhost            string
	User             string
	Password         string
	Connection       *amqp.Connection
	Channel          *amqp.Channel
	RequestQueue     string
	ResponseExchange string
}

type RequestMessage struct {
	Request       Request
	ReplyTo       string
	CorrelationId string
}

func SetupMessageQueue() (*MessageQueue, error) {
	q := new(MessageQueue)
	q.Host = os.Getenv("MAGELLAN_WORKER_AMQP_ADDR")
	q.Port = os.Getenv("MAGELLAN_WORKER_AMQP_PORT")
	q.Vhost = os.Getenv("MAGELLAN_WORKER_AMQP_VHOST")
	q.User = os.Getenv("MAGELLAN_WORKER_AMQP_USER")
	q.Password = os.Getenv("MAGELLAN_WORKER_AMQP_PASS")
	url := "amqp://" + q.User + ":" + q.Password + "@" + q.Host + ":" + q.Port + "/" + strings.Replace(q.Vhost, "/", "%2F", -1)
	log.Printf("connect to amqp URL = %s", url)
	var err error
	q.Connection, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	q.Channel, err = q.Connection.Channel()
	if err != nil {
		return nil, err
	}
	q.RequestQueue = os.Getenv("MAGELLAN_WORKER_REQUEST_QUEUE")
	q.ResponseExchange = os.Getenv("MAGELLAN_WORKER_RESPONSE_EXCHANGE")
	return q, nil
}

func (q *MessageQueue) Close() {
	q.Connection.Close()
}

func (q *MessageQueue) Consume(req_ch chan *RequestMessage) error {
	ch, err := q.Channel.Consume(q.RequestQueue, "_magellan_proxy_consumer", false, false, false, false, nil)
	if err != nil {
		q.SendToMyself(syscall.SIGTERM)
		return err
	}
	go func() {
		for msg := range ch {
			ret := new(RequestMessage)
			err = DecodeRequest(msg.Body, &ret.Request)
			if err != nil {
				log.Printf("fail to decode request message from TRMQ: %s", err.Error())
				msg.Nack(false, false)
			} else {
				ret.ReplyTo = msg.ReplyTo
				ret.CorrelationId = msg.CorrelationId
				req_ch <- ret
				// send Ack after message was accepted by a processing goroutine (note that the capacity of req_ch is 0).
				msg.Ack(false)
			}
		}
		log.Print("TRMQ connection closed.")
		q.SendToMyself(syscall.SIGTERM)
	}()

	return nil
}

func (q *MessageQueue) SendToMyself(signal os.Signal) {
	pid := os.Getpid()
	self, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("Error on os.FindProcess for %v because of %v\n", pid, err)
		return
	}
	if err := self.Signal(signal); err != nil {
		log.Printf("Error on send %v to %v because of %v\n", signal, pid, err)
	}
}

func (q *MessageQueue) Publish(req *RequestMessage, res *Response) error {
	p := amqp.Publishing{
		Headers:       amqp.Table{},
		DeliveryMode:  amqp.Persistent,
		Expiration:    "1000",
		CorrelationId: req.CorrelationId,
	}
	if err := res.Encode(&p.Body); err != nil {
		return err
	}
	if err := q.Channel.Publish(q.ResponseExchange, req.ReplyTo, true, false, p); err != nil {
		log.Printf("fail to publish response to TRMQ: %s", err.Error())
		return err
	}
	return nil
}

// vim:set noexpandtab ts=2:
