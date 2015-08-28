package magellan

import (
	"github.com/streadway/amqp"
	"github.com/ugorji/go/codec"
	"os"
	"reflect"
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

func SetupMessageQueue() (*MessageQueue, error) {
	q := new(MessageQueue)
	q.Host = os.Getenv("MAGELLAN_WORKER_AMQP_ADDR")
	q.Port = os.Getenv("MAGELLAN_WORKER_AMQP_PORT")
	q.Vhost = os.Getenv("MAGELLAN_WORKER_AMQP_VHOST")
	q.User = os.Getenv("MAGELLAN_WORKER_AMQP_USER")
	q.Password = os.Getenv("MAGELLAN_WORKER_AMQP_PASS")
	url := "amqp://" + q.User + ":" + q.Password + "@" + q.Host + ":" + q.Port + q.Vhost
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

func (q *MessageQueue) Consume() (chan *RequestMessage, error) {
	ch, err := q.Channel.Consume(q.RequestQueue, "_magellan_proxy_consumer", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	req_ch := make(chan *RequestMessage, 100)
	go func() {
		mh := codec.MsgpackHandle{RawToString: true}
		mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
		for msg := range ch {
			dec := codec.NewDecoderBytes(msg.Body, &mh)
			ret := new(RequestMessage)
			err = dec.Decode(ret)
			if err != nil {
				msg.Nack(false, false)
			} else {
				msg.Ack(false)
				req_ch <- ret
			}
		}
	}()

	return req_ch, nil
}

// vim:set noexpandtab ts=2:
