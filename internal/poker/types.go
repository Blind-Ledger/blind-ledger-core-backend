package poker

import "time"

// Card represents a playing card
type Card struct {
	Suit Suit `json:"suit"`
	Rank Rank `json:"rank"`
}

type Suit string

const (
	Hearts   Suit = "hearts"
	Diamonds Suit = "diamonds"
	Clubs    Suit = "clubs"
	Spades   Suit = "spades"
)

type Rank string

const (
	Two   Rank = "2"
	Three Rank = "3"
	Four  Rank = "4"
	Five  Rank = "5"
	Six   Rank = "6"
	Seven Rank = "7"
	Eight Rank = "8"
	Nine  Rank = "9"
	Ten   Rank = "10"
	Jack  Rank = "J"
	Queen Rank = "Q"
	King  Rank = "K"
	Ace   Rank = "A"
)

// Player represents a player in the game
type Player struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Stack    int          `json:"stack"`    // chips available
	Position int          `json:"position"` // seat position (0-3 for 4-max)
	Cards    []Card       `json:"cards"`    // hole cards (private)
	Status   PlayerStatus `json:"status"`

	// Betting state
	CurrentBet    int  `json:"current_bet"`    // amount bet this round
	TotalInvested int  `json:"total_invested"` // total put in pot this hand
	HasActed      bool `json:"has_acted"`      // has acted this betting round
	IsAllIn       bool `json:"is_all_in"`
}

type PlayerStatus string

const (
	StatusActive PlayerStatus = "active"
	StatusFolded PlayerStatus = "folded"
	StatusAllIn  PlayerStatus = "all_in"
	StatusOut    PlayerStatus = "out"
)

// GameState represents the current state of a poker hand
type GameState struct {
	ID      string    `json:"id"`
	Players []*Player `json:"players"`
	TableID string    `json:"table_id"`

	// Board state
	CommunityCards []Card    `json:"community_cards"`
	Pot            int       `json:"pot"`
	SidePots       []SidePot `json:"side_pots"` // for all-in scenarios

	// Game flow
	Phase      GamePhase `json:"phase"`
	DealerPos  int       `json:"dealer_pos"`  // dealer button position
	ActionPos  int       `json:"action_pos"`  // current player to act
	CurrentBet int       `json:"current_bet"` // minimum bet to call

	// Blinds
	SmallBlind int `json:"small_blind"`
	BigBlind   int `json:"big_blind"`

	// Deck state (for recovery)
	DeckSeed      string `json:"deck_seed"`  // for reproducible shuffles
	BurnCards     []Card `json:"burn_cards"` // cards burned
	RemainingDeck []Card `json:"-"`          // don't expose in JSON

	// Timing
	StartTime  time.Time `json:"start_time"`
	LastAction time.Time `json:"last_action"`

	// Results (when hand is complete)
	Winners  []Winner `json:"winners,omitempty"`
	Showdown bool     `json:"showdown,omitempty"`
}

type GamePhase string

const (
	PhasePreflop  GamePhase = "preflop"
	PhaseFlop     GamePhase = "flop"
	PhaseTurn     GamePhase = "turn"
	PhaseRiver    GamePhase = "river"
	PhaseShowdown GamePhase = "showdown"
	PhaseComplete GamePhase = "complete"
)

// SidePot for all-in scenarios with multiple players
type SidePot struct {
	Amount      int      `json:"amount"`
	EligibleIDs []string `json:"eligible_ids"` // player IDs eligible for this pot
}

// Winner represents a hand winner
type Winner struct {
	PlayerID string   `json:"player_id"`
	Amount   int      `json:"amount"`
	Hand     HandType `json:"hand"`
	Cards    []Card   `json:"cards"` // 5-card hand that won
}

// HandType represents poker hand rankings
type HandType string

const (
	HighCard      HandType = "high_card"
	Pair          HandType = "pair"
	TwoPair       HandType = "two_pair"
	ThreeOfAKind  HandType = "three_of_a_kind"
	Straight      HandType = "straight"
	Flush         HandType = "flush"
	FullHouse     HandType = "full_house"
	FourOfAKind   HandType = "four_of_a_kind"
	StraightFlush HandType = "straight_flush"
	RoyalFlush    HandType = "royal_flush"
)

// PlayerAction represents an action a player can take
type PlayerAction struct {
	PlayerID  string     `json:"player_id"`
	Type      ActionType `json:"type"`
	Amount    int        `json:"amount,omitempty"` // for bet/raise
	Timestamp time.Time  `json:"timestamp"`
}

type ActionType string

const (
	ActionFold  ActionType = "fold"
	ActionCheck ActionType = "check"
	ActionCall  ActionType = "call"
	ActionBet   ActionType = "bet"
	ActionRaise ActionType = "raise"
	ActionAllIn ActionType = "all_in"
)

// Hand evaluation result
type HandResult struct {
	Type    HandType `json:"type"`
	Cards   []Card   `json:"cards"`   // the 5 cards that make the hand
	Rank    int      `json:"rank"`    // comparable rank for tie-breaking
	Kickers []Rank   `json:"kickers"` // tie-breaker cards
}

// Engine interface - this is what we'll implement incrementally
type Engine interface {
	// Game lifecycle
	NewGame(tableID string, players []string, blinds Blinds) (*GameState, error)
	ProcessAction(gameID string, action PlayerAction) (*GameState, error)
	GetGameState(gameID string) (*GameState, error)

	// For MVP: auto-complete games
	AutoCompleteGame(gameID string) (*GameState, error)
}

type Blinds struct {
	Small int `json:"small"`
	Big   int `json:"big"`
}
