package game

import (
	"fmt"
	"sync"
)

// Player identifica a un participante.
type Player struct {
	Name string `json:"name"`
}

// TableState representa el estado de la mesa en cada etapa
type TableState struct {
	Host      string   `json:"host"`      // quien reparte
	Players   []Player `json:"players"`   // orden de turnos
	Pot       int      `json:"pot"`       // bote acumulado
	TurnIndex int      `json:"turnIndex"` // indice del jugador actual
}

// Manager define la interfaz de nuestro gestor de mesas.
type Manager interface {
	Join(tableID, playerName string) *TableState
	Bet(tableID, playerName string, amount int) (*TableState, error)
	Distribute(tableId string) (*TableState, error)
}

// manegerImpl es la implementacion concreta de Manager.
type managerImpl struct {
	mu     sync.Mutex
	tables map[string]*TableState
}

// NewManagrer crea un Manager limpio.
func NewManager() Manager {
	return &managerImpl{
		tables: make(map[string]*TableState),
	}
}

func (m *managerImpl) Join(tableID, playerName string) *TableState {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tables[tableID]
	if !ok {
		// La primera vez, el creador es host y turno 0
		t = &TableState{Host: playerName, TurnIndex: 0}
		m.tables[tableID] = t
	}
	// Evitar duplicados
	for _, p := range t.Players {
		if p.Name == playerName {
			return t
		}
	}
	t.Players = append(t.Players, Player{Name: playerName})
	return t
}

func (m *managerImpl) Bet(tableID, playerName string, amount int) (*TableState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tables[tableID]
	if !ok {
		return nil, fmt.Errorf("mesa %s no existe", tableID)
	}
	// Validar turno
	if t.Players[t.TurnIndex].Name != playerName {
		return t, fmt.Errorf("no es tu turno: turno de %s", t.Players[t.TurnIndex].Name)
	}
	t.Pot += amount // Avanzar turno ciclicamente
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
	// Logica simple: host se queda todo, reiniciar estado
	t.Pot = 0
	t.TurnIndex = 0
	return t, nil
}
