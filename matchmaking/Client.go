package matchmaking

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
}

func (c *Client) Read() {

	defer func() {
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(1 << 18)                                 // no larger than 256 KB
	_ = c.Conn.SetReadDeadline(time.Now().Add(65 * time.Second)) // closes the connection if idle for 65 seconds

	for {
		_, data, err := c.Conn.ReadMessage() // this is blocking

		if err != nil { // Read error or peer closed
			fmt.Println("closing Read client, ID", c.ID, "%v\n", err)
			break
		}

		// get event type
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
			fmt.Printf("[%s] ping received %s\n", time.Now().Format("15:04:05.000"), data)
		}

		_ = c.Conn.SetReadDeadline(time.Now().Add(65 * time.Second)) // reset the idle timer because we got a message

	}
}

// note we only use this to write because we must not write to a ws connection from multiple goroutines
func (c *Client) Write() {

	for {
		select { // select is blocking btw
		}
	}

}
