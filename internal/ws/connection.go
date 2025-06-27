package ws

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type Connection struct {
	ws      *websocket.Conn
	sendCh  chan []byte
	hub     *Hub
	channel string
}

func (c *Connection) readPump() {
	defer func() {
		c.hub.Unregister(c.channel, c)
		c.ws.Close()
	}()
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		// Reenvia al hub (y por ende a Redis)
		c.hub.Broadcast(c.channel, msg)
	}
}

func (c *Connection) writePump() {
	for data := range c.sendCh {
		c.ws.WriteMessage(websocket.TextMessage, data)
	}
}

func (c *Connection) send(data []byte) {
	select {
	case c.sendCh <- data:
	default:
		// cliente lento: cerrrar conecion para evitar bloqueo
		c.hub.Unregister(c.channel, c)
		c.ws.Close()
	}
}

func ServeWS(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tableID := vars["tableId"]
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn := &Connection{ws: wsConn, sendCh: make(chan []byte, 256), hub: hub, channel: tableID}
		hub.Register(tableID, conn)
		go conn.writePump()
		go conn.readPump()
	}
}
