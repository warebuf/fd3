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
	"github.com/warebuf/fd3/networking"
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

// maybe over the long-term, when I add users, it should be a ticketID, not a guestID
// TODO: this is not encrypted, the user can just edit it if they want, pretend they are someone else
func getGuestIDHandler(w http.ResponseWriter, r *http.Request) {

	id := uuid.NewString()
	expiry := time.Now().Add(2 * time.Hour)

	guestMap.Add(id, expiry)

	http.SetCookie(w, &http.Cookie{
		Name:     "guestID",
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,                 // true in prod (HTTPS)
		SameSite: http.SameSiteLaxMode, // same-site app; use None for cross-site embeds
		Expires:  expiry,
	})
	w.WriteHeader(http.StatusNoContent)

}

func matchmakingSSE(w http.ResponseWriter, r *http.Request) {

	fmt.Println("matchmakingSSE: IP", networking.GetIP(r))

	guestID := ""
	for _, c := range r.Cookies() {
		if c.Name == "guestID" {
			guestID = c.Value
		}
	}

	// check for cookie
	if guestID == "" {
		http.Error(w, "Invalid guestID", http.StatusUnauthorized)
		return
	}

	// check for expired cookie
	ok := guestMap.Get(guestID)
	if !ok {
		http.Error(w, "Expired guestID", http.StatusUnauthorized)
		return
	}

	// header set up
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// check if the server will allow us the force the SSE output
	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Println("Cannot Flush! Please hard investigate.")
		return
	}

	ctx := r.Context() // close when client disconnects

	// add the guest/ticket to the matchmaking queue
	match := make(chan string, 1)
	reject := make(chan struct{}, 1)
	matchmaker.Join <- &matchmaking.Ticket{
		TicketID: guestID,
		ClosedConnection:    ctx.Done(),
		Found:    match,
		Reject:   reject,
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // is a closed channel, anyone who checks it will see it is closed
			fmt.Println("Close Signal")
			guestMap.Remove(guestID)
			return
		case <-reject:
			fmt.Println("reject")
			return
		case <-match:
			w.Write([]byte("data: match\n\n")) // must end with a blank line
			flusher.Flush()
			return
		case <-ticker.C:
			if _, err := w.Write([]byte("data: ping\n\n")); err != nil { // must end with a blank line
				return
			}
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

// global variables
var guestMap *matchmaking.GuestStore
var matchmaker *matchmaking.Matchmaker

func init() {
	fmt.Println("server started")
	guestMap = matchmaking.NewGuestStore()
	matchmaker = matchmaking.CreateMatchMaking()
}

func main() {

	// matchmaking routes
	http.HandleFunc("/getGuestID", getGuestIDHandler)
	http.HandleFunc("/matchmakingSSE", matchmakingSSE)

	// websocket routes
	http.HandleFunc("/ws", wsHandler)

	log.Fatal(http.ListenAndServe(":3000", nil))

}
