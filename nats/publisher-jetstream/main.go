package main

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"log"
	"time"
)

const (
	subjectName = "my-js-subject"
	streamName  = "my-js-stream"
)

func main() {
	nc, err := nats.Connect("ws://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	js, _ := jetstream.New(nc)

	ctx := context.Background()
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subjectName},
	})
	if err != nil {
		log.Fatal(err)
	}

	defer nc.Close()
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			<-ticker.C
			msg := createMessage()
			if _, err := js.PublishMsg(ctx, msg); err != nil {
				log.Fatal(err)
			}
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
		Subject: subjectName,
		Data:    []byte(fmt.Sprintf("Message content, sent at %s", currentTime)),
		Header:  headers,
	}
	return msg
}
