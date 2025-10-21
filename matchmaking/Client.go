package matchmaking

import (
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
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(65 * time.Second)) // resets the idle time when we get a pong
	})

	for {
		_, data, err := c.Conn.ReadMessage() // this is blocking

		if err != nil { // Read error or peer closed
			fmt.Println("closing Read client, ID", c.ID)
			break
		}

		// deal with the data
		fmt.Println("data", data)

		_ = c.Conn.SetReadDeadline(time.Now().Add(65 * time.Second)) // reset the idle timer because we got a message

	}
}

func (c *Client) Write() {

	// init ping ticker
	pingTicker := time.NewTicker(60 * time.Second)
	defer pingTicker.Stop()

	for {
		select { // select is blocking btw
		case <-pingTicker.C: // browser auto-responds with pong
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
