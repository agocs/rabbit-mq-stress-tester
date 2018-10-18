package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type ConsumerConfig struct {
	Uri            string
	ThinkTime      int
	ExchangeConfig ExchangeConfig
	PrefetchCount  int
}

func Consume(config ConsumerConfig, doneChan chan bool) {
	log.Println("Consuming...")
	connection, err := amqp.Dial(config.Uri)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer channel.Close()

	if config.PrefetchCount > 0 {
		if channel.Qos(config.PrefetchCount, 0, false) != nil {
			log.Fatal(err.Error())
		}
	}

	q := MakeQueueAndBind(channel, config.ExchangeConfig)

	msgs, err := channel.Consume(q.Name, "", config.ThinkTime == 0, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	for d := range msgs {
		doneChan <- true
		var thisMessage MqMessage
		err4 := json.Unmarshal(d.Body, &thisMessage)
		if err4 != nil {
			log.Printf("Error unmarshalling! %s", err.Error())
		}
		log.Printf("Message age: %s", time.Since(thisMessage.TimeNow))

		if config.ThinkTime != 0 {
			go func() {
				time.Sleep(time.Duration(config.ThinkTime) * time.Millisecond)
				d.Ack(false)
			}()
		}
	}

	log.Println("done receiving")

}
