package match

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	clientMessageTypes "github.com/warebuf/fd3/match/clientMessageTypes"
)

// spawns per websocket connection
type Client struct {
	ID   string
	Conn *websocket.Conn

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
		receiveTime := time.Now().UnixMilli()

		// parse message type
		var msg struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Println("⚠️ invalid JSON:")
			continue
		}

		//
		switch msg.Type {
		case "p":
			// append the serverReceive timestamp to the []byte message
			newMsg := data[:len(data)-1] // drop trailing '}'
			newMsg = append(newMsg, []byte(`,"t2":`)...)
			newMsg = strconv.AppendInt(newMsg, receiveTime, 10)
			newMsg = append(newMsg, '}')
			c.HighPrio <- newMsg // TODO: can make this non-blocking, but whatever for now
		case "benchMove":
			var benchMoveData clientMessageTypes.BenchMove
			if err := strictUnmarshal(data, &benchMoveData); err != nil {
				return
			}
			fmt.Println("benchMoveData", benchMoveData) // add it to match event queue
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
