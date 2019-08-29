package mq

import (
	"github.com/streadway/amqp"
	"FileStoreServer/config"

	"log"
)

var conn *amqp.Connection
var channel *amqp.Channel

func initChannel() bool {
	if channel != nil {
		return true
	}

	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println(err)
		return false
	}

	channel, err = conn.Channel()
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func Publish(exchange, routineKey string, msg []byte) bool {
	if !initChannel() {
		return false
	}

	err := channel.Publish(
		exchange,
		routineKey,
		false,
		false,
		amqp.Publishing {
			ContentType: "text/plain",
			Body: msg,
		},
	)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}