package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/poker"
	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/tournament"
)

// Player identifica a un participante (mantener compatibilidad)
type Player struct {
	Name string `json:"name"`
}

// TableState representa el estado de la mesa (mantener compatibilidad)
type TableState struct {
	Host      string   `json:"host"`
	Players   []Player `json:"players"`
	Pot       int      `json:"pot"`
	TurnIndex int      `json:"turnIndex"`

	// Nuevos campos para poker real
	PokerTable *poker.PokerTable `json:"poker_table,omitempty"`
	Phase      string            `json:"phase,omitempty"`
}

// Manager define la interfaz de nuestro gestor de mesas
type Manager interface {
	Join(tableID, playerName string) *TableState
	Bet(tableID, playerName string, amount int) (*TableState, error)
	Distribute(tableId string) (*TableState, error)

	// Nuevos métodos para poker
	PokerAction(tableID, playerName, action string, amount int) (*TableState, error)
	GetTableState(tableID string) (*TableState, error)
	GetTableStateForPlayer(tableID, playerName string) (*TableState, error) // Nuevo: estado filtrado por jugador

	// Métodos para lobby/ready system
	SetPlayerReady(tableID, playerName string, ready bool) (*TableState, error)
	StartGame(tableID, playerName string) (*TableState, error)
	GetReadyStatus(tableID string) (map[string]bool, error)

	// Métodos para torneos
	CreateTournament(tournamentID, name string, buyIn int, tournamentType string) (*tournament.Tournament, error)
	RegisterForTournament(tournamentID, playerID, playerName string) error
	StartTournament(tournamentID string) error
	GetTournament(tournamentID string) (*tournament.Tournament, error)
	ListTournaments() map[string]*tournament.Tournament

	// Métodos para auto-restart
	SetAutoRestart(tableID string, enabled bool, delay time.Duration) error
	GetAutoRestartStatus(tableID string) (bool, time.Duration, error)
	ForceRestartHand(tableID string) error

	// Métodos para configuración de buy-in
	JoinWithBuyIn(tableID, playerName string, buyInAmount int) (*TableState, error)
	GetTableConfig(tableID string) (*poker.TableConfig, error)
	UpdateTableConfig(tableID string, config poker.TableConfig) error
	ValidateBuyIn(tableID string, buyInAmount int) error
}

// managerImpl es la implementación concreta de Manager
type managerImpl struct {
	mu               sync.Mutex
	tables           map[string]*TableState
	pokerEngine      *poker.PokerEngine
	tournamentManager *tournament.Manager
}

// NewManager crea un Manager con poker engine
func NewManager() Manager {
	pokerEngine := poker.NewPokerEngine()
	return &managerImpl{
		tables:            make(map[string]*TableState),
		pokerEngine:       pokerEngine,
		tournamentManager: tournament.NewManager(pokerEngine),
	}
}

func (m *managerImpl) Join(tableID, playerName string) *TableState {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tables[tableID]
	if !ok {
		// Crear nueva tabla
		t = &TableState{
			Host:      playerName,
			TurnIndex: 0,
			Players:   make([]Player, 0),
		}
		m.tables[tableID] = t
	}

	// Evitar duplicados en la lista legacy
	playerExists := false
	for _, p := range t.Players {
		if p.Name == playerName {
			playerExists = true
			break
		}
	}

	if !playerExists {
		t.Players = append(t.Players, Player{Name: playerName})
	}

	// Agregar al poker engine
	playerID := fmt.Sprintf("%s_%s", tableID, playerName)
	pokerTable, err := m.pokerEngine.AddPlayer(tableID, playerID, playerName)
	if err == nil {
		t.PokerTable = pokerTable
		t.Phase = pokerTable.Phase
		t.Pot = pokerTable.Pot

		// Sincronizar TurnIndex con poker engine
		if len(pokerTable.Players) > 0 {
			t.TurnIndex = pokerTable.CurrentPlayer
		}
	}

	return t
}

func (m *managerImpl) Bet(tableID, playerName string, amount int) (*TableState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tables[tableID]
	if !ok {
		return nil, fmt.Errorf("mesa %s no existe", tableID)
	}

	// Usar poker engine si está disponible
	if t.PokerTable != nil {
		return m.pokerActionInternal(tableID, playerName, "call", amount)
	}

	// Fallback a lógica legacy
	if len(t.Players) == 0 {
		return nil, fmt.Errorf("no hay jugadores en la mesa")
	}

	// Validar turno (legacy)
	if t.Players[t.TurnIndex].Name != playerName {
		return t, fmt.Errorf("no es tu turno: turno de %s", t.Players[t.TurnIndex].Name)
	}

	t.Pot += amount
	t.TurnIndex = (t.TurnIndex + 1) % len(t.Players)
	return t, nil
}

