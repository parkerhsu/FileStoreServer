package mq

import (
	"log"
)

// 开始监听队列，获取消息
func StartConsume(qName, cName string, callback func(msg []byte) bool ) {
	msgs, err := channel.Consume (
		qName,
		cName,
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		log.Println(err)
		return
	}

	//done := make(chan bool)

	for msg := range msgs {
		suc := callback(msg.Body)
		if !suc {

		}
	}
}