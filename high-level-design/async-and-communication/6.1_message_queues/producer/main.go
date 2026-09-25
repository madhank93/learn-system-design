package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ch *amqp.Channel

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
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}

	conn := dial(url)
	defer conn.Close()

	var err error
	ch, err = conn.Channel()
	if err != nil {
		log.Fatalf("channel: %v", err)
	}
	defer ch.Close()

	ch.QueueDeclare("work", true, false, false, false, nil)
	ch.ExchangeDeclare("events", "fanout", true, false, false, false, nil)

	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query().Get("msg")
		err := ch.Publish("", "work", false, false, amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(msg),
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "→ work queue: %s\n", msg)
		log.Printf("published to work: %s", msg)
	})

	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query().Get("msg")
		err := ch.Publish("events", "", false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "→ events fanout: %s\n", msg)
		log.Printf("published to events: %s", msg)
	})

	log.Println("Producer listening on :8080")
	http.ListenAndServe(":8080", nil)
}
