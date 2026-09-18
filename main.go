package main

import (
	"log"
	"net/http"

	"github.com/warebuf/fd3/server"
)

func main() {

	srv := server.New()

	// matchmaking routes
	http.HandleFunc("/getGuestID", srv.GetGuestIDHandler)
	http.HandleFunc("/matchmakingSSE", srv.MatchmakingSSE)

	// card catalog routes
	http.HandleFunc("/cardDefs", srv.CardDefsHandler)
	http.HandleFunc("/structureDefs", srv.StructureDefsHandler)

	// websocket routes
	http.HandleFunc("/ws/check", srv.WSPreflight)
	http.HandleFunc("/ws", srv.WSHandler)

	log.Fatal(http.ListenAndServe(":3000", nil))

}
