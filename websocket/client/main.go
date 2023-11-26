package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	url := "ws://172.28.71.204:80/mymsg"
	go connectWebSocket(url)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
		logrus.Print("Interrupt signal received, closing connections...")
	}
}

func connectWebSocket(url string) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer c.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				return
			}
			fmt.Printf("Received message: %s\n", message)
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			message := []byte(fmt.Sprintf("Message content, sent at %s", t.Format(time.StampMilli)))
			err := c.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error writing message:", err)
				return
			}
			fmt.Printf("Sent message: %s\n", message)
		}
	}
}
