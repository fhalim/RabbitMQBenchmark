package main

import ("fmt"
	"mylibrary"
	"github.com/streadway/amqp"
	"bufio"
	"os"
)

const (
	PI       = 3.14
	Language = "Go"
)
const (
	Red   = iota
	Green = iota
	Blue  = iota
)

func getChannel() (*amqp.Connection, *amqp.Channel) {
	conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
	channel, _ := conn.Channel()
	return conn, channel
}
func doPublish(channel *amqp.Channel) {
	for i := 0; i < 100000; i++ {
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
func doSubscribe(channel *amqp.Channel) {
	del, _ := channel.Consume("bmqueue", "", true, false, false, false, nil)
	go func(c <-chan amqp.Delivery) {
		for _ = range c {
			fmt.Print(".")
		}
	}(del)
}

func doBenchmark() {
	conn1, channel1 := getChannel()
	conn2, channel2 := getChannel()
	defer conn1.Close()
	defer conn2.Close()
	channel1.ExchangeDeclare("bmexch", "fanout", true, false, false, false, nil)
	channel1.QueueDeclare("bmqueue", true, false, false, false, nil)
	channel1.QueueBind("bmqueue", "", "bmexch", false, nil)
	go doPublish(channel1)
	go doSubscribe(channel2)
	fmt.Println("Please press enter to stop")
	bufio.NewReader(os.Stdin).ReadLine()
}
func doPrint() {
	message := "Hello go"
	greeting := &message

	*greeting = "hi"
	fmt.Println(message)
	fmt.Println(*greeting)
	fmt.Println(Green)
	fmt.Println(mylibrary.PrintMessage("pete"))
}
func main() {
	//doPrint()
	doBenchmark()

}