func (m *managerImpl) Distribute(tableID string) (*TableState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tables[tableID]
	if !ok {
		return nil, fmt.Errorf("mesa %s no existe", tableID)
	}

	// Usar poker engine si está disponible
	if t.PokerTable != nil {
		// En el poker engine, distribute se maneja automáticamente
		// Solo reiniciamos el estado
		t.Pot = t.PokerTable.Pot
		return t, nil
	}

	// Fallback a lógica legacy
	t.Pot = 0
	t.TurnIndex = 0
	return t, nil
}

// PokerAction - nueva función para acciones de poker específicas
func (m *managerImpl) PokerAction(tableID, playerName, action string, amount int) (*TableState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pokerActionInternal(tableID, playerName, action, amount)
}

func (m *managerImpl) pokerActionInternal(tableID, playerName, action string, amount int) (*TableState, error) {
	t, ok := m.tables[tableID]
	if !ok {
		return nil, fmt.Errorf("mesa %s no existe", tableID)
	}

	if t.PokerTable == nil {
		return nil, fmt.Errorf("poker engine not initialized for table %s", tableID)
	}

	// Ejecutar acción en poker engine
	playerID := fmt.Sprintf("%s_%s", tableID, playerName)
	updatedTable, err := m.pokerEngine.PlayerAction(tableID, playerID, action, amount)
	if err != nil {
		return t, err
	}

	// Actualizar estado legacy
	t.PokerTable = updatedTable
	t.Phase = updatedTable.Phase
	t.Pot = updatedTable.Pot
	t.TurnIndex = updatedTable.CurrentPlayer

	return t, nil
}

func (m *managerImpl) GetTableState(tableID string) (*TableState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tables[tableID]
	if !ok {
		return nil, fmt.Errorf("mesa %s no existe", tableID)
	}

	// Sincronizar con poker engine si está disponible
	if t.PokerTable != nil {
		pokerTable, err := m.pokerEngine.GetTable(tableID)
		if err == nil {
			t.PokerTable = pokerTable
			t.Phase = pokerTable.Phase
			t.Pot = pokerTable.Pot
			t.TurnIndex = pokerTable.CurrentPlayer
		}
	}

	return t, nil
}

// GetTableStateForPlayer obtiene el estado de la mesa con cartas filtradas para un jugador específico
func (m *managerImpl) GetTableStateForPlayer(tableID, playerName string) (*TableState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tables[tableID]
	if !ok {
		return nil, fmt.Errorf("mesa %s no existe", tableID)
	}

	// Crear copia del estado
	filteredState := *t

	// Sincronizar con poker engine usando función filtrada
	if t.PokerTable != nil {
		playerID := fmt.Sprintf("%s_%s", tableID, playerName)
		pokerTable, err := m.pokerEngine.GetTableForPlayer(tableID, playerID)
		if err == nil {
			filteredState.PokerTable = pokerTable
			filteredState.Phase = pokerTable.Phase
			filteredState.Pot = pokerTable.Pot
			filteredState.TurnIndex = pokerTable.CurrentPlayer
		}
	}

	return &filteredState, nil
}

// CreateTournament crea un nuevo torneo
func (m *managerImpl) CreateTournament(tournamentID, name string, buyIn int, tournamentType string) (*tournament.Tournament, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch tournamentType {
	case "standard":
		return m.tournamentManager.CreateStandardTournament(tournamentID, name, buyIn)
	case "turbo":
		return m.tournamentManager.CreateTurboTournament(tournamentID, name, buyIn)
	default:
		return m.tournamentManager.CreateStandardTournament(tournamentID, name, buyIn)
	}
}

// RegisterForTournament registra un jugador en un torneo
func (m *managerImpl) RegisterForTournament(tournamentID, playerID, playerName string) error {
	tournament, err := m.tournamentManager.GetTournament(tournamentID)
	if err != nil {
		return err
	}

	return tournament.RegisterPlayer(playerID, playerName)
}

// StartTournament inicia un torneo manualmente
func (m *managerImpl) StartTournament(tournamentID string) error {
	tournament, err := m.tournamentManager.GetTournament(tournamentID)
	if err != nil {
		return err
	}

	return tournament.StartTournament()
}

// GetTournament obtiene un torneo por ID
func (m *managerImpl) GetTournament(tournamentID string) (*tournament.Tournament, error) {
	return m.tournamentManager.GetTournament(tournamentID)
}

// ListTournaments lista todos los torneos
func (m *managerImpl) ListTournaments() map[string]*tournament.Tournament {
	return m.tournamentManager.ListTournaments()
}

