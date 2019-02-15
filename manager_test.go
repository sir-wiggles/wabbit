package wabbit

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func Test(t *testing.T) {

	ch := make(chan bool)

	w := NewWabbit("amqp://guest:guest@localhost:5672")

	err := w.QueueDeclare(QueueArgs{
		Name:    "foobar",
		Durable: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	log.Println("publishing")
	for i := 0; i < 10; i++ {
		go w.Publish("", "foobar", false, false, amqp.Publishing{
			Body: []byte(time.Now().String()),
		})
		time.Sleep(time.Millisecond * 100)
	}

	time.Sleep(time.Second * 2)
	log.Println("consumeing")
	messages := make(chan amqp.Delivery)
	go w.Consume("foobar", messages)

	log.Println("consumeing")
	for m := range messages {
		fmt.Println(string(m.Body))
	}

	<-ch

}

//var connectTests = []struct {
//    name          string
//    dialer        *MockAmqpDialer
//    err           error
//    connectionSet bool
//    channelSet    bool
//}{
//    {
//        name: "returns error on dial call",
//        dialer: &MockAmqpDialer{
//            DialFn: func(url string) (Connection, error) {
//                return nil, errors.New("dial error")
//            },
//        },
//        err: errors.New("dial error"),
//    },
//    {
//        name: "returns error on channel call",
//        dialer: &MockAmqpDialer{
//            DialFn: func(url string) (Connection, error) {
//                connection := &MockAmqpConnection{
//                    ChannelFn: func() (Channel, error) {
//                        return nil, errors.New("channel error")
//                    },
//                }
//                return connection, nil
//            },
//        },
//        err: errors.New("channel error"),
//    },
//    {
//        name: "notify the ready channel on connection",
//        dialer: &MockAmqpDialer{
//            DialFn: func(url string) (Connection, error) {

//                channel := &MockAmqpChannel{}

//                connection := &MockAmqpConnection{
//                    ChannelFn: func() (Channel, error) {
//                        return channel, nil
//                    },
//                    NotifyCloseFn: func(ch chan *amqp.Error) chan *amqp.Error {
//                        return ch
//                    },
//                }
//                return connection, nil
//            },
//        },
//        err:           nil,
//        connectionSet: true,
//        channelSet:    true,
//    },
//}

//func Test_connect(t *testing.T) {

//    g := gomega.NewGomegaWithT(t)

//    for _, tt := range connectTests {
//        t.Run(tt.name, func(t *testing.T) {
//            m := &Manager{
//                dialer: tt.dialer,
//                ready:  make(chan bool, 1),
//            }

//            err := m.connect()

//            if tt.err != nil {
//                g.Expect(err).Should(gomega.MatchError(tt.err))
//                return
//            }

//            g.Expect(
//                m.connection.(*MockAmqpConnection).NotifyCloseInvoked,
//            ).Should(gomega.BeTrue())

//            g.Expect(m.connection).ShouldNot(gomega.BeNil())
//            g.Expect(m.channel).ShouldNot(gomega.BeNil())
//        })
//    }
//}

//var connectionLoopTests = []struct {
//    name      string
//    dialer    *MockAmqpDialer
//    err       error
//    calls     int
//    connected bool
//}{
//    {
//        name: "returns when signaled to shutdown",
//        dialer: &MockAmqpDialer{
//            DialFn: func(url string) (Connection, error) {

//                channel := &MockAmqpChannel{}

//                connection := &MockAmqpConnection{
//                    ChannelFn: func() MockChannelFn {
//                        count := 0
//                        return func() (Channel, error) {
//                            defer func() { count++ }()
//                            if count < 1 {
//                                return nil, errors.New("channel error")
//                            }
//                            return channel, nil
//                        }
//                    }(),
//                    NotifyCloseFn: func(ch chan *amqp.Error) chan *amqp.Error {
//                        return ch
//                    },
//                }
//                return connection, nil
//            },
//        },
//        calls:     1,
//        connected: false,
//    },
//    {
//        name: "returns when connection is make",
//        dialer: &MockAmqpDialer{
//            DialFn: func(url string) (Connection, error) {

//                channel := &MockAmqpChannel{}

//                connection := &MockAmqpConnection{
//                    ChannelFn: func() MockChannelFn {
//                        count := 0
//                        return func() (Channel, error) {
//                            defer func() { count++ }()
//                            if count < 0 {
//                                return nil, errors.New("channel error")
//                            }
//                            return channel, nil
//                        }
//                    }(),
//                    NotifyCloseFn: func(ch chan *amqp.Error) chan *amqp.Error {
//                        return ch
//                    },
//                }
//                return connection, nil
//            },
//        },
//        calls:     1,
//        connected: true,
//    },
//    {
//        name: "should loop until connection is made",
//        dialer: &MockAmqpDialer{
//            DialFn: func() MockDialFn {
//                count := 0
//                channel := &MockAmqpChannel{}
//                connection := &MockAmqpConnection{
//                    ChannelFn: func() MockChannelFn {
//                        return func() (Channel, error) {
//                            defer func() { count++ }()
//                            if count < 2 {
//                                return nil, errors.New("channel error")
//                            }
//                            return channel, nil
//                        }
//                    }(),
//                    NotifyCloseFn: func(ch chan *amqp.Error) chan *amqp.Error {
//                        return ch
//                    },
//                }
//                return func(url string) (Connection, error) {
//                    return connection, nil
//                }
//            }(),
//        },
//        calls:     3,
//        connected: true,
//    },
//}

//func Test_connectionLoop(t *testing.T) {
//    g := gomega.NewGomegaWithT(t)

//    for _, tt := range connectionLoopTests {
//        t.Run(tt.name, func(t *testing.T) {

//            m := Manager{
//                dialer:   tt.dialer,
//                ready:    make(chan bool, 1),
//                shutdown: make(chan bool, 1),
//            }
//            go func() {
//                time.Sleep(
//                    (time.Duration(tt.calls-1) * time.Second) + (100 * time.Millisecond),
//                )
//                m.shutdown <- true
//            }()
//            m.connectionLoop()

//            if !tt.connected {
//                g.Expect(m.shutdown).Should(gomega.Receive())
//            }

//            if tt.connected {
//                g.Expect(
//                    m.connection.(*MockAmqpConnection).ChannelFnCalled,
//                ).Should(gomega.Equal(tt.calls))
//            }

//        })
//    }
//}
