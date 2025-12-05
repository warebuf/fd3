package main

import (
	// std lib
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	// 3rd party imports
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	// our imports
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/match"
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

// maybe over the long-term, when I add users, it should be a ticketID, not a guestID
// this is not encrypted, the user can just edit it if they want, pretend they are someone else
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

	// fmt.Println("matchmakingSSE: IP", networking.GetIP(r))

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
	matchID := make(chan string, 1)
	reject := make(chan struct{}, 1)
	matchmaker.Join <- &matchmaking.Ticket{
		TicketID:         guestID,
		ClosedConnection: ctx.Done(),
		Found:            matchID,
		RejectedTicket:   reject,
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
		case id := <-matchID:
			w.Write([]byte("data: match " + id + "\n\n")) // must end with a blank line
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

func wsPreflight(w http.ResponseWriter, r *http.Request) {

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

	// check if guestID is whitelisted
	matchID := r.URL.Query().Get("matchID")
	match, exists := matchMap[matchID]
	if !exists {
		http.Error(w, "Match ID does not exist", http.StatusUnauthorized)
		return
	}
	teamIdx, exists := match.Whitelist[guestID]
	if !exists {
		http.Error(w, "guestID not whitelisted", http.StatusUnauthorized)
		return
	}

	// write the response
	resp := struct {
		TeamIdx       uint      `json:"teamIdx"`
		NextRoundTime time.Time `json:"nextRoundTime"`
	}{
		TeamIdx:       teamIdx,
		NextRoundTime: match.NextRoundTime,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {

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

	// check if guestID is whitelisted
	matchID := r.URL.Query().Get("matchID")
	matchObj, exists := matchMap[matchID]
	if !exists {
		http.Error(w, "Match ID does not exist", http.StatusUnauthorized)
		return
	}
	if _, exists = matchObj.Whitelist[guestID]; !exists {
		http.Error(w, "guestID not whitelisted", http.StatusUnauthorized)
		return
	}

	// if a connection already exists, close it
	if oldConn, exists := matchObj.ByUser[guestID]; exists {
		oldConn.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "second connection"))
		oldConn.Conn.Close() // triggers Read() error
	}

	// upgrade HTTPS to WSS, create the websocket object
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		http.Error(w, "error creating websocket connection", http.StatusUnauthorized)
		return
	}
	conn.EnableWriteCompression(true)

	c := &match.Client{
		ID:       guestID,
		Conn:     conn,
		MatchObj: matchObj,
		HighPrio: make(chan []byte, 32),
		LowPrio:  make(chan []byte, 32),
		Done:     make(chan struct{}),
	}

	matchObj.ByUser[guestID] = c

	go c.Write()
	c.Read() // note this is not a goroutine because every request spawns a goroutine

}

func getModelData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unitModels)
}

func simStep(w http.ResponseWriter, r *http.Request) {

	var recState *engine.GameState

	if err := json.NewDecoder(r.Body).Decode(&recState); err != nil {
		fmt.Println("Decode error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nextAutoEventTime, success := match.CalcNextAutoEvent(recState)
	if !success {
		return
	}

	incrementedState := match.IncrementState(recState, nextAutoEventTime)

	dueUnits := []int{} // matches length of dueAbilities
	dueAbilities := []int{}
	for i := 0; i < len(incrementedState.AbilityStartIdx); i++ {
		for j := 0; j < incrementedState.AbilityCount[i]; j++ {
			abilityIdx := incrementedState.AbilityStartIdx[i] + j
			isValid := match.CheckValidAbility(incrementedState, i, incrementedState.Abilities[abilityIdx])
			if incrementedState.AbilityCharge[abilityIdx] >= incrementedState.AbilitySPD[abilityIdx] && isValid {
				dueAbilities = append(dueAbilities, abilityIdx)
				dueUnits = append(dueUnits, i)
			}
		}
	}

	finalState, modified := match.HandleDueAbilities(incrementedState, dueUnits, dueAbilities)

	if !modified {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalState)

}

func simMatch(w http.ResponseWriter, r *http.Request) {

	var recState *engine.GameState

	if err := json.NewDecoder(r.Body).Decode(&recState); err != nil {
		fmt.Println("Decode error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	iterationState := engine.CopyGameState(recState.Timestamp, recState)
	for y := 0; y < 100; y++ {

		nextAutoEventTime, success := match.CalcNextAutoEvent(iterationState)
		if !success {
			break
		}

		incrementedState := match.IncrementState(iterationState, nextAutoEventTime)

		dueUnits := []int{} // matches length of dueAbilities
		dueAbilities := []int{}
		for i := 0; i < len(incrementedState.AbilityStartIdx); i++ {
			for j := 0; j < incrementedState.AbilityCount[i]; j++ {
				abilityIdx := incrementedState.AbilityStartIdx[i] + j
				isValid := match.CheckValidAbility(incrementedState, i, incrementedState.Abilities[abilityIdx])
				if incrementedState.AbilityCharge[abilityIdx] >= incrementedState.AbilitySPD[abilityIdx] && isValid {
					dueAbilities = append(dueAbilities, abilityIdx)
					dueUnits = append(dueUnits, i)
				}
			}
		}

		modified := false
		iterationState, modified = match.HandleDueAbilities(incrementedState, dueUnits, dueAbilities)

		if !modified {
			break
		}

	}

	fmt.Println(engine.ConvertStateToHPKey(iterationState))
	fmt.Println(engine.ConvertStateToModelNoKey(iterationState))

}

// global variables
var matchMap map[string]*match.Match
var guestMap *matchmaking.GuestStore
var matchmaker *matchmaking.Matchmaker
var unitModels []engine.Model

func init() {
	fmt.Println("server started")

	// create guest store
	guestMap = matchmaking.NewGuestStore()

	// create matchmaking goroutine
	matchMap = make(map[string]*match.Match)
	unitModels = engine.LoadModelSpecs()
	matchmaker = matchmaking.CreateMatchMaking(matchMap, unitModels)

}

func main() {

	// matchmaking routes
	http.HandleFunc("/getGuestID", getGuestIDHandler)
	http.HandleFunc("/matchmakingSSE", matchmakingSSE)

	// websocket routes
	http.HandleFunc("/ws/check", wsPreflight)
	http.HandleFunc("/ws", wsHandler)

	// simulate routes
	http.HandleFunc("/getModelData", getModelData)
	http.HandleFunc("/simStep", simStep)
	http.HandleFunc("/simMatch", simMatch)

	log.Fatal(http.ListenAndServe(":3000", nil))

}
