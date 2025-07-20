package ws

import (
	"encoding/json"
	"fmt"

	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/game"
)

// MessageType define los tipos de mensaje
type MessageType string

const (
	// Mensajes legacy (mantener compatibilidad)
	TypeJoin       MessageType = "join"
	TypeBet        MessageType = "bet"
	TypeDistribute MessageType = "distribute"
	TypeUpdate     MessageType = "update"
	TypeError      MessageType = "error"

	// Nuevos mensajes para poker
	TypePokerAction MessageType = "poker_action"
	TypePokerUpdate MessageType = "poker_update"
	TypeGetState    MessageType = "get_state"

	// Mensajes para torneos
	TypeTournamentCreate   MessageType = "tournament_create"
	TypeTournamentRegister MessageType = "tournament_register"
	TypeTournamentStart    MessageType = "tournament_start"
	TypeTournamentUpdate   MessageType = "tournament_update"
	TypeTournamentList     MessageType = "tournament_list"
	TypeTournamentInfo     MessageType = "tournament_info"
)

// InboundPayload para mensajes de entrada
type InboundPayload struct {
	Player string `json:"player,omitempty"`
	Amount int    `json:"amount,omitempty"`

	// Nuevos campos para poker
	Action string `json:"action,omitempty"` // fold, call, raise, all_in

	// Campos para torneos
	TournamentID   string `json:"tournament_id,omitempty"`
	TournamentName string `json:"tournament_name,omitempty"`
	BuyIn          int    `json:"buy_in,omitempty"`
	TournamentType string `json:"tournament_type,omitempty"` // standard, turbo
}

// Validate valida el payload según el tipo de mensaje
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
		if p.Amount > 1000000 {
			return fmt.Errorf("bet amount too large (max 1,000,000)")
		}

	case TypePokerAction:
		if p.Player == "" {
			return fmt.Errorf("player name is required for poker action")
		}
		if p.Action == "" {
			return fmt.Errorf("action is required for poker action")
		}

		// Validar acciones específicas
		switch p.Action {
		case "fold", "call", "all_in":
			// Estas acciones no requieren amount
		case "raise":
			if p.Amount <= 0 {
				return fmt.Errorf("raise amount must be positive")
			}
		default:
			return fmt.Errorf("invalid poker action: %s", p.Action)
		}

	case TypeDistribute, TypeGetState:
		// No requieren validación especial

	case TypeTournamentCreate:
		if p.TournamentID == "" {
			return fmt.Errorf("tournament_id is required")
		}
		if p.TournamentName == "" {
			return fmt.Errorf("tournament_name is required")
		}
		if p.BuyIn <= 0 {
			return fmt.Errorf("buy_in must be positive")
		}

	case TypeTournamentRegister:
		if p.TournamentID == "" {
			return fmt.Errorf("tournament_id is required")
		}
		if p.Player == "" {
			return fmt.Errorf("player name is required")
		}

	case TypeTournamentStart:
		if p.TournamentID == "" {
			return fmt.Errorf("tournament_id is required")
		}

	case TypeTournamentInfo, TypeTournamentList:
		// No requieren validación especial

	default:
		return fmt.Errorf("unknown message type: %s", msgType)
	}
	return nil
}

// OutboundPayload para mensajes de salida
type OutboundPayload struct {
	State *game.TableState `json:"state,omitempty"`
	Error string           `json:"error,omitempty"`

	// Información adicional para poker
	Message     string `json:"message,omitempty"`
	PlayerTurn  string `json:"player_turn,omitempty"`
	GamePhase   string `json:"game_phase,omitempty"`
	ActionValid bool   `json:"action_valid,omitempty"`

	// Información para torneos
	Tournament    interface{} `json:"tournament,omitempty"`
	Tournaments   interface{} `json:"tournaments,omitempty"`
	Registered    bool        `json:"registered,omitempty"`
	PlayersCount  int         `json:"players_count,omitempty"`
	BlindLevel    interface{} `json:"blind_level,omitempty"`
}

// Envelope genérico para todos los mensajes
type Envelope struct {
	Type      MessageType     `json:"type"`
	Version   int             `json:"version"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp int64           `json:"timestamp,omitempty"`
}

// Helpers de serialización mejorados
func PackOutbound(msgType MessageType, version int, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	env := Envelope{
		Type:      msgType,
		Version:   version,
		Payload:   data,
		Timestamp: getCurrentTimestamp(),
	}

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
	case TypeJoin, TypeBet, TypeDistribute, TypePokerAction, TypeGetState,
		 TypeTournamentCreate, TypeTournamentRegister, TypeTournamentStart,
		 TypeTournamentInfo, TypeTournamentList:
		// Tipos válidos
	default:
		return "", nil, fmt.Errorf("unknown message type: %s", env.Type)
	}

	return env.Type, env.Payload, nil
}

// Utility functions
func CreateErrorMessage(errMsg string) ([]byte, error) {
	return PackOutbound(TypeError, 1, OutboundPayload{Error: errMsg})
}

func CreateSuccessMessage(state *game.TableState, message string) ([]byte, error) {
	payload := OutboundPayload{
		State:   state,
		Message: message,
	}

	// Agregar información adicional de poker si está disponible
	if state.PokerTable != nil {
		payload.GamePhase = state.PokerTable.Phase
		if len(state.PokerTable.Players) > state.PokerTable.CurrentPlayer {
			payload.PlayerTurn = state.PokerTable.Players[state.PokerTable.CurrentPlayer].Name
		}
	}

	return PackOutbound(TypeUpdate, 1, payload)
}

func CreatePokerUpdate(state *game.TableState) ([]byte, error) {
	payload := OutboundPayload{
		State: state,
	}

	if state.PokerTable != nil {
		payload.GamePhase = state.PokerTable.Phase
		if len(state.PokerTable.Players) > state.PokerTable.CurrentPlayer {
			payload.PlayerTurn = state.PokerTable.Players[state.PokerTable.CurrentPlayer].Name
		}
	}

	return PackOutbound(TypePokerUpdate, 1, payload)
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return int64(1000) // Placeholder - implementar tiempo real
}
