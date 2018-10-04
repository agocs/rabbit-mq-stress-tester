package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

func Consume(uri string, thinkTime int, doneChan chan bool) {
	log.Println("Consuming...")
	connection, err := amqp.Dial(uri)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer channel.Close()

	q := MakeQueue(channel)

	msgs, err := channel.Consume(q.Name, "", thinkTime == 0, false, false, false, nil)
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

		if thinkTime != 0 {
		  go func() {
        time.Sleep(time.Duration(thinkTime) * time.Millisecond)
        d.Ack(false)
      }()
		}
	}

	log.Println("done receiving")

}
