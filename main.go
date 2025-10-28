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

// TODO: this is not encrypted, the user can just edit it if they want
// TODO: figure out what happens when the cookie expires
func getGuestIDHandler(w http.ResponseWriter, r *http.Request) {

	// basically every game, give the user a new guest cookie
	if c, err := r.Cookie("guestID"); err == nil && c.Value != "" {
		fmt.Println("c", c)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	id := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     "guestID",
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,                 // true in prod (HTTPS)
		SameSite: http.SameSiteLaxMode, // same-site app; use None for cross-site embeds
		Expires:  time.Now().Add(365 * 24 * time.Hour),
	})
	w.WriteHeader(http.StatusNoContent)

}

func matchmakingSSE(w http.ResponseWriter, r *http.Request) {

	fmt.Println("matchmakingSSE")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Println("Cannot Flush! Please hard investigate.")
		return
	}

	ctx := r.Context() // close when client disconnects

	matchCh := make(chan struct{}, 1)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-matchCh:
			w.Write([]byte("match"))
			flusher.Flush()
			return
		case <-ticker.C:
			fmt.Println("matchmaking ping")
			w.Write([]byte("ping"))
			flusher.Flush()
		}
	}

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
	http.HandleFunc("/matchmakingSSE", matchmakingSSE)
	http.HandleFunc("/getGuestID", getGuestIDHandler)

	log.Fatal(http.ListenAndServe(":3000", nil))
}
