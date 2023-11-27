package main

import (
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	nc, err := nats.Connect("ws://localhost:8080")
	if err != nil {
		logrus.Fatal(err)
	}
	defer nc.Close()
	sub, err := nc.Subscribe("my.subject", func(msg *nats.Msg) {
		printMessage(msg)
	})
	if err != nil {
		log.Fatal(err)
	}
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
	err = sub.Unsubscribe()
	if err != nil {
		return
	}
	time.Sleep(1 * time.Second)
}

func printMessage(msg *nats.Msg) {
	logrus.Info("Received message: ", string(msg.Data))
	/*	values := msg.Header.Values("my-header")
		logrus.Info("First header value: %s", msg.Header.Get("my-header"))
		logrus.Info("First header value: %s", values[0])
		logrus.Info("Second header value: %s", values[0])*/
}
