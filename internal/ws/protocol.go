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

	// Mensajes para lobby/ready system
	TypeSetReady    MessageType = "set_ready"
	TypeStartGame   MessageType = "start_game"
	TypeReadyStatus MessageType = "ready_status"

	// Mensajes para torneos
	TypeTournamentCreate   MessageType = "tournament_create"
	TypeTournamentRegister MessageType = "tournament_register"
	TypeTournamentStart    MessageType = "tournament_start"
	TypeTournamentUpdate   MessageType = "tournament_update"
	TypeTournamentList     MessageType = "tournament_list"
	TypeTournamentInfo     MessageType = "tournament_info"

	// Mensajes para configuración de buy-in
	TypeJoinWithBuyIn    MessageType = "join_with_buy_in"
	TypeGetTableConfig   MessageType = "get_table_config"
	TypeUpdateTableConfig MessageType = "update_table_config"
	TypeValidateBuyIn    MessageType = "validate_buy_in"
)

// InboundPayload para mensajes de entrada
type InboundPayload struct {
	Player string `json:"player,omitempty"`
	Amount int    `json:"amount,omitempty"`

	// Nuevos campos para poker
	Action string `json:"action,omitempty"` // fold, call, raise, all_in

	// Campos para lobby/ready system
	Ready bool `json:"ready,omitempty"` // true/false para set_ready

	// Campos para torneos
	TournamentID   string `json:"tournament_id,omitempty"`
	TournamentName string `json:"tournament_name,omitempty"`
	BuyIn          int    `json:"buy_in,omitempty"`
	TournamentType string `json:"tournament_type,omitempty"` // standard, turbo

	// Campos para configuración de buy-in
	BuyInAmount  int  `json:"buy_in_amount,omitempty"`  // Cantidad de buy-in para join_with_buy_in
	SmallBlind   int  `json:"small_blind,omitempty"`    // Para configuración de mesa
	BigBlind     int  `json:"big_blind,omitempty"`      // Para configuración de mesa
	MinBuyIn     int  `json:"min_buy_in,omitempty"`     // Buy-in mínimo permitido
	MaxBuyIn     int  `json:"max_buy_in,omitempty"`     // Buy-in máximo permitido
	IsCashGame   bool `json:"is_cash_game,omitempty"`   // true = cash game, false = torneo
	AutoRestart  bool `json:"auto_restart,omitempty"`   // Si las manos se reinician automáticamente
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

	case TypeSetReady:
		if p.Player == "" {
			return fmt.Errorf("player name is required for set_ready")
		}

	case TypeStartGame:
		if p.Player == "" {
			return fmt.Errorf("player name is required for start_game")
		}

	case TypeReadyStatus:
		// No requiere validación especial

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

	case TypeJoinWithBuyIn:
		if p.Player == "" {
			return fmt.Errorf("player name is required for join_with_buy_in")
		}
		if len(p.Player) > 50 {
			return fmt.Errorf("player name too long (max 50 chars)")
		}
		if p.BuyInAmount <= 0 {
			return fmt.Errorf("buy_in_amount must be positive")
		}

	case TypeGetTableConfig:
		// No requiere validación especial

	case TypeUpdateTableConfig:
		if p.SmallBlind <= 0 {
			return fmt.Errorf("small_blind must be positive")
		}
		if p.BigBlind <= 0 {
			return fmt.Errorf("big_blind must be positive")
		}
		if p.MinBuyIn <= 0 {
			return fmt.Errorf("min_buy_in must be positive")
		}
		if p.MaxBuyIn <= 0 {
			return fmt.Errorf("max_buy_in must be positive")
		}
		if p.MinBuyIn > p.MaxBuyIn {
			return fmt.Errorf("min_buy_in cannot be greater than max_buy_in")
		}

	case TypeValidateBuyIn:
		if p.BuyInAmount <= 0 {
			return fmt.Errorf("buy_in_amount must be positive")
		}

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

	// Información para lobby/ready system
	ReadyStatus map[string]bool `json:"ready_status,omitempty"`
	IsHost      bool            `json:"is_host,omitempty"`

	// Información para torneos
	Tournament    interface{} `json:"tournament,omitempty"`
	Tournaments   interface{} `json:"tournaments,omitempty"`
	Registered    bool        `json:"registered,omitempty"`
	PlayersCount  int         `json:"players_count,omitempty"`
	BlindLevel    interface{} `json:"blind_level,omitempty"`

	// Información para configuración de buy-in
	TableConfig      interface{} `json:"table_config,omitempty"`      // Configuración completa de la mesa
	BuyInValid       bool        `json:"buy_in_valid,omitempty"`      // Si el buy-in es válido
	ValidationError  string      `json:"validation_error,omitempty"`  // Error de validación si aplica
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
		 TypeSetReady, TypeStartGame, TypeReadyStatus,
		 TypeTournamentCreate, TypeTournamentRegister, TypeTournamentStart,
		 TypeTournamentInfo, TypeTournamentList,
		 TypeJoinWithBuyIn, TypeGetTableConfig, TypeUpdateTableConfig, TypeValidateBuyIn:
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
