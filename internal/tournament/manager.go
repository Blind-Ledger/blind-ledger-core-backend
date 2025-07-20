package tournament

import (
	"fmt"
	"sync"
	"time"

	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/poker"
)

// Manager maneja múltiples torneos
type Manager struct {
	mu          sync.RWMutex
	tournaments map[string]*Tournament
	pokerEngine *poker.PokerEngine
}

// NewManager crea un nuevo manager de torneos
func NewManager(pokerEngine *poker.PokerEngine) *Manager {
	return &Manager{
		tournaments: make(map[string]*Tournament),
		pokerEngine: pokerEngine,
	}
}

// CreateTournament crea un nuevo torneo
func (m *Manager) CreateTournament(id string, config TournamentConfig) (*Tournament, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verificar que no existe
	if _, exists := m.tournaments[id]; exists {
		return nil, fmt.Errorf("tournament %s already exists", id)
	}

	// Validar configuración
	if err := m.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid tournament config: %w", err)
	}

	tournament := NewTournament(id, config, m.pokerEngine)
	m.tournaments[id] = tournament

	return tournament, nil
}

// validateConfig valida la configuración del torneo
func (m *Manager) validateConfig(config TournamentConfig) error {
	if config.Name == "" {
		return fmt.Errorf("tournament name is required")
	}

	if config.BuyIn <= 0 {
		return fmt.Errorf("buy-in must be positive")
	}

	if config.StartingStack <= 0 {
		return fmt.Errorf("starting stack must be positive")
	}

	if config.MaxPlayers < 2 {
		return fmt.Errorf("max players must be at least 2")
	}

	if config.MinPlayers < 2 {
		return fmt.Errorf("min players must be at least 2")
	}

	if config.MinPlayers > config.MaxPlayers {
		return fmt.Errorf("min players cannot exceed max players")
	}

	if config.MaxTablesSize < 2 || config.MaxTablesSize > 10 {
		return fmt.Errorf("table size must be between 2 and 10")
	}

	if len(config.BlindLevels) == 0 {
		return fmt.Errorf("at least one blind level is required")
	}

	// Validar niveles de blinds
	for i, level := range config.BlindLevels {
		if level.SmallBlind <= 0 || level.BigBlind <= 0 {
			return fmt.Errorf("blind level %d: blinds must be positive", i)
		}

		if level.BigBlind <= level.SmallBlind {
			return fmt.Errorf("blind level %d: big blind must be greater than small blind", i)
		}

		if level.Duration <= 0 {
			return fmt.Errorf("blind level %d: duration must be positive", i)
		}
	}

	return nil
}

// GetTournament obtiene un torneo por ID
func (m *Manager) GetTournament(id string) (*Tournament, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tournament, exists := m.tournaments[id]
	if !exists {
		return nil, fmt.Errorf("tournament %s not found", id)
	}

	return tournament, nil
}

// ListTournaments lista todos los torneos
func (m *Manager) ListTournaments() map[string]*Tournament {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Hacer copia para evitar modificaciones concurrentes
	result := make(map[string]*Tournament)
	for id, tournament := range m.tournaments {
		result[id] = tournament
	}

	return result
}

// ListActiveTournaments lista torneos activos
func (m *Manager) ListActiveTournaments() map[string]*Tournament {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Tournament)
	for id, tournament := range m.tournaments {
		status := tournament.GetStatus()
		if status == StatusRegistering || status == StatusStarting || 
		   status == StatusActive || status == StatusFinalTable {
			result[id] = tournament
		}
	}

	return result
}

// DeleteTournament elimina un torneo (solo si está terminado o cancelado)
func (m *Manager) DeleteTournament(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tournament, exists := m.tournaments[id]
	if !exists {
		return fmt.Errorf("tournament %s not found", id)
	}

	status := tournament.GetStatus()
	if status != StatusFinished && status != StatusCancelled {
		return fmt.Errorf("cannot delete active tournament")
	}

	delete(m.tournaments, id)
	return nil
}

