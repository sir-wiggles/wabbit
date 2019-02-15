package wabbit

import "github.com/streadway/amqp"

type Dialer interface {
	Dial(url string) (Connection, error)
}

type AmqpDial struct{}

func (d *AmqpDial) Dial(url string) (Connection, error) {
	conn, err := amqp.Dial(url)
	return &AmqpConnection{conn}, err
}

type Connection interface {
	Channel() (Channel, error)
	NotifyClose(closed chan *amqp.Error) chan *amqp.Error
}

type AmqpConnection struct {
	connnetion *amqp.Connection
}

func (a *AmqpConnection) Channel() (Channel, error) {
	ch, err := a.connnetion.Channel()
	return &AmqpChannel{ch}, err
}

func (a *AmqpConnection) NotifyClose(ch chan *amqp.Error) chan *amqp.Error {
	return a.connnetion.NotifyClose(ch)
}

type Channel interface {
	QueueDeclare(
		name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table,
	) (amqp.Queue, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

type AmqpChannel struct {
	channel *amqp.Channel
}

func (a *AmqpChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return a.channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (a *AmqpChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return a.channel.Publish(exchange, key, mandatory, immediate, msg)
}

func (a *AmqpChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return a.channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}
