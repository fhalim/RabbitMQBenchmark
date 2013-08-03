package main

import ("fmt"
	"github.com/streadway/amqp"
	"time"
	"strconv"
	"flag"
	"strings"
)

const (
	persistentExchangeName = "bmexch"
	persistentQueueName    = "bmqueue"
)

func getChannel(conn *amqp.Connection) *amqp.Channel {
	channel, _ := conn.Channel()
	return channel
}
func doPublish(channel *amqp.Channel, count int, exchangeName string) {
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

func benchmarkPubSub(url string, iterations int) {
	fmt.Println("Benchmarking pubsub")
	conn1, _ := amqp.Dial(url)
	channel1 := getChannel(conn1)
	channel2 := getChannel(conn1)
	defer conn1.Close()
	initializePermanentAmqpResources(channel1, persistentExchangeName, persistentQueueName)
	fmt.Println("Publishing", iterations, "messages")
	startTime := time.Now()
	go doPublish(channel1, iterations, persistentExchangeName)
	<-doSubscribe(channel2, iterations, persistentQueueName)
	endTime := time.Now()
	timeElapsed := endTime.Sub(startTime)
	fmt.Println("Published and received", iterations, "messages on", url, "in", time.Duration(timeElapsed),
		"at", (float64(iterations)/timeElapsed.Seconds()), "records/sec")
}

func initializeAndDestroyEphemeralQueueAndExchange(connection *amqp.Connection, exchangeName string, queueName string, noWait bool) {
	channel := getChannel(connection)
	defer channel.Close()
	channel.QueueDeclare(queueName, false, true, true, false, nil)
	channel.ExchangeDeclare(exchangeName, "fanout", false, true, false, false, nil)
	channel.QueueBind(queueName, "", exchangeName, false, nil)
	channel.Consume(queueName, "", true, true, true, false, nil)
}


func initializeAndDestroyTemporaryQueueAndExchange(connection *amqp.Connection, exchangeName string, queueName string, noWait bool) {
	channel := getChannel(connection)
	defer channel.Close()
	queue, _ := channel.QueueDeclare("", false, true, true, false, nil)
	channel.ExchangeDeclare(exchangeName, "fanout", false, true, false, false, nil)
	channel.QueueBind(queue.Name, "", exchangeName, false, nil)
	channel.Consume(queue.Name, "", true, true, true, false, nil)
}


func benchmarkTransientQueueExchangeCreationDeletion(url string, iterations int) {
	fmt.Println("Benchmarking transient queue creation and cleanup")
	conn1, _ := amqp.Dial(url)
	defer conn1.Close()
	startTime := time.Now()
	for idx := 0; idx < iterations; idx++ {
		initializeAndDestroyEphemeralQueueAndExchange(conn1, "exchange" + strconv.Itoa(idx), "queue" + strconv.Itoa(idx), true)
	}
	endTime := time.Now()
	timeElapsed := endTime.Sub(startTime)
	fmt.Println("Created and destroyed queue and exchange", iterations, "times on", url, "in", time.Duration(timeElapsed),
		"at", (float64(iterations)/timeElapsed.Seconds()), "records/sec")
}

func benchmarkTemporaryQueueExchangeCreationDeletion(url string, iterations int) {
	fmt.Println("Benchmarking temporary queue creation and cleanup")
	conn1, _ := amqp.Dial(url)
	defer conn1.Close()
	startTime := time.Now()
	for idx := 0; idx < iterations; idx++ {
		initializeAndDestroyTemporaryQueueAndExchange(conn1, "exchange" + strconv.Itoa(idx), "queue" + strconv.Itoa(idx), true)
	}
	endTime := time.Now()
	timeElapsed := endTime.Sub(startTime)
	fmt.Println("Created and destroyed queue and exchange", iterations, "times on", url, "in", time.Duration(timeElapsed),
		"at", (float64(iterations)/timeElapsed.Seconds()), "records/sec")
}

func main() {
	urlsOpt := flag.String("urls", "amqp://guest:guest@localhost:5672/,amqp://guest:guest@localhost:5673/", "URLS to run the tests against")
	urls := strings.Split(*urlsOpt, ",")
	iterations := flag.Int("iterations", 1000, "Number of iterations")
	benchmarkPublishSubscribe := flag.Bool("runpubsub", true, "Benchmark publish/subscribe")
	benchmarkTransientQueueCreationCleanup := flag.Bool("runtransientqueue", true, "Benchmark transient queue creation/cleanup")
	benchmarkTemporaryQueueCreationCleanup := flag.Bool("runtemporaryqueue", true, "Benchmark temporary queue creation/cleanup")

	flag.Parse()

	for _, url := range urls {
		if (*benchmarkPublishSubscribe) {benchmarkPubSub(url, *iterations)}
		if (*benchmarkTransientQueueCreationCleanup) {benchmarkTransientQueueExchangeCreationDeletion(url, *iterations)}
		if (*benchmarkTemporaryQueueCreationCleanup) {benchmarkTemporaryQueueExchangeCreationDeletion(url, *iterations)}
	}

}
