package ws

import (
	"encoding/json"
	"fmt"

	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/game"
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

// 2. Payloads con validación
type InboundPayload struct {
	Player string `json:"player,omitempty"`
	Amount int    `json:"amount,omitempty"`
}

// Validate valida el payload de entrada
func (p InboundPayload) Validate(msgType MessageType) error {
	switch msgType {
	case TypeJoin:
		if p.Player == "" {
			return fmt.Errorf("player name is required for join")
		}
		if len(p.Player) > 50 {
			return fmt.Errorf("player name too long (max 50 chars)")
		}
	case TypeBet:
		if p.Player == "" {
			return fmt.Errorf("player name is required for bet")
		}
		if p.Amount <= 0 {
			return fmt.Errorf("bet amount must be positive")
		}
		if p.Amount > 1000000 { // Max bet limit
			return fmt.Errorf("bet amount too large (max 1,000,000)")
		}
	case TypeDistribute:
		// No validation needed for distribute
	default:
		return fmt.Errorf("unknown message type: %s", msgType)
	}
	return nil
}

type OutboundPayload struct {
	State *game.TableState `json:"state,omitempty"`
	Error string           `json:"error,omitempty"`
}

// 3. Envelope genérico
type Envelope struct {
	Type    MessageType     `json:"type"`
	Version int             `json:"version"`
	Payload json.RawMessage `json:"payload"`
}

// 4. Helpers de (de)serialización mejorados
func PackOutbound(msgType MessageType, version int, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	env := Envelope{Type: msgType, Version: version, Payload: data}
	result, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal envelope: %w", err)
	}

	return result, nil
}

func UnpackInbound(raw []byte) (MessageType, json.RawMessage, error) {
	if len(raw) == 0 {
		return "", nil, fmt.Errorf("empty message")
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return "", nil, fmt.Errorf("invalid json envelope: %w", err)
	}

	// Validar version
	if env.Version < 1 {
		return "", nil, fmt.Errorf("invalid version: %d", env.Version)
	}

	// Validar tipo de mensaje
	switch env.Type {
	case TypeJoin, TypeBet, TypeDistribute:
		// Tipos válidos
	default:
		return "", nil, fmt.Errorf("unknown message type: %s", env.Type)
	}

	return env.Type, env.Payload, nil
}

// Utility function para crear mensajes de error estándar
func CreateErrorMessage(errMsg string) ([]byte, error) {
	return PackOutbound(TypeError, 1, OutboundPayload{Error: errMsg})
}
