package main

import (
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func dial(url string) *amqp.Connection {
	for i := 0; i < 10; i++ {
		conn, err := amqp.Dial(url)
		if err == nil {
			return conn
		}
		log.Printf("[retry %d] waiting for RabbitMQ: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	log.Fatal("could not connect to RabbitMQ")
	return nil
}

func main() {
	id := os.Getenv("CONSUMER_ID")
	queueName := os.Getenv("QUEUE_NAME")
	exchangeName := os.Getenv("EXCHANGE_NAME")
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}

	conn := dial(url)
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("[%s] channel: %v", id, err)
	}
	defer ch.Close()

	ch.Qos(1, 0, false) // prefetch 1 — fair dispatch

	var queue amqp.Queue

	if exchangeName != "" {
		ch.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
		queue, _ = ch.QueueDeclare("", false, true, true, false, nil) // exclusive, auto-delete
		ch.QueueBind(queue.Name, "", exchangeName, false, nil)
	} else {
		queue, _ = ch.QueueDeclare(queueName, true, false, false, false, nil)
	}

	msgs, _ := ch.Consume(queue.Name, "", false, false, false, false, nil)

	log.Printf("[%s] waiting on queue: %s", id, queue.Name)
	for msg := range msgs {
		log.Printf("[%s] received: %s", id, msg.Body)
		time.Sleep(200 * time.Millisecond) // simulate processing
		msg.Ack(false)
		log.Printf("[%s] acked: %s", id, msg.Body)
	}
}
