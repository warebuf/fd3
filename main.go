package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const port = ":3000"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {

		origin := r.Header.Get("Origin")

		switch origin {
		case "http://localhost" + port:
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
		log.Println("upgrade:", err)
		return
	}

	defer conn.Close()

}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/ws", wsHandler)

	log.Fatal(http.ListenAndServe(port, nil))
}
