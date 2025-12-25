package main

import (
	// std lib
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	// 3rd party imports
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	// our imports
	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/summon"
	"github.com/warebuf/fd3/match"
	"github.com/warebuf/fd3/matchmaking"
	"github.com/warebuf/fd3/performance"
	"github.com/warebuf/fd3/utility/compressState"
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

// TODO:
// these 3 functions need to be modularized (maybe added to the performance lib) and adapted to the auto, manual, passive model
func simStep(w http.ResponseWriter, r *http.Request) {

	var recState engine.GameState

	if err := json.NewDecoder(r.Body).Decode(&recState); err != nil {
		fmt.Println("Decode error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nextAutoEventTime, success := match.CalcNextAutoEvent(&recState)
	if !success {
		return
	}

	incrementedState := match.IncrementState(&recState, nextAutoEventTime)

	dueUnits := []int{} // matches length of dueFunctions
	dueFunctions := []int{}
	for i := 0; i < len(incrementedState.FunctionStartIdx); i++ {
		for j := 0; j < incrementedState.FunctionCount[i]; j++ {
			funcIdx := incrementedState.FunctionStartIdx[i] + j
			if incrementedState.FunctionType[j] == 0 { // only process auto funcs
				isValid := match.CheckValidFunction(incrementedState, i, incrementedState.Functions[funcIdx])
				if incrementedState.FunctionCharge[funcIdx] >= incrementedState.FunctionSPD[funcIdx] && isValid {
					dueFunctions = append(dueFunctions, funcIdx)
					dueUnits = append(dueUnits, i)
				}
			}

		}
	}

	finalState, modified := match.HandleDueFunctions(incrementedState, dueUnits, dueFunctions)

	if !modified {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalState)

}

func simMatch(w http.ResponseWriter, r *http.Request) {

	var recState engine.GameState

	if err := json.NewDecoder(r.Body).Decode(&recState); err != nil {
		fmt.Println("Decode error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	iterationState := engine.CopyGameState(recState.Timestamp, &recState)
	y := 0
	for y = 0; y < 100; y++ {

		nextAutoEventTime, success := match.CalcNextAutoEvent(iterationState)
		if !success {
			break
		}

		incrementedState := match.IncrementState(iterationState, nextAutoEventTime)

		dueUnits := []int{} // matches length of dueAbilities
		dueFunctions := []int{}
		for i := 0; i < len(incrementedState.FunctionStartIdx); i++ {
			for j := 0; j < incrementedState.FunctionCount[i]; j++ {
				funcIdx := incrementedState.FunctionStartIdx[i] + j
				if incrementedState.FunctionType[funcIdx] == 0 { // only process auto funcs
					isValid := match.CheckValidFunction(incrementedState, i, incrementedState.Functions[funcIdx])
					if incrementedState.FunctionCharge[funcIdx] >= incrementedState.FunctionSPD[funcIdx] && isValid {
						dueFunctions = append(dueFunctions, funcIdx)
						dueUnits = append(dueUnits, i)
					}
				}
			}
		}

		modified := false
		iterationState, modified = match.HandleDueFunctions(incrementedState, dueUnits, dueFunctions)

		if !modified {
			break
		}

	}

	fmt.Println("iterations", y)
	fmt.Println(engine.ConvertStateToHPKey(iterationState))
	fmt.Println(engine.ConvertStateToModelNoKey(iterationState))

}

func simProb(w http.ResponseWriter, r *http.Request) {

	var recState engine.GameState

	if err := json.NewDecoder(r.Body).Decode(&recState); err != nil {
		fmt.Println("Decode error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	HPMap := make(map[string]int)
	ModelMap := make(map[string]int)
	IterationMap := make(map[string]int)

	iterations := 100000
	rounds := 50

	for x := 0; x < iterations; x++ {

		iterationState := engine.CopyGameState(recState.Timestamp, &recState)
		y := 0
		for y = 0; y < rounds; y++ {

			nextAutoEventTime, success := match.CalcNextAutoEvent(iterationState)
			if !success {
				break
			}

			incrementedState := match.IncrementState(iterationState, nextAutoEventTime)

			dueUnits := []int{} // matches length of dueFunctions
			dueFunctions := []int{}
			for i := 0; i < len(incrementedState.FunctionStartIdx); i++ {
				for j := 0; j < incrementedState.FunctionCount[i]; j++ {
					funcIdx := incrementedState.FunctionStartIdx[i] + j
					if incrementedState.FunctionType[funcIdx] == 0 { // only process auto funcs
						isValid := match.CheckValidFunction(incrementedState, i, incrementedState.Functions[funcIdx])
						if incrementedState.FunctionCharge[funcIdx] >= incrementedState.FunctionSPD[funcIdx] && isValid {
							dueFunctions = append(dueFunctions, funcIdx)
							dueUnits = append(dueUnits, i)
						}
					}
				}
			}

			modified := false
			iterationState, modified = match.HandleDueFunctions(incrementedState, dueUnits, dueFunctions)

			if !modified {
				break
			}

		}
		HPKey := engine.ConvertStateToHPKey(iterationState)
		ModelKey := engine.ConvertStateToModelNoKey(iterationState)
		HPMap[HPKey]++
		ModelMap[ModelKey]++
		IterationMap[ModelKey] += y
		//fmt.Println(HPKey)
		//fmt.Println("y", y)
	}

	// printing out the states in order
	keys := make([]string, 0, len(ModelMap))
	for k, v := range ModelMap {
		if v == 0 {
			continue
		}
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		ai := float64(ModelMap[keys[i]]) / float64(iterations)
		aj := float64(ModelMap[keys[j]]) / float64(iterations)
		return ai > aj // descending
	})

	for _, k := range keys {
		value := ModelMap[k]
		avg := float64(value) / float64(iterations)

		compressed := compressState.CompressPattern(k)
		//iterAvg := float64(IterationMap[k]) / float64(value)
		//fmt.Printf("%s %.1f%% iterations avg: %.2f\n", k, avg*100, iterAvg)
		if avg*100 >= 1 {
			fmt.Printf("%.0f%% %s \n", avg*100, compressed)
		} else {
			fmt.Printf("%.1f%% %s \n", avg*100, compressed)
		}
	}
}

func simAll(w http.ResponseWriter, r *http.Request) {

	iterations := 10000
	rounds := 50

	combos := performance.GenerateCombos(10, unitModels) // generate all possible combinatons
	teamsWithDups := performance.GenerateTeams(combos)   // generate all possible teams
	teams := performance.RemoveDups(teamsWithDups)       // sort each arr, remove dups

	lineupToAns := make(map[string]string)

	for i := 0; i < len(teams); i++ {

		// create the state
		state := engine.CreateFirstGameState(time.Now(), 0)

		key1 := compressState.IntToAlphaNumeric(teams[i][0])
		key2 := compressState.IntToAlphaNumeric(teams[i][1])
		key := key1 + ";" + key2
		fmt.Println(i, key)

		// summon units
		for j := 0; j < len(teams[i][0]); j++ {
			state = summon.SpawnSummon(unitModels, state, 0, teams[i][0][j])
		}
		for j := 0; j < len(teams[i][1]); j++ {
			state = summon.SpawnSummon(unitModels, state, 1, teams[i][1][j])
		}

		HPMap := make(map[string]int)
		ModelMap := make(map[string]int)
		IterationMap := make(map[string]int)

		for x := 0; x < iterations; x++ {
			iterationState := engine.CopyGameState(state.Timestamp, state)
			y := 0
			for y = 0; y < rounds; y++ {

				nextAutoEventTime, success := match.CalcNextAutoEvent(iterationState)
				if !success {
					break
				}

				incrementedState := match.IncrementState(iterationState, nextAutoEventTime)

				dueUnits := []int{} // matches length of dueAbilities
				dueAbilities := []int{}
				for i := 0; i < len(incrementedState.FunctionStartIdx); i++ {
					for j := 0; j < incrementedState.FunctionCount[i]; j++ {
						funcIdx := incrementedState.FunctionStartIdx[i] + j
						if incrementedState.FunctionType[funcIdx] == 0 {
							isValid := match.CheckValidFunction(incrementedState, i, incrementedState.Functions[funcIdx])
							if incrementedState.FunctionCharge[funcIdx] >= incrementedState.FunctionSPD[funcIdx] && isValid {
								dueAbilities = append(dueAbilities, funcIdx)
								dueUnits = append(dueUnits, i)
							}
						}
					}
				}

				modified := false
				iterationState, modified = match.HandleDueFunctions(incrementedState, dueUnits, dueAbilities)

				if !modified {
					break
				}

			}
			HPKey := engine.ConvertStateToHPKey(iterationState)
			ModelKey := engine.ConvertStateToModelNoKey(iterationState)
			HPMap[HPKey]++
			ModelMap[ModelKey]++
			IterationMap[ModelKey] += y
			//fmt.Println(HPKey)
			//fmt.Println("y", y)
		}

		// printing out the states in order of percent
		keys := make([]string, 0, len(ModelMap))
		for k, v := range ModelMap {
			if v == 0 {
				continue
			}
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			ai := float64(ModelMap[keys[i]]) / float64(iterations)
			aj := float64(ModelMap[keys[j]]) / float64(iterations)
			return ai > aj // descending
		})

		var ans strings.Builder
		for _, k := range keys {
			value := ModelMap[k]
			avg := float64(value) / float64(iterations)

			compressed := compressState.CompressPattern(k)
			if avg*100 >= 1 {
				fmt.Fprintf(&ans, "%.0f%% %s\n", avg*100, compressed)
				fmt.Printf("%.0f%% %s \n", avg*100, compressed)
			} else {
				//fmt.Fprintf(&ans, "%.1f%% %s\n", avg*100, compressed)
				fmt.Printf("%.1f%% %s \n", avg*100, compressed)
			}
		}
		lineupToAns[key] = ans.String()

	}

	compressState.Excel(teams, lineupToAns)

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
	http.HandleFunc("/simProb", simProb)
	http.HandleFunc("/simAll", simAll)

	log.Fatal(http.ListenAndServe(":3000", nil))

}
