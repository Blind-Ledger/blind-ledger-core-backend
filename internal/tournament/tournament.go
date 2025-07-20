package tournament

import (
	"fmt"
	"sync"
	"time"

	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/poker"
)

// TournamentStatus representa el estado del torneo
type TournamentStatus string

const (
	StatusRegistering TournamentStatus = "registering"
	StatusStarting    TournamentStatus = "starting"
	StatusActive      TournamentStatus = "active"
	StatusFinalTable  TournamentStatus = "final_table"
	StatusFinished    TournamentStatus = "finished"
	StatusCancelled   TournamentStatus = "cancelled"
)

// BlindLevel define un nivel de blinds
type BlindLevel struct {
	Level      int           `json:"level"`
	SmallBlind int           `json:"small_blind"`
	BigBlind   int           `json:"big_blind"`
	Ante       int           `json:"ante"`
	Duration   time.Duration `json:"duration"` // Duración en minutos
}

// TournamentConfig configuración del torneo
type TournamentConfig struct {
	Name              string       `json:"name"`
	BuyIn             int          `json:"buy_in"`            // Costo de entrada
	StartingStack     int          `json:"starting_stack"`    // Stack inicial
	MaxPlayers        int          `json:"max_players"`       // Máximo de jugadores
	MinPlayers        int          `json:"min_players"`       // Mínimo para iniciar
	BlindLevels       []BlindLevel `json:"blind_levels"`      // Estructura de blinds
	MaxTablesSize     int          `json:"max_tables_size"`   // Jugadores por mesa
	RegistrationDelay time.Duration `json:"registration_delay"` // Tiempo antes de iniciar
}

// RegisteredPlayer jugador registrado
type RegisteredPlayer struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	BuyInPaid    int       `json:"buy_in_paid"`
	RegisteredAt time.Time `json:"registered_at"`
	IsEliminated bool      `json:"is_eliminated"`
	Position     int       `json:"position"` // Posición final (0 = sin eliminar)
}

// TournamentTable mesa del torneo
type TournamentTable struct {
	ID          string              `json:"id"`
	PokerTable  *poker.PokerTable   `json:"poker_table"`
	PlayerIDs   []string            `json:"player_ids"`
	IsActive    bool                `json:"is_active"`
	IsFinalTable bool               `json:"is_final_table"`
}

// Tournament estructura principal del torneo
type Tournament struct {
	mu sync.RWMutex

	ID           string              `json:"id"`
	Config       TournamentConfig    `json:"config"`
	Status       TournamentStatus    `json:"status"`
	Players      map[string]*RegisteredPlayer `json:"players"`
	Tables       map[string]*TournamentTable  `json:"tables"`
	PrizePool    int                 `json:"prize_pool"`
	CurrentLevel int                 `json:"current_level"`
	LevelStartTime time.Time         `json:"level_start_time"`
	StartTime    time.Time           `json:"start_time"`
	EndTime      *time.Time          `json:"end_time,omitempty"`
	Winners      []string            `json:"winners,omitempty"`
	
	// Control interno
	pokerEngine    *poker.PokerEngine
	levelTimer     *time.Timer
	nextPlayerPos  int // Para tracking de posiciones finales
}

// NewTournament crea un nuevo torneo
func NewTournament(id string, config TournamentConfig, pokerEngine *poker.PokerEngine) *Tournament {
	return &Tournament{
		ID:           id,
		Config:       config,
		Status:       StatusRegistering,
		Players:      make(map[string]*RegisteredPlayer),
		Tables:       make(map[string]*TournamentTable),
		PrizePool:    0,
		CurrentLevel: 0,
		pokerEngine:  pokerEngine,
		nextPlayerPos: 1,
	}
}

