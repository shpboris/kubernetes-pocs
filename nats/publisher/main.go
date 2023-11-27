package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"log"
	"time"
)

func main() {
	nc, err := nats.Connect("ws://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			<-ticker.C
			msg := createMessage()
			err = nc.PublishMsg(msg)
			logrus.Info("Published message: ", string(msg.Data))
			if err != nil {
				logrus.Error(err)
			}
		}
	}()
	select {}
}

func createMessage() *nats.Msg {
	headers := make(map[string][]string)
	headers["my-header"] = []string{"my-value-1", "my-value2"}
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	msg := &nats.Msg{
		Subject: "my.subject",
		Data:    []byte(fmt.Sprintf("Message content, sent at %s", currentTime)),
		Header:  headers,
	}
	return msg
}
