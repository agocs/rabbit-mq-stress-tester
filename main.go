package main

import (
	"log"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/streadway/amqp"
)

var totalTime int64 = 0
var totalCount int64 = 0

type MqMessage struct {
	TimeNow        time.Time
	SequenceNumber int
	Payload        string
}

func main() {
	app := cli.NewApp()
	app.Name = "tester"
	app.Usage = "Make the rabbit cry"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "server, s", Value: "rabbit-mq-test.cs1cloud.internal", Usage: "Hostname for RabbitMQ server"},
		cli.StringFlag{Name: "port, P", Value: "5672", Usage: "Port for RabbitMQ server"},
		cli.StringFlag{Name: "user, u", Value: "guest", Usage: "user for RabbitMQ server"},
		cli.StringFlag{Name: "password, pass", Value: "guest", Usage: "user password for RabbitMQ server"},
		cli.StringFlag{Name: "vhost, V", Value: "", Usage: "vhost for RabbitMQ server"},
		cli.IntFlag{Name: "producer, p", Value: 0, Usage: "Number of messages to produce, -1 to produce forever"},
		cli.StringFlag{Name: "exchange, x", Value: "", Usage: "Name of the exchange to send messages to"},
		cli.IntFlag{Name: "wait, w", Value: 0, Usage: "Number of nanoseconds to wait between publish events"},
		cli.IntFlag{Name: "consumer, c", Value: -1, Usage: "Number of messages to consume. 0 consumes forever"},
		cli.IntFlag{Name: "think-time, t", Value: 0, Usage: "Number milliseconds to wait before acknowledge. 0 auto ack"},
		cli.IntFlag{Name: "prefetch-count, f", Value: 0, Usage: "Number of unacknowledged messages. 0 unlimited"},
		cli.IntFlag{Name: "bytes, b", Value: 0, Usage: "number of extra bytes to add to the RabbitMQ message payload. About 50K max"},
		cli.IntFlag{Name: "concurrency, n", Value: 50, Usage: "number of reader/writer Goroutines"},
		cli.IntFlag{Name: "delay-messages, d", Value: 0, Usage: "Configures exchange to use delayed_message_exchange plugin (required to be installed in cluster). Only when -x is used. 0 doesn't use it"},
		cli.BoolFlag{Name: "quiet, q", Usage: "Print only errors to stdout"},
		cli.BoolFlag{Name: "wait-for-ack, a", Usage: "Wait for an ack or nack after enqueueing a message"},
	}
	app.Action = func(c *cli.Context) error {
		runApp(c)
		return nil
	}
	app.Run(os.Args)
}

func runApp(c *cli.Context) {
	println("Running!")
	porto := "amqp://"
	uri := porto + c.String("user") + ":" + c.String("password") + "@" + c.String("server") +
		":" + c.String("port")

	if c.String("vhost") != "" {
		uri += "/" + c.String("vhost")
	}

	exConfig := ExchangeConfig{c.String("exchange"), 0}

	if c.String("exchange") != "" && c.Int("delay-messages") > 0 {
		exConfig.DelayMessages = c.Int("delay-messages")
	}

	if c.Int("consumer") > -1 {
		thinkTime := c.Int("think-time")
		if thinkTime < 0 {
			log.Fatal("think-time should be non-negative")
		}
		config := ConsumerConfig{
			uri,
			thinkTime,
			exConfig,
			c.Int("prefetch-count"),
		}
		makeConsumers(c.Int("concurrency"), c.Int("consumer"), config)
	}

	if c.Int("producer") != 0 {
		config := ProducerConfig{
			uri,
			c.Int("bytes"),
			c.Bool("quiet"),
			c.Bool("wait-for-ack"),
			exConfig,
		}
		makeProducers(c.Int("producer"), c.Int("wait"), c.Int("concurrency"), config)
	}
}

type ExchangeConfig struct {
	Name          string
	DelayMessages int
}

func MakeQueueAndBind(c *amqp.Channel, exConfig ExchangeConfig) amqp.Queue {

	q := MakeQueue(c)

	// declare exchange and bind queue if not using nameless exchange
	if exConfig.Name != "" {
		MakeExchange(c, exConfig)

		err := c.QueueBind(q.Name, q.Name, exConfig.Name, false, nil)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	return q
}

func MakeExchange(c *amqp.Channel, exConfig ExchangeConfig) {
	if exConfig.DelayMessages > 0 {
		err := c.ExchangeDeclare(exConfig.Name, "x-delayed-message", true, false, false, false, map[string]interface{}{"x-delayed-type": "direct"})
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err := c.ExchangeDeclare(exConfig.Name, "direct", true, false, false, false, nil)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func MakeQueue(c *amqp.Channel) amqp.Queue {
	q, err := c.QueueDeclare("stress-test", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	return q
}

func makeProducers(n int, wait int, concurrency int, config ProducerConfig) {

	taskChan := make(chan int)
	for i := 0; i < concurrency; i++ {
		go Produce(config, taskChan, i)
	}

	start := time.Now()
	log.Print("Start sending messages to producers ...")
	for i := 0; i < n; i++ {
		taskChan <- i
		time.Sleep(time.Duration(int64(wait)))
	}

	log.Print("Waiting for producers to finish ...")
	time.Sleep(time.Duration(10000))

	close(taskChan)

	log.Printf("Finished: %s", time.Since(start))
}

func makeConsumers(concurrency int, toConsume int, config ConsumerConfig) {

	doneChan := make(chan bool)

	for i := 0; i < concurrency; i++ {
		// go Consume(uri, thinkTime, doneChan)
		go Consume(config, doneChan)
	}

	start := time.Now()

	if toConsume > 0 {
		for i := 0; i < toConsume; i++ {
			<-doneChan
			if i == 1 {
				start = time.Now()
			}
			log.Println("Consumed: ", i)
		}
	} else {

		for {
			<-doneChan
		}
	}

	log.Printf("Done consuming! %s", time.Since(start))
}
