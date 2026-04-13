package match

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

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

		case "hand":
			var msg handMessage
			if err := strictUnmarshal(data, &msg); err != nil {
				return
			}
			ScheduleInputEvent(receiveTime, c.MatchObj, "hand", c.ID, msg.CardIndex)

		case "fullState": // send all game states and events occurring within the next 200 ms
			// note: have to worry about race conditions, I am not handling that right now
			fmt.Println("received fullState from", c.ID)
			c.MatchObj.SendPublishedState(c.ID)
		}

		_ = c.Conn.SetReadDeadline(time.Now().Add(65 * time.Second)) // reset the idle timer because we got a message
	}

}

// note we only use this to write because we must not write to a ws connection from multiple goroutines
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

func sendData(c *Client, data []byte) bool {
	c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		fmt.Println("write error:", err)
		return false
	}
	return true
}
