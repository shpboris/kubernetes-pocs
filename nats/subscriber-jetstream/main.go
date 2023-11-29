package main

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	subjectName  = "my-js-subject"
	streamName   = "my-js-stream"
	consumerName = "my-js-consumer"
)

func main() {
	nc, err := nats.Connect("ws://localhost:8080")
	if err != nil {
		logrus.Fatal(err)
	}
	defer nc.Close()
	js, _ := jetstream.New(nc)

	ctx := context.Background()
	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subjectName},
	})
	if err != nil {
		log.Fatal(err)
	}

	cons, _ := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       consumerName,
		FilterSubject: subjectName,
	})
	consumeContext, err := cons.Consume(func(msg jetstream.Msg) {
		printMessage(msg)
		err = msg.Ack()
		if err != nil {
			return
		}
	})
	if err != nil {
		logrus.Println("Subscribe failed")
		return
	}
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
	consumeContext.Stop()
}

func printMessage(msg jetstream.Msg) {
	logrus.Info("Received message: ", string(msg.Data()))
	/*	values := msg.Headers().Values("my-header")
		logrus.Info("First header value: ", msg.Headers().Get("my-header"))
		logrus.Info("First header value: ", values[0])
		logrus.Info("Second header value: ", values[0])*/
}
