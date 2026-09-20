package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/warebuf/fd3/decks"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/match"
	"github.com/warebuf/fd3/matchmaking"
)

type Server struct {
	Upgrader   websocket.Upgrader
	MatchMap   map[string]*match.Match
	GuestMap   *matchmaking.GuestStore
	Matchmaker *matchmaking.Matchmaker
	SavedDecks *decks.Store
}

func New() *Server {
	fmt.Println("server started")

	s := &Server{
		Upgrader: websocket.Upgrader{
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
		},
		MatchMap:   make(map[string]*match.Match),
		GuestMap:   matchmaking.NewGuestStore(),
		SavedDecks: decks.NewStore("data/savedDecks.json"),
	}

	// create matchmaking goroutine
	s.Matchmaker = matchmaking.CreateMatchMaking(s.MatchMap)

	return s
}

type getGuestIDBody struct {
	Deck         []int `json:"deck"`
	StartingHand []int `json:"startingHand"`
	StructNo     int   `json:"structNo"`
}

// maybe over the long-term, when I add users, it should be a ticketID, not a guestID
// this is not encrypted, the user can just edit it if they want, pretend they are someone else
func (s *Server) GetGuestIDHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "non-POST method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check for valid deck
	var reqBody getGuestIDBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !matchmaking.IsValidSubmission(reqBody.Deck, reqBody.StartingHand, reqBody.StructNo) {
		http.Error(w, "invalid deck", http.StatusBadRequest)
		return
	}

	id := uuid.NewString()
	expiry := time.Now().Add(2 * time.Hour)

	s.GuestMap.Add(id, expiry, reqBody.Deck, reqBody.StartingHand, reqBody.StructNo)

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

func (s *Server) MatchmakingSSE(w http.ResponseWriter, r *http.Request) {

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
	guest, ok := s.GuestMap.Get(guestID)
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
	s.Matchmaker.Join <- &matchmaking.Ticket{
		TicketID:         guestID,
		Deck:             guest.Deck,
		StartingHand:     guest.StartingHand,
		StructureDefNo:   guest.StructNo,
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
			s.GuestMap.Remove(guestID)
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

func (s *Server) WSPreflight(w http.ResponseWriter, r *http.Request) {

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
	_, ok := s.GuestMap.Get(guestID)
	if !ok {
		http.Error(w, "Expired guestID", http.StatusUnauthorized)
		return
	}

	// check if guestID is whitelisted
	matchID := r.URL.Query().Get("matchID")
	matchObj, exists := s.MatchMap[matchID]
	if !exists {
		http.Error(w, "Match ID does not exist", http.StatusUnauthorized)
		return
	}
	teamIndex, exists := matchObj.Whitelist[guestID]
	if !exists {
		http.Error(w, "guestID not whitelisted", http.StatusUnauthorized)
		return
	}

	// write the response
	resp := struct {
		TeamIndex uint `json:"teamIndex"`
	}{
		TeamIndex: teamIndex,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) WSHandler(w http.ResponseWriter, r *http.Request) {

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
	_, ok := s.GuestMap.Get(guestID)
	if !ok {
		http.Error(w, "Expired guestID", http.StatusUnauthorized)
		return
	}

	// check if guestID is whitelisted
	matchID := r.URL.Query().Get("matchID")
	matchObj, exists := s.MatchMap[matchID]
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
	conn, err := s.Upgrader.Upgrade(w, r, nil)
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

// the fields a deck-builder UI needs to browse and pick cards; intentionally leaves out ability
// internals (Cons, TargetFlags, Effects) since Description already covers the mechanics in prose
type cardDefResponse struct {
	DefNo       int    `json:"DefNo"`
	Type        int    `json:"Type"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Archetype   int    `json:"Archetype"`
	Taunt       bool   `json:"Taunt"`
	Bodyguard   bool   `json:"Bodyguard"`
	IsToken     bool   `json:"IsToken"`
	MaxCharges  int    `json:"MaxCharges"`
	BaseHP      int    `json:"BaseHP"`
	Cost        int    `json:"Cost"`
}

// public, read-only card catalog - no guest session required
func (s *Server) CardDefsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "non-GET method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := make([]cardDefResponse, 0, len(defs.CardDefs))
	for defNo, def := range defs.CardDefs {
		if defNo == 0 || defs.IsRetired(defNo) { // placeholder so DefNo lines up with index, or a removed card
			continue
		}
		resp = append(resp, cardDefResponse{
			DefNo:       defNo,
			Type:        def.Type,
			Name:        def.Name,
			Description: def.Description,
			Archetype:   def.Archetype,
			Taunt:       def.Taunt,
			Bodyguard:   def.Bodyguard,
			IsToken:     def.IsToken,
			MaxCharges:  def.MaxCharges,
			BaseHP:      def.BaseHP,
			Cost:        def.Cost,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type structureDefResponse struct {
	DefNo       int    `json:"DefNo"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	BaseHP      int    `json:"BaseHP"`
	DrawCost    int    `json:"DrawCost"`
}

// public, read-only structure catalog - no guest session required
func (s *Server) StructureDefsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "non-GET method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := make([]structureDefResponse, 0, len(defs.StructureDefs))
	for defNo, def := range defs.StructureDefs {
		if defNo == 0 {
			continue
		}
		resp = append(resp, structureDefResponse{
			DefNo:       defNo,
			Name:        def.Name,
			Description: def.Description,
			BaseHP:      def.BaseHP,
			DrawCost:    def.DrawCost,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type saveDeckBody struct {
	Name     string `json:"name"`
	Deck     []int  `json:"deck"`
	StructNo int    `json:"structNo"`
}

// SavedDecksHandler serves a shared, public list of named decks: GET lists them, POST adds one.
// Anyone can read or add; nothing here is scoped to a guest session.
func (s *Server) SavedDecksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.SavedDecks.List())

	case http.MethodPost:
		var reqBody saveDeckBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if !matchmaking.IsValidSubmission(reqBody.Deck, nil, reqBody.StructNo) {
			http.Error(w, "invalid deck", http.StatusBadRequest)
			return
		}

		entry, err := s.SavedDecks.Add(reqBody.Name, reqBody.Deck, reqBody.StructNo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entry)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
