package tournament

import (
	"time"

	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/poker"
)

// Tournament represents a complete tournament structure
type Tournament struct {
	ID        string `json:"id"`
	TableID   string `json:"table_id"`
	Name      string `json:"name"`
	Organizer string `json:"organizer"`

	// Tournament Configuration
	MaxPlayers     int          `json:"max_players"`
	EntryFee       int          `json:"entry_fee"`       // in WBTC wei/satoshi
	PrizeStructure []PrizeLevel `json:"prize_structure"` // % distribution

	// Tournament State
	Status       TournamentStatus `json:"status"`
	Participants []Participant    `json:"participants"`
	CurrentGame  *poker.GameState `json:"current_game,omitempty"`

	// Blockchain Integration
	SmartContractAddr string `json:"smart_contract_addr,omitempty"`
	EscrowBalance     int    `json:"escrow_balance"`

	// Timing
	StartTime       time.Time  `json:"start_time"`
	RegistrationEnd time.Time  `json:"registration_end"`
	ActualStart     *time.Time `json:"actual_start,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`

	// Results
	Results []TournamentResult `json:"results,omitempty"`
	Winners []Winner           `json:"winners,omitempty"`
}

type TournamentStatus string

const (
	StatusRegistering TournamentStatus = "registering" // Players can join
	StatusWaiting     TournamentStatus = "waiting"     // Waiting for minimum players
	StatusInProgress  TournamentStatus = "in_progress" // Game is running
	StatusCompleted   TournamentStatus = "completed"   // Tournament finished
	StatusCancelled   TournamentStatus = "cancelled"   // Tournament cancelled
)

type Participant struct {
	PlayerID   string    `json:"player_id"`
	PlayerName string    `json:"player_name"`
	WalletAddr string    `json:"wallet_addr"`
	JoinedAt   time.Time `json:"joined_at"`
	IsActive   bool      `json:"is_active"`

	// Tournament Progress
	CurrentStack int        `json:"current_stack,omitempty"`
	Position     *int       `json:"position,omitempty"` // Final position when eliminated
	EliminatedAt *time.Time `json:"eliminated_at,omitempty"`
}

type PrizeLevel struct {
	Position   int `json:"position"`   // 1st, 2nd, 3rd, etc.
	Percentage int `json:"percentage"` // % of total prize pool
}

type TournamentResult struct {
	Position    int               `json:"position"`
	PlayerID    string            `json:"player_id"`
	PlayerName  string            `json:"player_name"`
	PrizeAmount int               `json:"prize_amount"`
	FinalHand   *poker.HandResult `json:"final_hand,omitempty"`
}

type Winner struct {
	PlayerID    string `json:"player_id"`
	PlayerName  string `json:"player_name"`
	WalletAddr  string `json:"wallet_addr"`
	Position    int    `json:"position"`
	PrizeAmount int    `json:"prize_amount"`
}

// TournamentConfig holds configuration for new tournaments
type TournamentConfig struct {
	Name                 string        `json:"name"`
	MaxPlayers           int           `json:"max_players"`
	EntryFee             int           `json:"entry_fee"`
	PrizeStructure       []PrizeLevel  `json:"prize_structure"`
	RegistrationDuration time.Duration `json:"registration_duration"`

	// Poker-specific config
	StartingStack int           `json:"starting_stack"`
	SmallBlind    int           `json:"small_blind"`
	BigBlind      int           `json:"big_blind"`
	BlindIncrease time.Duration `json:"blind_increase,omitempty"` // Future: blind level increases
}

// TournamentManager interface for tournament operations
type Manager interface {
	// Tournament lifecycle
	CreateTournament(tableID string, organizer string, config TournamentConfig) (*Tournament, error)
	JoinTournament(tournamentID string, participant Participant) (*Tournament, error)
	StartTournament(tournamentID string) (*Tournament, error)
	CancelTournament(tournamentID string) (*Tournament, error)

	// Game integration
	ProcessGameAction(tournamentID string, action poker.PlayerAction) (*Tournament, error)
	CompleteGame(tournamentID string) (*Tournament, error)

	// State management
	GetTournament(tournamentID string) (*Tournament, error)
	GetActiveTournaments() ([]Tournament, error)

	// Results and payouts
	FinalizeTournament(tournamentID string) (*Tournament, error)
}

// Event types for WebSocket broadcasting
type TournamentEvent struct {
	Type         EventType   `json:"type"`
	TournamentID string      `json:"tournament_id"`
	Timestamp    time.Time   `json:"timestamp"`
	Data         interface{} `json:"data"`
}

type EventType string

const (
	EventPlayerJoined        EventType = "player_joined"
	EventTournamentStarted   EventType = "tournament_started"
	EventGameStarted         EventType = "game_started"
	EventGameAction          EventType = "game_action"
	EventGameCompleted       EventType = "game_completed"
	EventPlayerEliminated    EventType = "player_eliminated"
	EventTournamentCompleted EventType = "tournament_completed"
	EventError               EventType = "error"
)

// For blockchain integration
type BlockchainTournament struct {
	ID            string   `json:"id"`
	Organizer     string   `json:"organizer"`
	Participants  []string `json:"participants"` // wallet addresses
	EntryFee      int      `json:"entry_fee"`
	EscrowBalance int      `json:"escrow_balance"`
	IsCompleted   bool     `json:"is_completed"`
	Winners       []string `json:"winners,omitempty"` // wallet addresses
	PrizeAmounts  []int    `json:"prize_amounts,omitempty"`
}

// Default tournament configurations for MVP
func DefaultSitAndGoConfig() TournamentConfig {
	return TournamentConfig{
		Name:          "4-Player Sit & Go",
		MaxPlayers:    4,
		EntryFee:      100, // 100 WBTC wei (adjust as needed)
		StartingStack: 1000,
		SmallBlind:    10,
		BigBlind:      20,
		PrizeStructure: []PrizeLevel{
			{Position: 1, Percentage: 65}, // Winner gets 65%
			{Position: 2, Percentage: 35}, // Runner-up gets 35%
		},
		RegistrationDuration: 10 * time.Minute,
	}
}

// Validation helpers
func (t *Tournament) IsJoinable() bool {
	return t.Status == StatusRegistering &&
		len(t.Participants) < t.MaxPlayers &&
		time.Now().Before(t.RegistrationEnd)
}

func (t *Tournament) CanStart() bool {
	return t.Status == StatusWaiting &&
		len(t.Participants) >= 2 && // Minimum 2 players
		len(t.Participants) <= t.MaxPlayers
}

func (t *Tournament) IsActive() bool {
	return t.Status == StatusInProgress
}

func (t *Tournament) IsComplete() bool {
	return t.Status == StatusCompleted
}

func (t *Tournament) GetTotalPrizePool() int {
	return len(t.Participants) * t.EntryFee
}

func (t *Tournament) GetActivePlayers() []Participant {
	var active []Participant
	for _, p := range t.Participants {
		if p.IsActive {
			active = append(active, p)
		}
	}
	return active
}
