package wabbit

import (
	"github.com/streadway/amqp"
)

type MockDialFn func(url string) (Connection, error)

type MockAmqpDialer struct {
	DialFn      MockDialFn
	DialInvoked bool
}

func (m *MockAmqpDialer) Dial(url string) (Connection, error) {
	m.DialInvoked = true
	return m.DialFn(url)
}

type MockChannelFn func() (Channel, error)
type MockNotifyCloseFn func(ch chan *amqp.Error) chan *amqp.Error

type MockAmqpConnection struct {
	ChannelFn       MockChannelFn
	ChannelFnCalled int
	ChannelInvoked  bool

	NotifyCloseFn      MockNotifyCloseFn
	NotifyCloseInvoked bool
}

func (m *MockAmqpConnection) Channel() (Channel, error) {
	m.ChannelInvoked = true
	m.ChannelFnCalled += 1
	return m.ChannelFn()
}

func (m *MockAmqpConnection) NotifyClose(ch chan *amqp.Error) chan *amqp.Error {
	m.NotifyCloseInvoked = true
	return m.NotifyCloseFn(ch)
}

type MockQueueDeclareFn func(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)

type MockAmqpChannel struct {
	QueueDeclareFn      MockQueueDeclareFn
	QueueDeclareInvoked bool
}

func (m *MockAmqpChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	m.QueueDeclareInvoked = true
	return m.QueueDeclareFn(name, durable, autoDelete, exclusive, noWait, args)
}
