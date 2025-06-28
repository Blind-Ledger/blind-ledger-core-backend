package ws

import (
	"encoding/json"
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
		_, raw, err := c.ws.ReadMessage()
		if err != nil {
			log.Printf("❌ ReadMessage error: %v\n", err)
			break
		}
		// 1) Unpack
		msgType, payloadRaw, err := UnpackInbound(raw)
		if err != nil {
			// opcional: enviar PackOutbound(TypeError,…)
			continue
		}
		// 2) Switch sobre msgType
		switch msgType {
		case TypeJoin:
			var p InboundPayload
			json.Unmarshal(payloadRaw, &p)
			state := c.hub.mgr.Join(c.channel, p.Player)
			out, _ := PackOutbound(TypeUpdate, 1, OutboundPayload{State: state})
			c.hub.Broadcast(c.channel, out)

		case TypeBet:
			var b InboundPayload
			json.Unmarshal(payloadRaw, &b)
			state, err := c.hub.mgr.Bet(c.channel, b.Player, b.Amount)
			if err != nil {
				errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: err.Error()})
				c.send(errMsg)
				continue
			}
			out, _ := PackOutbound(TypeUpdate, 1, OutboundPayload{State: state})
			c.hub.Broadcast(c.channel, out)

		case TypeDistribute:
			state, err := c.hub.mgr.Distribute(c.channel)
			if err != nil {
				errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: err.Error()})
				c.send(errMsg)
				continue
			}
			out, _ := PackOutbound(TypeUpdate, 1, OutboundPayload{State: state})
			c.hub.Broadcast(c.channel, out)

		default:
			// opcional: enviar error de tipo no soportado
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
