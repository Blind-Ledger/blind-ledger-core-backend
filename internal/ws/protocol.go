package ws

import (
	"encoding/json"

	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/game"
)

// 1. Define el tipo de mensaje
type MessageType string

const (
	TypeJoin       MessageType = "join"
	TypeBet        MessageType = "bet"
	TypeDistribute MessageType = "distribute"
	TypeUpdate     MessageType = "update"
	TypeError      MessageType = "error"
)

// 2. Payloads
// InboundPayload recibe tanto Join, Bet, Distribute
type InboundPayload struct {
	Player string `json:"player,omitempty"`
	Amount int    `json:"amount,omitempty"`
}

// OutboundPayload enviara TableState o errores
type OutboundPayload struct {
	State *game.TableState `json:"state,omitempty"`
	Error string           `json:"error,omitempty"`
}

// 3. Envelope generico
type Envelope struct {
	Type    MessageType     `json:"type"`
	Version int             `json:"version"`
	Payload json.RawMessage `json:"payload"`
}

// 4. Helpers de (de)serializacion
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
