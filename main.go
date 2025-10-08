package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const port = ":3000"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {

		origin := r.Header.Get("Origin")

		switch origin {
		case "http://localhost:5173":
			return true
		case "https://fidemons.com":
			return true
		}

		return false
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}
		fmt.Printf("[%s] Received: %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)

	}
}

func main() {

	http.HandleFunc("/ws", wsHandler)

	log.Fatal(http.ListenAndServe(port, nil))
}
