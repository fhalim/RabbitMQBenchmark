package main

import ("fmt"
	"github.com/streadway/amqp"
	"time"
	"strconv"
)

const (
	persistentExchangeName = "bmexch"
	persistentQueueName    = "bmqueue"
	messageCount           = 10000
)

func getChannel(conn *amqp.Connection) *amqp.Channel {
	channel, _ := conn.Channel()
	return channel
}
func doPublish(channel *amqp.Channel, count int, exchangeName string) {
	for i := 0; i < count; i++ {
		go channel.Publish(exchangeName, "", false, false, amqp.Publishing {
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            []byte("Hello world"),
				DeliveryMode:    amqp.Persistent,
				Priority:        0,
			})
	}
}
func doSubscribe(channel *amqp.Channel, count int, queueName string) (<-chan bool) {
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
func initializePermanentAmqpResources(channel *amqp.Channel, exchangeName string, queueName string) {
	fmt.Println("Initializing resources")
	channel.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	channel.QueueDeclare(queueName, true, false, false, false, nil)
	channel.QueueBind(queueName, "", exchangeName, false, nil)
	channel.QueuePurge(queueName, false)
}

func benchmarkPubSub(url string) {
	conn1, _ := amqp.Dial(url)
	channel1 := getChannel(conn1)
	channel2 := getChannel(conn1)
	defer conn1.Close()
	initializePermanentAmqpResources(channel1, persistentExchangeName, persistentQueueName)
	fmt.Println("Publishing", messageCount, "messages")
	startTime := time.Now()
	go doPublish(channel1, messageCount, persistentExchangeName)
	<-doSubscribe(channel2, messageCount, persistentQueueName)
	endTime := time.Now()
	timeElapsed := endTime.Sub(startTime)
	fmt.Println("Published and received", messageCount, "messages on", url, "in", time.Duration(timeElapsed),
		"at", (float64(messageCount)/timeElapsed.Seconds()), "records/sec")
}

func initializeAndDestroyEphemeralQueueAndExchange(connection *amqp.Connection, exchangeName string, queueName string, noWait bool) {
	channel := getChannel(connection)
	defer channel.Close()
	channel.QueueDeclare(queueName, false, true, true, false, nil)
	channel.ExchangeDeclare(exchangeName, "fanout", false, true, false, false, nil)
	channel.QueueBind(queueName, "", exchangeName, false, nil)
	channel.Consume(queueName, "", true, true, true, false, nil)
}

func benchmarkTempQueueExchangeCreationDeletion(url string) {
	conn1, _ := amqp.Dial(url)
	defer conn1.Close()
	startTime := time.Now()
	for idx := 0; idx < messageCount; idx++ {
		initializeAndDestroyEphemeralQueueAndExchange(conn1, "exchange" + strconv.Itoa(idx), "queue" + strconv.Itoa(idx), true)
	}
	endTime := time.Now()
	timeElapsed := endTime.Sub(startTime)
	fmt.Println("Created and destroyed queue and exchange", messageCount, "times on", url, "in", time.Duration(timeElapsed),
		"at", (float64(messageCount)/timeElapsed.Seconds()), "records/sec")
}

func main() {
	urls := []string{"amqp://guest:guest@localhost:5672/", "amqp://guest:guest@localhost:5673/"}
	for _, url := range urls {
		benchmarkPubSub(url)
		benchmarkTempQueueExchangeCreationDeletion(url)
	}

}
