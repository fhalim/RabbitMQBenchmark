package main

import ("fmt"
	"github.com/streadway/amqp"
	"time"
)

func getChannel() (*amqp.Connection, *amqp.Channel) {
	conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
	channel, _ := conn.Channel()
	return conn, channel
}
func doPublish(channel *amqp.Channel, count int) {
	for i := 0; i < count; i++ {
		channel.Publish("bmexch", "", false, false, amqp.Publishing {
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
	del, _ := channel.Consume("bmqueue", "", true, false, false, false, nil)
	completed := 0
	outChannel := make(chan bool)
	go func(c <-chan amqp.Delivery) {
		for _ = range c {
			//fmt.Print(".")
			completed += 1
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
	channel1.ExchangeDeclare("bmexch", "fanout", true, false, false, false, nil)
	channel1.QueueDeclare("bmqueue", true, false, false, false, nil)
	channel1.QueueBind("bmqueue", "", "bmexch", false, nil)
	messageCount := 100000
	fmt.Println("Publishing", messageCount, "messages")
	startTime := time.Now()
	go doPublish(channel1, messageCount)
	<-doSubscribe(channel2, messageCount)
	endTime := time.Now()
	timeElapsed := endTime.Sub(startTime)
	fmt.Println("Published", messageCount, "messages in", time.Duration(timeElapsed), "at", (float64(messageCount)/timeElapsed.Seconds()))
}
func main() {
	doBenchmark()
}
