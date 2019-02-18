package wabbit

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func Test(t *testing.T) {

	w := NewWabbit("amqp://guest:guest@localhost:5672")
	queue := "foobar"

	err := w.QueueDeclare(&QueueParameters{
		Name:    queue,
		Durable: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	pp := &PublisherParameters{
		Queue: queue,
	}
	pc1 := w.Publish(pp)
	pc2 := w.Publish(pp)
	pc3 := w.Publish(pp)

	cp := &ConsumerParameters{
		Queue:   queue,
		AutoAck: true,
	}
	cc := w.Consume(cp)

	ticker := time.NewTicker(time.Second * 6)

	for {
		select {
		case pc1 <- amqp.Publishing{Body: []byte(fmt.Sprintf("%d, %s", 1, time.Now().String()))}:
			time.Sleep(time.Millisecond * 500)
		case pc2 <- amqp.Publishing{Body: []byte(fmt.Sprintf("%d, %s", 2, time.Now().String()))}:
			time.Sleep(time.Millisecond * 700)
		case pc3 <- amqp.Publishing{Body: []byte(fmt.Sprintf("%d, %s", 3, time.Now().String()))}:
			time.Sleep(time.Millisecond * 900)
		case msg := <-cc:
			log.Println("Consumed: ", string(msg.Body))
		case <-ticker.C:
			log.Println("shutting down...")
			w.Shutdown()
			return
		}
	}
}
