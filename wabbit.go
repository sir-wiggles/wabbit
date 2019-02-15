package wabbit

import (
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type Wabbit struct {
	*sync.Mutex
	cond *sync.Cond

	url        string
	dialer     Dialer
	connection Connection
	channel    Channel
	connected  bool

	shutdown     chan bool
	disconnected chan *amqp.Error
}

func NewWabbit(url string) *Wabbit {
	w := &Wabbit{
		Mutex:    &sync.Mutex{},
		cond:     &sync.Cond{L: &sync.Mutex{}},
		dialer:   &AmqpDial{},
		shutdown: make(chan bool),
		url:      url,
	}

	w.init()
	time.Sleep(time.Second)
	return w
}

func (w *Wabbit) init() {
	var wg = &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		var ticker = time.NewTicker(time.Millisecond)
		for {
			select {

			case <-ticker.C:
				ticker.Stop()
				w.Lock()
				w.connectionLoop()
				w.connected = true
				w.Unlock()
				wg.Done()

			case <-w.disconnected:
				w.Lock()
				w.connectionLoop()
				w.connected = true
				w.Unlock()

			case <-w.shutdown:
				return
			}
			w.cond.Broadcast()
		}
	}()

	wg.Wait()
}

func (w *Wabbit) connectionLoop() {
	for {
		if err := w.connect(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
}

func (w *Wabbit) connect() error {

	connection, err := w.dialer.Dial(w.url)
	if err != nil {
		return err
	}

	channel, err := connection.Channel()
	if err != nil {
		return err
	}

	w.connection = connection
	w.channel = channel
	w.disconnected = connection.NotifyClose(make(chan *amqp.Error))

	return nil
}

type QueueArgs struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Arguments  amqp.Table
}

func (w *Wabbit) QueueDeclare(args QueueArgs) error {
	_, err := w.channel.QueueDeclare(
		args.Name,
		args.Durable,
		args.AutoDelete,
		args.Exclusive,
		args.NoWait,
		args.Arguments,
	)
	return err
}

func (w *Wabbit) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	for {
		w.Lock()
		if !w.connected {
			w.Unlock()

			w.cond.L.Lock()
			w.cond.Wait()
			w.cond.L.Unlock()

		} else {
			w.Unlock()
		}
		msg.Body = []byte(time.Now().String())
		if err := w.channel.Publish(exchange, key, mandatory, immediate, msg); err != nil {
			log.Println("Publishing error: ", err)
		}

		time.Sleep(time.Millisecond * 3000)
	}
}

func (w *Wabbit) Consume(queue string, messages chan amqp.Delivery) error {

	consumer := ""
	autoAck := true
	exclusive := false
	noLocal := false
	noWait := false
	var args amqp.Table

	var deliveries <-chan amqp.Delivery
	var err error

	for {
		w.Lock()
		if !w.connected {
			w.Unlock()

			w.cond.L.Lock()
			w.cond.Broadcast()
			w.cond.L.Unlock()
		} else {
			deliveries, err = w.channel.Consume(
				queue, consumer, autoAck, exclusive, noLocal, noWait, args,
			)
			if err != nil {
				return err
			}
			w.Unlock()
		}

		for m := range deliveries {
			messages <- m
		}

	}

}
