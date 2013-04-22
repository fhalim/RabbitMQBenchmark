package main

import ("fmt"
	"github.com/streadway/amqp"
	"time"
)

const (
	exchangeName = "bmexch"
	queueName    = "bmqueue"
	url          = "amqp://guest:guest@localhost:5672/"
	messageCount = 100000
)

func getChannel() (*amqp.Connection, *amqp.Channel) {
	conn, _ := amqp.Dial(url)
	channel, _ := conn.Channel()
	return conn, channel
}
func doPublish(channel *amqp.Channel, count int) {
	for i := 0; i < count; i++ {
		channel.Publish(exchangeName, "", false, false, amqp.Publishing {
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            []byte("Hello world"),
				DeliveryMode:    amqp.Persistent,
				Priority:        0,
			})
	}

}
func doSubscribe(channel *amqp.Channel, count int) (<-chan bool) {
	del, _ := channel.Consume(queueName, "", true, false, false, false, nil)
	completed := 0
	outChannel := make(chan bool)
	go func(c <-chan amqp.Delivery) {
		for _ = range c {
			completed++
			if completed == count {
				outChannel <- true
			}
		}
	}(del)
	return outChannel
}

func doBenchmark() {
	conn1, channel1 := getChannel()
	conn2, channel2 := getChannel()
	defer conn1.Close()
	defer conn2.Close()
	channel1.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	channel1.QueueDeclare(queueName, true, false, false, false, nil)
	channel1.QueueBind(queueName, "", exchangeName, false, nil)
	fmt.Println("Publishing", messageCount, "messages")
	startTime := time.Now()
	go doPublish(channel1, messageCount)
	<-doSubscribe(channel2, messageCount)
	endTime := time.Now()
	timeElapsed := endTime.Sub(startTime)
	fmt.Println("Published", messageCount, "messages in", time.Duration(timeElapsed), "at", (float64(messageCount)/timeElapsed.Seconds()), "records/sec")
}
func main() {
	doBenchmark()
}
