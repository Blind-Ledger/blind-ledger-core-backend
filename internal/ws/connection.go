package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/game"
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
		
		// 1) Unpack message
		msgType, payloadRaw, err := UnpackInbound(raw)
		if err != nil {
			log.Printf("❌ Failed to unpack message: %v", err)
			continue
		}
		
		// 2) Process different message types
		switch msgType {
		case TypeJoin:
			c.handleJoin(payloadRaw)
			
		case TypeBet:
			c.handleBet(payloadRaw)
			
		case TypeDistribute:
			c.handleDistribute()
			
		case TypeTournamentAction:
			c.handleTournamentAction(payloadRaw)
			
		default:
			log.Printf("⚠️ Unsupported message type: %s", msgType)
		}
	}
}

func (c *Connection) handleJoin(payloadRaw json.RawMessage) {
	var p InboundPayload
	if err := json.Unmarshal(payloadRaw, &p); err != nil {
		log.Printf("❌ Failed to unmarshal join payload: %v", err)
		return
	}
	
	// Use the new coordinator for join
	state := c.hub.coordinator.LegacyJoin(c.channel, p.Player)
	
	// Send response in existing format for backward compatibility
	out, _ := PackOutbound(TypeUpdate, 1, OutboundPayload{State: state})
	c.hub.Broadcast(c.channel, out)
}

func (c *Connection) handleBet(payloadRaw json.RawMessage) {
	var p InboundPayload
	if err := json.Unmarshal(payloadRaw, &p); err != nil {
		log.Printf("❌ Failed to unmarshal bet payload: %v", err)
		return
	}
	
	// Use the new coordinator for betting
	state, err := c.hub.coordinator.LegacyBet(c.channel, p.Player, p.Amount)
	if err != nil {
		errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: err.Error()})
		c.send(errMsg)
		return
	}
	
	out, _ := PackOutbound(TypeUpdate, 1, OutboundPayload{State: state})
	c.hub.Broadcast(c.channel, out)
}

func (c *Connection) handleDistribute() {
	// For MVP: Auto-complete the tournament
	tournaments, err := c.hub.coordinator.GetActiveTournaments()
	if err != nil {
		log.Printf("❌ Failed to get active tournaments: %v", err)
		return
	}
	
	// Find tournament for this channel (table)
	var targetTournament string
	for _, t := range tournaments {
		if t.TableID == c.channel {
			targetTournament = t.ID
			break
		}
	}
	
	if targetTournament == "" {
		errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: "No active tournament found"})
		c.send(errMsg)
		return
	}
	
	// Auto-complete tournament
	tournament, err := c.hub.coordinator.AutoCompleteTournament(targetTournament)
	if err != nil {
		errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: err.Error()})
		c.send(errMsg)
		return
	}
	
	// Convert to legacy format and broadcast
	state := c.hub.coordinator.convertToLegacyState(tournament)
	out, _ := PackOutbound(TypeUpdate, 1, OutboundPayload{State: state})
	c.hub.Broadcast(c.channel, out)
}

func (c *Connection) handleTournamentAction(payloadRaw json.RawMessage) {
	var p TournamentPayload
	if err := json.Unmarshal(payloadRaw, &p); err != nil {
		log.Printf("❌ Failed to unmarshal tournament payload: %v", err)
		return
	}
	
	switch p.Action {
	case "create":
		c.handleCreateTournament(p)
	case "join":
		c.handleJoinTournament(p)
	case "start":
		c.handleStartTournament(p)
	case "action":
		c.handleGameAction(p)
	default:
		log.Printf("⚠️ Unknown tournament action: %s", p.Action)
	}
}

func (c *Connection) handleCreateTournament(p TournamentPayload) {
	config := game.DefaultSitAndGoConfig() // Use default config for MVP
	tournament, err := c.hub.coordinator.CreateTournament(c.channel, p.PlayerID, config)
	if err != nil {
		errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: err.Error()})
		c.send(errMsg)
		return
	}
	
func (c *Connection) handleGameAction(p TournamentPayload) {
	tournament, err := c.hub.coordinator.ProcessPlayerAction(p.TournamentID, p.PlayerID, p.ActionType, p.Amount)
	if err != nil {
		errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: err.Error()})
		c.send(errMsg)
		return
	}
	
	out, _ := PackOutbound(TypeTournamentUpdate, 1, TournamentOutboundPayload{Tournament: tournament})
	c.hub.Broadcast(c.channel, out)
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
		// cliente lento: cerrar conexión para evitar bloqueo
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
}amentOutboundPayload{Tournament: tournament})
	c.hub.Broadcast(c.channel, out)
}

func (c *Connection) handleJoinTournament(p TournamentPayload) {
	tournament, err := c.hub.coordinator.JoinTournament(p.TournamentID, p.PlayerID, p.PlayerName, p.WalletAddr)
	if err != nil {
		errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: err.Error()})
		c.send(errMsg)
		return
	}
	
	out, _ := PackOutbound(TypeTournamentUpdate, 1, TournamentOutboundPayload{Tournament: tournament})
	c.hub.Broadcast(c.channel, out)
}

func (c *Connection) handleStartTournament(p TournamentPayload) {
	tournament, err := c.hub.coordinator.StartTournament(p.TournamentID)
	if err != nil {
		errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: err.Error()})
		c.send(errMsg)
		return
	}
	
	out, _ := PackOutbound(TypeTournamentUpdate, 1, TournamentOutboundPayload{Tournament: tournament})
	c.hub.Broadcast(c.channel, out)
}

func (c *Connection) handleGameAction(p TournamentPayload) {
	tournament, err := c.hub.coordinator.ProcessPlayerAction(p.TournamentID, p.PlayerID, p.ActionType, p.Amount)
	if err != nil {
		errMsg, _ := PackOutbound(TypeError, 1, OutboundPayload{Error: err.Error()})
		c.send(errMsg)
		return
	}
	
	out, _ := PackOutbound(TypeTournamentUpdate, 1, Tourn