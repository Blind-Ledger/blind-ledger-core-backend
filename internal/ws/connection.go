package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

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

	// Configure websocket
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			break
		}

		// 1) Unpack with validation
		msgType, payloadRaw, err := UnpackInbound(raw)
		if err != nil {
			log.Printf("⚠️ Invalid message: %v", err)
			errMsg, _ := CreateErrorMessage(err.Error())
			c.send(errMsg)
			continue
		}

		// 2) Parse payload
		var payload InboundPayload
		if err := json.Unmarshal(payloadRaw, &payload); err != nil {
			log.Printf("⚠️ Invalid payload: %v", err)
			errMsg, _ := CreateErrorMessage("Invalid payload format")
			c.send(errMsg)
			continue
		}

		// 3) Validate payload
		if err := payload.Validate(msgType); err != nil {
			log.Printf("⚠️ Payload validation failed: %v", err)
			errMsg, _ := CreateErrorMessage(err.Error())
			c.send(errMsg)
			continue
		}

		// 4) Process message
		c.handleMessage(msgType, payload)
	}
}

func (c *Connection) handleMessage(msgType MessageType, payload InboundPayload) {
	switch msgType {
	case TypeJoin:
		c.handleJoin(payload)
	case TypeBet:
		c.handleBet(payload)
	case TypeDistribute:
		c.handleDistribute()
	default:
		log.Printf("⚠️ Unhandled message type: %s", msgType)
	}
}

func (c *Connection) handleJoin(payload InboundPayload) {
	log.Printf("👤 Player %s joining table %s", payload.Player, c.channel)

	state := c.hub.mgr.Join(c.channel, payload.Player)

	out, err := PackOutbound(TypeUpdate, 1, OutboundPayload{State: state})
	if err != nil {
		log.Printf("❌ Failed to pack join response: %v", err)
		errMsg, _ := CreateErrorMessage("Internal server error")
		c.send(errMsg)
		return
	}

	c.hub.Broadcast(c.channel, out)
}

func (c *Connection) handleBet(payload InboundPayload) {
	log.Printf("💰 Player %s betting %d on table %s", payload.Player, payload.Amount, c.channel)

	state, err := c.hub.mgr.Bet(c.channel, payload.Player, payload.Amount)
	if err != nil {
		log.Printf("⚠️ Bet failed: %v", err)
		errMsg, _ := CreateErrorMessage(err.Error())
		c.send(errMsg)
		return
	}

	out, err := PackOutbound(TypeUpdate, 1, OutboundPayload{State: state})
	if err != nil {
		log.Printf("❌ Failed to pack bet response: %v", err)
		errMsg, _ := CreateErrorMessage("Internal server error")
		c.send(errMsg)
		return
	}

	c.hub.Broadcast(c.channel, out)
}

func (c *Connection) handleDistribute() {
	log.Printf("🎯 Distributing pot for table %s", c.channel)

	state, err := c.hub.mgr.Distribute(c.channel)
	if err != nil {
		log.Printf("⚠️ Distribute failed: %v", err)
		errMsg, _ := CreateErrorMessage(err.Error())
		c.send(errMsg)
		return
	}

	out, err := PackOutbound(TypeUpdate, 1, OutboundPayload{State: state})
	if err != nil {
		log.Printf("❌ Failed to pack distribute response: %v", err)
		errMsg, _ := CreateErrorMessage("Internal server error")
		c.send(errMsg)
		return
	}

	c.hub.Broadcast(c.channel, out)
}

func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.sendCh:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("❌ Write error: %v", err)
				return
			}

		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Connection) send(data []byte) {
	select {
	case c.sendCh <- data:
	default:
		log.Printf("⚠️ Send buffer full, closing connection")
		c.hub.Unregister(c.channel, c)
		close(c.sendCh)
	}
}

func ServeWS(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tableID := mux.Vars(r)["tableId"]

		// Validate table ID
		if tableID == "" {
			http.Error(w, "Table ID is required", http.StatusBadRequest)
			return
		}

		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("❌ WebSocket upgrade failed: %v", err)
			return
		}

		conn := &Connection{
			ws:      wsConn,
			sendCh:  make(chan []byte, 256),
			hub:     hub,
			channel: tableID,
		}

		log.Printf("✅ New WebSocket connection for table: %s", tableID)
		hub.Register(tableID, conn)

		// Start pumps
		go conn.writePump()
		go conn.readPump()
	}
}