// RegisterPlayer registra un jugador en el torneo
func (t *Tournament) RegisterPlayer(playerID, playerName string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Verificar estado del torneo
	if t.Status != StatusRegistering {
		return fmt.Errorf("tournament registration is closed")
	}

	// Verificar si ya está registrado
	if _, exists := t.Players[playerID]; exists {
		return fmt.Errorf("player already registered")
	}

	// Verificar límite de jugadores
	if len(t.Players) >= t.Config.MaxPlayers {
		return fmt.Errorf("tournament is full")
	}

	// Registrar jugador
	player := &RegisteredPlayer{
		ID:           playerID,
		Name:         playerName,
		BuyInPaid:    t.Config.BuyIn,
		RegisteredAt: time.Now(),
		IsEliminated: false,
		Position:     0,
	}

	t.Players[playerID] = player
	t.PrizePool += t.Config.BuyIn

	// Verificar si se puede iniciar el torneo
	if len(t.Players) >= t.Config.MinPlayers {
		// Programar inicio automático después del delay
		if t.levelTimer == nil {
			t.levelTimer = time.AfterFunc(t.Config.RegistrationDelay, func() {
				t.StartTournament()
			})
		}
	}

	return nil
}

// UnregisterPlayer desregistra un jugador (solo durante registro)
func (t *Tournament) UnregisterPlayer(playerID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Status != StatusRegistering {
		return fmt.Errorf("cannot unregister after tournament starts")
	}

	player, exists := t.Players[playerID]
	if !exists {
		return fmt.Errorf("player not registered")
	}

	delete(t.Players, playerID)
	t.PrizePool -= player.BuyInPaid

	// Cancelar timer si no hay suficientes jugadores
	if len(t.Players) < t.Config.MinPlayers && t.levelTimer != nil {
		t.levelTimer.Stop()
		t.levelTimer = nil
	}

	return nil
}

// StartTournament inicia el torneo
func (t *Tournament) StartTournament() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Status != StatusRegistering {
		return fmt.Errorf("tournament already started or finished")
	}

	if len(t.Players) < t.Config.MinPlayers {
		return fmt.Errorf("not enough players to start tournament")
	}

	t.Status = StatusStarting
	t.StartTime = time.Now()
	t.CurrentLevel = 0
	t.LevelStartTime = time.Now()

	// Crear mesas iniciales
	err := t.createInitialTables()
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	t.Status = StatusActive

	// Iniciar timer de blinds
	t.startBlindTimer()

	return nil
}

// createInitialTables crea las mesas iniciales balanceadas
func (t *Tournament) createInitialTables() error {
	playerIDs := make([]string, 0, len(t.Players))
	for id := range t.Players {
		playerIDs = append(playerIDs, id)
	}

	// Calcular número de mesas necesarias
	playersPerTable := t.Config.MaxTablesSize
	numTables := (len(playerIDs) + playersPerTable - 1) / playersPerTable

	// Distribuir jugadores en mesas
	for i := 0; i < numTables; i++ {
		tableID := fmt.Sprintf("%s_table_%d", t.ID, i+1)
		
		// Calcular jugadores para esta mesa
		startIdx := i * playersPerTable
		endIdx := startIdx + playersPerTable
		if endIdx > len(playerIDs) {
			endIdx = len(playerIDs)
		}

		tablePlayerIDs := playerIDs[startIdx:endIdx]

		// Crear mesa de poker
		pokerTable := t.pokerEngine.CreateTable(tableID)
		
		// Configurar blinds iniciales
		if len(t.Config.BlindLevels) > 0 {
			level := t.Config.BlindLevels[0]
			pokerTable.SmallBlind = level.SmallBlind
			pokerTable.BigBlind = level.BigBlind
		}

		// Agregar jugadores a la mesa de poker
		for _, playerID := range tablePlayerIDs {
			player := t.Players[playerID]
			_, err := t.pokerEngine.AddPlayer(tableID, playerID, player.Name)
			if err != nil {
				return fmt.Errorf("failed to add player %s to table %s: %w", playerID, tableID, err)
			}
		}

		// Crear tabla del torneo
		tournamentTable := &TournamentTable{
			ID:           tableID,
			PokerTable:   pokerTable,
			PlayerIDs:    tablePlayerIDs,
			IsActive:     true,
			IsFinalTable: false,
		}

		t.Tables[tableID] = tournamentTable
	}

	return nil
}

// startBlindTimer inicia el timer para el próximo nivel de blinds
func (t *Tournament) startBlindTimer() {
	if t.CurrentLevel >= len(t.Config.BlindLevels) {
		return
	}

	level := t.Config.BlindLevels[t.CurrentLevel]
	
	t.levelTimer = time.AfterFunc(level.Duration, func() {
		t.advanceBlindLevel()
	})
}

