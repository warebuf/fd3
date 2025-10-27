package main

import (
	// std lib
	"fmt"
	"log"
	"net/http"
	"time"

	// 3rd party imports
	"github.com/google/uuid"
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

func getGuestIDHandler(w http.ResponseWriter, r *http.Request) {

	if c, err := r.Cookie("guest_id"); err == nil && c.Value != "" {
		fmt.Println("c", c)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	id := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     "guest_id",
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,                 // true in prod (HTTPS)
		SameSite: http.SameSiteLaxMode, // same-site app; use None for cross-site embeds
		Expires:  time.Now().Add(365 * 24 * time.Hour),
	})
	w.WriteHeader(http.StatusNoContent)

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
		ID:       "a",
		Conn:     conn,
		HighPrio: make(chan []byte, 32),
		LowPrio:  make(chan []byte, 32),
	}

	go c.Write()
	c.Read() // note this is not a goroutine because every request spawns a goroutine
}

func main() {

	fmt.Println("server started")

	http.HandleFunc("/ws", wsHandler)

	log.Fatal(http.ListenAndServe(":3000", nil))
}