// CancelTournament cancela un torneo
func (m *Manager) CancelTournament(id string) error {
	tournament, err := m.GetTournament(id)
	if err != nil {
		return err
	}

	tournament.mu.Lock()
	defer tournament.mu.Unlock()

	if tournament.Status == StatusFinished || tournament.Status == StatusCancelled {
		return fmt.Errorf("tournament already finished or cancelled")
	}

	// Detener timers
	if tournament.levelTimer != nil {
		tournament.levelTimer.Stop()
	}

	tournament.Status = StatusCancelled
	now := time.Now()
	tournament.EndTime = &now

	return nil
}

// CreateStandardTournament crea un torneo con configuración estándar
func (m *Manager) CreateStandardTournament(id, name string, buyIn int) (*Tournament, error) {
	config := TournamentConfig{
		Name:              name,
		BuyIn:             buyIn,
		StartingStack:     1500,
		MaxPlayers:        18,
		MinPlayers:        4,
		MaxTablesSize:     6,
		RegistrationDelay: 2 * time.Minute,
		BlindLevels: []BlindLevel{
			{Level: 1, SmallBlind: 10, BigBlind: 20, Ante: 0, Duration: 10 * time.Minute},
			{Level: 2, SmallBlind: 15, BigBlind: 30, Ante: 0, Duration: 10 * time.Minute},
			{Level: 3, SmallBlind: 25, BigBlind: 50, Ante: 0, Duration: 10 * time.Minute},
			{Level: 4, SmallBlind: 50, BigBlind: 100, Ante: 0, Duration: 10 * time.Minute},
			{Level: 5, SmallBlind: 75, BigBlind: 150, Ante: 0, Duration: 10 * time.Minute},
			{Level: 6, SmallBlind: 100, BigBlind: 200, Ante: 25, Duration: 10 * time.Minute},
			{Level: 7, SmallBlind: 150, BigBlind: 300, Ante: 25, Duration: 10 * time.Minute},
			{Level: 8, SmallBlind: 200, BigBlind: 400, Ante: 50, Duration: 10 * time.Minute},
			{Level: 9, SmallBlind: 300, BigBlind: 600, Ante: 75, Duration: 10 * time.Minute},
			{Level: 10, SmallBlind: 400, BigBlind: 800, Ante: 100, Duration: 10 * time.Minute},
		},
	}

	return m.CreateTournament(id, config)
}

// CreateTurboTournament crea un torneo turbo (blinds más rápidos)
func (m *Manager) CreateTurboTournament(id, name string, buyIn int) (*Tournament, error) {
	config := TournamentConfig{
		Name:              name + " (Turbo)",
		BuyIn:             buyIn,
		StartingStack:     1500,
		MaxPlayers:        18,
		MinPlayers:        4,
		MaxTablesSize:     6,
		RegistrationDelay: 1 * time.Minute,
		BlindLevels: []BlindLevel{
			{Level: 1, SmallBlind: 10, BigBlind: 20, Ante: 0, Duration: 5 * time.Minute},
			{Level: 2, SmallBlind: 15, BigBlind: 30, Ante: 0, Duration: 5 * time.Minute},
			{Level: 3, SmallBlind: 25, BigBlind: 50, Ante: 0, Duration: 5 * time.Minute},
			{Level: 4, SmallBlind: 50, BigBlind: 100, Ante: 0, Duration: 5 * time.Minute},
			{Level: 5, SmallBlind: 75, BigBlind: 150, Ante: 0, Duration: 5 * time.Minute},
			{Level: 6, SmallBlind: 100, BigBlind: 200, Ante: 25, Duration: 5 * time.Minute},
			{Level: 7, SmallBlind: 150, BigBlind: 300, Ante: 25, Duration: 5 * time.Minute},
			{Level: 8, SmallBlind: 200, BigBlind: 400, Ante: 50, Duration: 5 * time.Minute},
			{Level: 9, SmallBlind: 300, BigBlind: 600, Ante: 75, Duration: 5 * time.Minute},
			{Level: 10, SmallBlind: 400, BigBlind: 800, Ante: 100, Duration: 5 * time.Minute},
		},
	}

	return m.CreateTournament(id, config)
}