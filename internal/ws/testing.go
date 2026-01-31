package ws

// NewTestClient creates a Client without a real websocket connection.
// Intended for use in tests only.
func NewTestClient(hub *Hub, userID int64) *Client {
	return &Client{
		hub:    hub,
		conn:   nil,
		send:   make(chan []byte, 256),
		UserID: userID,
	}
}

// SendChan returns the client's send channel for test assertions.
func (c *Client) SendChan() <-chan []byte {
	return c.send
}
