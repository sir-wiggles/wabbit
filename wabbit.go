package wabbit

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

type session struct {
	*amqp.Connection
	*amqp.Channel
}

type Wabbit struct {
	url string

	connection   *amqp.Connection
	sessions     chan *session
	shutdown     chan bool
	disconnected chan *amqp.Error
	queues       map[string]*QueueParameters
}

func NewWabbit(url string) *Wabbit {
	w := &Wabbit{
		sessions: make(chan *session),
		shutdown: make(chan bool),
		url:      url,
		queues:   make(map[string]*QueueParameters),
	}

	w.connect()
	return w
}

func (w Wabbit) Shutdown() {
	w.shutdown <- true
}

func (w *Wabbit) connect() {
	go func() {
		defer func() {
			w.connection.Close()
		}()
	Connection:
		for {
			connection, err := amqp.Dial(w.url)
			if err != nil {
				log.Println(err.(*amqp.Error).Reason)
				time.Sleep(time.Second * 2)
				continue
			}
			w.connection = connection
			w.disconnected = connection.NotifyClose(make(chan *amqp.Error))

		Channel:
			for {
				channel, err := connection.Channel()
				if err != nil {
					log.Println(err.(*amqp.Error).Reason)
					break Connection
				}

				select {
				case w.sessions <- &session{connection, channel}:
				case <-w.disconnected:
					break Channel
				case <-w.shutdown:
					return
				}
			}
		}
	}()
}

type QueueParameters struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Arguments  amqp.Table
}

func (w *Wabbit) QueueDeclare(params *QueueParameters) error {
	s := <-w.sessions
	_, err := s.QueueDeclare(
		params.Name,
		params.Durable,
		params.AutoDelete,
		params.Exclusive,
		params.NoWait,
		params.Arguments,
	)
	if err != nil {
		return err
	}

	w.queues[params.Name] = params
	return nil
}

type PublisherParameters struct {
	Exchange  string
	Queue     string
	Mandatory bool
	Immediate bool
}

func (w *Wabbit) Publish(params *PublisherParameters) chan amqp.Publishing {
	var input = make(chan amqp.Publishing)

	go func() {
		for session := range w.sessions {
			for msg := range input {
				if err := session.Publish(
					params.Exchange,
					params.Queue,
					params.Mandatory,
					params.Immediate,
					msg,
				); err != nil {
					log.Println("Publishing error:", err)
					break
				}
			}
		}
	}()
	return input
}

type ConsumerParameters struct {
	Queue     string //:= "foobar"
	Consumer  string //:= ""
	AutoAck   bool   //:= true
	Exclusive bool   //:= false
	NoLocal   bool   //:= false
	NoWait    bool   //:= false
	Args      amqp.Table
}

func (w *Wabbit) Consume(params *ConsumerParameters) chan amqp.Delivery {
	var output = make(chan amqp.Delivery)

	go func() {
		for session := range w.sessions {
			deliveries, err := session.Consume(
				params.Queue,
				params.Consumer,
				params.AutoAck,
				params.Exclusive,
				params.NoLocal,
				params.NoWait,
				params.Args,
			)

			if err != nil {
				log.Println("Consume Error:", err)
				continue
			}

			for d := range deliveries {
				output <- d
			}
		}
	}()

	return output
}
