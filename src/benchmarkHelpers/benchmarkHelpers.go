/**
 * Created with IntelliJ IDEA.
 * User: fawad
 * Date: 8/4/13
 * Time: 7:27 PM
 * To change this template use File | Settings | File Templates.
 */
package benchmarkHelpers

import ("fmt"
	"github.com/streadway/amqp"
	"time"
	"strconv"
)

type BenchmarkParams struct {
	Url        string
	Iterations int
}

func BenchmarkPubSub(parm BenchmarkParams, exchangeName string, queueName string) {
	fmt.Println("Benchmarking pubsub")
	conn1, _ := amqp.Dial(parm.Url)
	channel1 := getChannel(conn1)
	channel2 := getChannel(conn1)
	defer conn1.Close()
	initializePermanentAmqpResources(channel1, exchangeName, queueName)
	fmt.Println("Publishing", parm.Iterations, "messages")
	benchmark("Publish and receive over permanent resources", parm, func() {go doPublish(channel1, parm.Iterations, exchangeName)
			<-doSubscribe(channel2, parm.Iterations, queueName)})
}

func BenchmarkTransientQueueExchangeCreationDeletion(parm BenchmarkParams) {
	fmt.Println("Benchmarking transient queue creation and cleanup")
	conn1, _ := amqp.Dial(parm.Url)
	defer conn1.Close()
	benchmark("Transient queue/exchange creation/cleanup", parm, func() {for idx := 0; idx < parm.Iterations; idx++ {
			initializeAndDestroyEphemeralQueueAndExchange(conn1, "exchange" + strconv.Itoa(idx), "queue" + strconv.Itoa(idx), true)
		}})
}

func BenchmarkTemporaryQueueExchangeCreationDeletion(parm BenchmarkParams) {
	fmt.Println("Benchmarking temporary queue creation and cleanup")
	conn1, _ := amqp.Dial(parm.Url)
	defer conn1.Close()
	benchmark("Temporary queue/exchange creation/cleanup", parm, func() {
			for idx := 0; idx < parm.Iterations; idx++ {
				initializeAndDestroyTemporaryQueueAndExchange(conn1, "exchange" + strconv.Itoa(idx), "queue" + strconv.Itoa(idx), true)
			}})
}

func benchmark(description string, parm BenchmarkParams, function func ()) {
	startTime := time.Now()
	function()
	endTime := time.Now()
	timeElapsed := endTime.Sub(startTime)
	fmt.Println(description, parm.Iterations, "iterations against", parm.Url, "completed in", time.Duration(timeElapsed),
		"at", (float64(parm.Iterations)/timeElapsed.Seconds()), "/sec")
}

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