// advanceBlindLevel avanza al siguiente nivel de blinds
func (t *Tournament) advanceBlindLevel() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Status != StatusActive {
		return
	}

	t.CurrentLevel++
	t.LevelStartTime = time.Now()

	// Aplicar nuevos blinds a todas las mesas
	if t.CurrentLevel < len(t.Config.BlindLevels) {
		level := t.Config.BlindLevels[t.CurrentLevel]
		
		for _, table := range t.Tables {
			if table.IsActive && table.PokerTable != nil {
				table.PokerTable.SmallBlind = level.SmallBlind
				table.PokerTable.BigBlind = level.BigBlind
			}
		}

		// Programar siguiente nivel
		t.startBlindTimer()
	}
}

// EliminatePlayer elimina un jugador del torneo
func (t *Tournament) EliminatePlayer(playerID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	player, exists := t.Players[playerID]
	if !exists {
		return fmt.Errorf("player not found")
	}

	if player.IsEliminated {
		return fmt.Errorf("player already eliminated")
	}

	// Marcar como eliminado
	player.IsEliminated = true
	player.Position = len(t.Players) - t.nextPlayerPos + 1
	t.nextPlayerPos++

	// Verificar si el torneo ha terminado
	activePlayers := t.getActivePlayers()
	if len(activePlayers) <= 1 {
		t.finishTournament(activePlayers)
		return nil
	}

	// Verificar si es mesa final
	if len(activePlayers) <= t.Config.MaxTablesSize && t.Status != StatusFinalTable {
		t.createFinalTable(activePlayers)
	}

	return nil
}

// getActivePlayers retorna jugadores activos
func (t *Tournament) getActivePlayers() []string {
	var active []string
	for id, player := range t.Players {
		if !player.IsEliminated {
			active = append(active, id)
		}
	}
	return active
}

// createFinalTable crea la mesa final
func (t *Tournament) createFinalTable(activePlayers []string) {
	t.Status = StatusFinalTable

	// Desactivar todas las mesas actuales
	for _, table := range t.Tables {
		table.IsActive = false
	}

	// Crear mesa final
	finalTableID := fmt.Sprintf("%s_final", t.ID)
	pokerTable := t.pokerEngine.CreateTable(finalTableID)
	
	// Configurar blinds actuales
	if t.CurrentLevel < len(t.Config.BlindLevels) {
		level := t.Config.BlindLevels[t.CurrentLevel]
		pokerTable.SmallBlind = level.SmallBlind
		pokerTable.BigBlind = level.BigBlind
	}

	// Mover jugadores a mesa final
	for _, playerID := range activePlayers {
		player := t.Players[playerID]
		t.pokerEngine.AddPlayer(finalTableID, playerID, player.Name)
	}

	finalTable := &TournamentTable{
		ID:           finalTableID,
		PokerTable:   pokerTable,
		PlayerIDs:    activePlayers,
		IsActive:     true,
		IsFinalTable: true,
	}

	t.Tables[finalTableID] = finalTable
}

// finishTournament termina el torneo
func (t *Tournament) finishTournament(winners []string) {
	t.Status = StatusFinished
	now := time.Now()
	t.EndTime = &now
	t.Winners = winners

	// Detener timer de blinds
	if t.levelTimer != nil {
		t.levelTimer.Stop()
	}

	// Distribuir premios (lógica básica - 100% al ganador)
	if len(winners) > 0 {
		// En una implementación real, aquí se distribuirían los premios
		// según una estructura de payout
	}
}

// GetStatus retorna el estado actual del torneo
func (t *Tournament) GetStatus() TournamentStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status
}

// GetPlayerCount retorna el número de jugadores registrados
func (t *Tournament) GetPlayerCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.Players)
}

// GetActiveTables retorna las mesas activas
func (t *Tournament) GetActiveTables() map[string]*TournamentTable {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	active := make(map[string]*TournamentTable)
	for id, table := range t.Tables {
		if table.IsActive {
			active[id] = table
		}
	}
	return active
}

// GetCurrentBlindLevel retorna el nivel actual de blinds
func (t *Tournament) GetCurrentBlindLevel() BlindLevel {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if t.CurrentLevel >= len(t.Config.BlindLevels) {
		// Retornar último nivel si nos pasamos
		return t.Config.BlindLevels[len(t.Config.BlindLevels)-1]
	}
	return t.Config.BlindLevels[t.CurrentLevel]
}