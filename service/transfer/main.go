package main

import(
	"FileStoreServer/mq"
	"FileStoreServer/config"
	"FileStoreServer/store/oss"

	"log"
	"encoding/json"
	"os"
	"bufio"
)

func ProcessTransfer(msg []byte) bool {
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, pubData)
	if err != nil {
		log.Println(err)
		return false
	}

	filed, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err)
		return false
	}

	err = oss.Bucket().PutObject(pubData.DestLocation, bufio.NewReader(filed))
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func main() {
	log.Println("start transfer")
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer)
}