package matchmaking

import (
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

	HighPrio chan []byte // dynamically created as a buffered channel
	LowPrio  chan []byte // dynamically created as a buffered channel

	Done chan struct{}
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
			fmt.Println("closing Read client, ID", c.ID, "%v\n", err)
			break
		}

		// detect receive time immediately
		t2 := time.Now().UnixMilli()

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
			// append the t2 timestamp to the []byte message
			newMsg := data[:len(data)-1] // drop trailing '}'
			newMsg = append(newMsg, []byte(`,"t2":`)...)
			newMsg = strconv.AppendInt(newMsg, t2, 10)
			newMsg = append(newMsg, '}')
			c.HighPrio <- newMsg // can make this non-blocking, but whatever for now
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
			return
		case data := <-c.HighPrio:
			if err := sendData(c, data); err == false {
				return
			}
		default:
			// if no high prio, take first available
			select {
			case <-c.Done:
				return
			case data := <-c.HighPrio:
				if err := sendData(c, data); err == false {
					return
				}
			case data := <-c.LowPrio:
				if err := sendData(c, data); err == false {
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
