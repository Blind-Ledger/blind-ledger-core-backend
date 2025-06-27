package ws

import (
	"log"
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
		// 1) Lee el raw message
		_, raw, err := c.ws.ReadMessage()
		if err != nil {
			log.Printf("❌ WebSocket ReadMessage error (canal %s): %v\n", c.channel, err)
			break
		}

		// 2) Loguea el contenido recibido
		log.Printf("📨 Mensaje crudo recibido en canal %q: %s\n", c.channel, string(raw))

		// 3) Reenvía al hub (y por ende a Redis)
		if err := c.hub.Broadcast(c.channel, raw); err != nil {
			log.Printf("❌ Error al publicar en Redis: %v\n", err)
		}
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
		tableID := mux.Vars(r)["tableId"]
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn := &Connection{
			ws:      wsConn,
			sendCh:  make(chan []byte, 256),
			hub:     hub,
			channel: tableID,
		}
		hub.Register(tableID, conn)
		go conn.writePump()
		go conn.readPump()
	}
}
