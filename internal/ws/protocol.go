package ws

import (
	"encoding/json"

	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/game"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/tournament"
)

// 1. Define el tipo de mensaje
type MessageType string

const (
	TypeJoin             MessageType = "join"
	TypeBet              MessageType = "bet"
	TypeDistribute       MessageType = "distribute"
	TypeUpdate           MessageType = "update"
	TypeError            MessageType = "error"
	TypeTournamentAction MessageType = "tournament_action" // New
	TypeTournamentUpdate MessageType = "tournament_update" // New
)

// 2. Payloads existentes
// InboundPayload recibe tanto Join, Bet, Distribute
type InboundPayload struct {
	Player string `json:"player,omitempty"`
	Amount int    `json:"amount,omitempty"`
}

// OutboundPayload enviará TableState o errores (mantenemos para backward compatibility)
type OutboundPayload struct {
	State *game.LegacyTableState `json:"state,omitempty"`
	Error string                 `json:"error,omitempty"`
}

// 3. Nuevos payloads para torneos
type TournamentPayload struct {
	Action       string `json:"action"` // create, join, start, action
	TournamentID string `json:"tournament_id,omitempty"`
	PlayerID     string `json:"player_id,omitempty"`
	PlayerName   string `json:"player_name,omitempty"`
	WalletAddr   string `json:"wallet_addr,omitempty"`
	ActionType   string `json:"action_type,omitempty"` // fold, call, raise, etc.
	Amount       int    `json:"amount,omitempty"`
}

type TournamentOutboundPayload struct {
	Tournament *tournament.Tournament `json:"tournament,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// 4. Envelope genérico (sin cambios)
type Envelope struct {
	Type    MessageType     `json:"type"`
	Version int             `json:"version"`
	Payload json.RawMessage `json:"payload"`
}

// 5. Helpers de (de)serialización
func PackOutbound(msgType MessageType, version int, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	env := Envelope{Type: msgType, Version: version, Payload: data}
	return json.Marshal(env)
}

func UnpackInbound(raw []byte) (MessageType, json.RawMessage, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return "", nil, err
	}
	return env.Type, env.Payload, nil
}
