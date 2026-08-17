package match

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/warebuf/fd3/match/eventpq"
)

type handMessage struct {
	Type      string `json:"Type"`
	CardIndex int    `json:"CardIndex"`
}

type preselectMessage struct {
	Type            string `json:"Type"`
	CardIndex       int    `json:"CardIndex"`
	TargetCardIndex int    `json:"TargetCardIndex"`
	AbilityIndex    int    `json:"AbilityIndex"` // client doesn't set this yet; pre-seeded to -1 below so we can tell it apart from a real 0
}

type abilityMessage struct {
	Type         string `json:"Type"`
	CardIndex    int    `json:"CardIndex"`
	AbilityIndex int    `json:"AbilityIndex"` // -1 when the client named no slot; only a card with one real Active slot resolves without it
}

type readyMessage struct {
	Type string `json:"type"`
}

// spawns per websocket connection
type Client struct {
	ID   string
	Conn *websocket.Conn

	MatchObj *Match

	HighPrio chan []byte // dynamically created as a buffered channel
	LowPrio  chan []byte // dynamically created as a buffered channel

	Done chan struct{}
}

// throws error if message does not match struct exactly
func strictUnmarshal(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func (c *Client) Read() {

	defer func() {
		c.Conn.Close()
		close(c.Done) // signal to close Write()
	}()

	c.Conn.SetReadLimit(1 << 18)                                 // no larger than 256 KB
	_ = c.Conn.SetReadDeadline(time.Now().Add(65 * time.Second)) // closes the connection if idle for 65 seconds

	for {
		_, data, err := c.Conn.ReadMessage() // this is blocking

		if err != nil { // Read error or peer closed
			fmt.Println("closing Read client, ID", c.ID)
			break
		}

		// detect receive time immediately
		receiveTime := time.Now()

		// parse message type
		var msg struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Println("⚠️ invalid JSON:")
			continue
		}

		// handle message types
		switch msg.Type {

		case "p": // for pings
			// append the serverReceive timestamp to the []byte message
			newMsg := data[:len(data)-1] // drop trailing '}'
			newMsg = append(newMsg, []byte(`,"t2":`)...)
			newMsg = strconv.AppendInt(newMsg, receiveTime.UnixMilli(), 10)
			newMsg = append(newMsg, '}')
			// non blocking send
			select {
			case c.HighPrio <- newMsg:
			default:
			}

		case "autosummon":
			var msg handMessage
			if err := strictUnmarshal(data, &msg); err != nil {
				continue
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "autosummon", c.ID, msg.CardIndex, -1, -1)

		case "unautosummon":
			var msg handMessage
			if err := strictUnmarshal(data, &msg); err != nil {
				continue
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "unautosummon", c.ID, msg.CardIndex, -1, -1)

		case "active":
			var msg handMessage
			if err := strictUnmarshal(data, &msg); err != nil {
				continue
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "active", c.ID, msg.CardIndex, -1, -1)

		case "ready":
			var msg readyMessage
			if err := strictUnmarshal(data, &msg); err != nil {
				continue
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "ready", c.ID, -1, -1, -1)

		case "unready":
			ScheduleInputEvent(receiveTime, c.MatchObj, "unready", c.ID, -1, -1, -1)

		case "instant":
			var msg handMessage
			if err := strictUnmarshal(data, &msg); err != nil {
				continue
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "instant", c.ID, msg.CardIndex, -1, -1)

		case "unitAbility":
			var msg handMessage
			if err := strictUnmarshal(data, &msg); err != nil {
				continue
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "unitAbility", c.ID, msg.CardIndex, -1, -1)

		case "trapAbility":
			msg := abilityMessage{AbilityIndex: -1}
			if err := strictUnmarshal(data, &msg); err != nil {
				continue
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "trapAbility", c.ID, msg.CardIndex, -1, msg.AbilityIndex)

		case "preselect":
			msg := preselectMessage{AbilityIndex: -1}
			if err := strictUnmarshal(data, &msg); err != nil {
				continue
			}
			if msg.AbilityIndex == -1 {
				fmt.Println("⚠️ client", c.ID, "didn't set AbilityIndex on preselect, defaulting to 0")
				msg.AbilityIndex = 0
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "preselect", c.ID, msg.CardIndex, msg.TargetCardIndex, msg.AbilityIndex)

		case "toggleAutoAbility":
			var msg handMessage
			if err := strictUnmarshal(data, &msg); err != nil {
				continue
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "toggleAutoAbility", c.ID, msg.CardIndex, -1, -1)

		case "toggleStructureAutoAbility":
			ScheduleInputEvent(receiveTime, c.MatchObj, "toggleStructureAutoAbility", c.ID, -1, -1, -1)

		case "fullPause":
			ScheduleInputEvent(receiveTime, c.MatchObj, "fullPause", c.ID, -1, -1, -1)

		case "structureDraw":
			ScheduleInputEvent(receiveTime, c.MatchObj, "structureDraw", c.ID, -1, -1, -1)

		case "fullState": // send all game states and events occurring within the next 200 ms
			// note: have to worry about race conditions, I am not handling that right now
			c.MatchObj.SendPublishedState(c.ID)
		}

		_ = c.Conn.SetReadDeadline(time.Now().Add(65 * time.Second)) // reset the idle timer because we got a message
	}

}

func (c *Client) Write() {
	for {
		select {
		// high prio first
		case <-c.Done:
			fmt.Println("closing White client, ID", c.ID)
			return
		case data := <-c.HighPrio:
			if err := sendData(c, data); !err {
				return
			}
		default:
			// if no high prio, take first available
			select {
			case <-c.Done:
				fmt.Println("closing White client, ID", c.ID)
				return
			case data := <-c.HighPrio:
				if err := sendData(c, data); !err {
					return
				}
			case data := <-c.LowPrio:
				if err := sendData(c, data); !err {
					return
				}
			}
		}
	}
}

func ScheduleInputEvent(timestamp time.Time, match *Match, eventType string, clientID string, cardIndex int, targetCardIndex int, abilityIndex int) {

	// check if client is in the whitelist
	clientIndex, ok := match.Whitelist[clientID]
	if !ok {
		return
	}

	// this function should only be doing rate limit checks, not sim checks

	ie := &eventpq.InputEvent{
		WallTime:        timestamp,
		Type:            eventType,
		TeamIndex:       int(clientIndex),
		CardIndex:       cardIndex,
		TargetCardIndex: targetCardIndex,
		AbilityIndex:    abilityIndex,
	}

	match.InputEventChan <- ie

}

func sendData(c *Client, data []byte) bool {
	c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		fmt.Println("write error:", err)
		return false
	}
	return true
}
