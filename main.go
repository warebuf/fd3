package main

import (
	// std lib
	"fmt"
	"log"
	"net/http"

	// 3rd party imports
	"github.com/gorilla/websocket"

	// our imports
	"github.com/warebuf/fd3/matchmaking"
)

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

	// upgrade HTTPS to WSS, create the websocket object
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	conn.EnableWriteCompression(true)

	c := &matchmaking.Client{
		ID:   "a",
		Conn: conn,
	}

	go c.Write()
	c.Read() // note this is not a goroutine because every request spawns a goroutine
}

func main() {

	http.HandleFunc("/ws", wsHandler)

	log.Fatal(http.ListenAndServe(":3000", nil))
}