// SetPlayerReady marca a un jugador como listo/no listo
func (m *managerImpl) SetPlayerReady(tableID, playerName string, ready bool) (*TableState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Obtener el estado de la mesa
	t, ok := m.tables[tableID]
	if !ok {
		return nil, fmt.Errorf("mesa %s no existe", tableID)
	}

	if t.PokerTable == nil {
		return nil, fmt.Errorf("poker engine not initialized for table %s", tableID)
	}

	// Ejecutar en poker engine
	playerID := fmt.Sprintf("%s_%s", tableID, playerName)
	updatedTable, err := m.pokerEngine.SetPlayerReady(tableID, playerID, ready)
	if err != nil {
		return t, err
	}

	// Actualizar estado
	t.PokerTable = updatedTable
	t.Phase = updatedTable.Phase

	return t, nil
}

// StartGame inicia el juego (solo por el host)
func (m *managerImpl) StartGame(tableID, playerName string) (*TableState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Obtener el estado de la mesa
	t, ok := m.tables[tableID]
	if !ok {
		return nil, fmt.Errorf("mesa %s no existe", tableID)
	}

	if t.PokerTable == nil {
		return nil, fmt.Errorf("poker engine not initialized for table %s", tableID)
	}

	// Ejecutar en poker engine
	playerID := fmt.Sprintf("%s_%s", tableID, playerName)
	updatedTable, err := m.pokerEngine.StartGame(tableID, playerID)
	if err != nil {
		return t, err
	}

	// Actualizar estado
	t.PokerTable = updatedTable
	t.Phase = updatedTable.Phase
	t.Pot = updatedTable.Pot
	t.TurnIndex = updatedTable.CurrentPlayer

	return t, nil
}

// GetReadyStatus obtiene el estado de ready de todos los jugadores
func (m *managerImpl) GetReadyStatus(tableID string) (map[string]bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pokerEngine.GetReadyStatus(tableID)
}

// SetAutoRestart configura el auto-restart para una mesa
func (m *managerImpl) SetAutoRestart(tableID string, enabled bool, delay time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pokerEngine.SetAutoRestart(tableID, enabled, delay)
}

// GetAutoRestartStatus obtiene el estado del auto-restart para una mesa
func (m *managerImpl) GetAutoRestartStatus(tableID string) (bool, time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pokerEngine.GetAutoRestartStatus(tableID)
}

// ForceRestartHand fuerza el reinicio de una mano
func (m *managerImpl) ForceRestartHand(tableID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pokerEngine.ForceRestartHand(tableID)
}

// JoinWithBuyIn permite a un jugador unirse a una mesa con un buy-in personalizado
func (m *managerImpl) JoinWithBuyIn(tableID, playerName string, buyInAmount int) (*TableState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tables[tableID]
	if !ok {
		// Crear nueva tabla
		t = &TableState{
			Host:      playerName,
			TurnIndex: 0,
			Players:   make([]Player, 0),
		}
		m.tables[tableID] = t
	}

	// Evitar duplicados en la lista legacy
	playerExists := false
	for _, p := range t.Players {
		if p.Name == playerName {
			playerExists = true
			break
		}
	}

	if !playerExists {
		t.Players = append(t.Players, Player{Name: playerName})
	}

	// Agregar al poker engine con buy-in personalizado
	playerID := fmt.Sprintf("%s_%s", tableID, playerName)
	pokerTable, err := m.pokerEngine.AddPlayerWithBuyIn(tableID, playerID, playerName, buyInAmount)
	if err != nil {
		// Si hay error, remover de la lista legacy
		if !playerExists {
			t.Players = t.Players[:len(t.Players)-1]
		}
		return t, err
	}

	// Actualizar estado
	t.PokerTable = pokerTable
	t.Phase = pokerTable.Phase
	t.Pot = pokerTable.Pot

	// Sincronizar TurnIndex con poker engine
	if len(pokerTable.Players) > 0 {
		t.TurnIndex = pokerTable.CurrentPlayer
	}

	return t, nil
}

// GetTableConfig obtiene la configuración de buy-in de una mesa
func (m *managerImpl) GetTableConfig(tableID string) (*poker.TableConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pokerEngine.GetTableConfig(tableID)
}

// UpdateTableConfig actualiza la configuración de buy-in de una mesa
func (m *managerImpl) UpdateTableConfig(tableID string, config poker.TableConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pokerEngine.UpdateTableConfig(tableID, config)
}

// ValidateBuyIn valida si un monto de buy-in es válido para una mesa
func (m *managerImpl) ValidateBuyIn(tableID string, buyInAmount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pokerEngine.ValidateBuyIn(tableID, buyInAmount)
}
