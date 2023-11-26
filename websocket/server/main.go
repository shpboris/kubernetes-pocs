package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/mymsg", processMessage)
	port := 8080
	logrus.Print("Server running on: ", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		logrus.Fatal("Error starting server", err)
	}
}

func processMessage(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			logrus.Error(err)
			return
		}
		modifiedMessage := []byte("Processed: " + string(p))
		if err := conn.WriteMessage(messageType, modifiedMessage); err != nil {
			logrus.Println(err)
			return
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
