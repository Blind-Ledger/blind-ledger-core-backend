package game

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/poker"
	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/tournament"
)

// Coordinator manages the integration between poker engine, tournaments, and WebSocket communication
type Coordinator struct {
	pokerEngine       poker.Engine
	tournamentManager tournament.Manager

	// WebSocket integration (we'll keep using the existing ws package)
	wsEventChannel chan WSEvent
}

// WSEvent represents events to be sent via WebSocket
type WSEvent struct {
	Channel string
	Data    []byte
}

func NewCoordinator() *Coordinator {
	// Initialize poker engine
	pokerEngine := poker.NewEngine()

	// Initialize tournament manager
	tournamentManager := tournament.NewManager(pokerEngine)

	coord := &Coordinator{
		pokerEngine:       pokerEngine,
		tournamentManager: tournamentManager,
		wsEventChannel:    make(chan WSEvent, 100),
	}

	// Set up event handlers
	coord.setupEventHandlers()

	return coord
}

func (c *Coordinator) setupEventHandlers() {
	// Listen for tournament events and broadcast them via WebSocket
	c.tournamentManager.AddEventHandler(func(event tournament.TournamentEvent) {
		c.broadcastTournamentEvent(event)
	})
}

// Tournament operations
func (c *Coordinator) CreateTournament(tableID string, organizer string, config tournament.TournamentConfig) (*tournament.Tournament, error) {
	log.Printf("🎮 Creating tournament for table %s by %s", tableID, organizer)

	t, err := c.tournamentManager.CreateTournament(tableID, organizer, config)
	if err != nil {
		log.Printf("❌ Failed to create tournament: %v", err)
		return nil, err
	}

	log.Printf("✅ Tournament created: %s", t.ID)
	return t, nil
}

func (c *Coordinator) JoinTournament(tournamentID string, playerID, playerName, walletAddr string) (*tournament.Tournament, error) {
	log.Printf("👤 Player %s joining tournament %s", playerName, tournamentID)

	participant := tournament.Participant{
		PlayerID:   playerID,
		PlayerName: playerName,
		WalletAddr: walletAddr,
	}

	t, err := c.tournamentManager.JoinTournament(tournamentID, participant)
	if err != nil {
		log.Printf("❌ Failed to join tournament: %v", err)
		return nil, err
	}

	log.Printf("✅ Player %s joined tournament %s (%d/%d players)",
		playerName, tournamentID, len(t.Participants), t.MaxPlayers)

	return t, nil
}

func (c *Coordinator) StartTournament(tournamentID string) (*tournament.Tournament, error) {
	log.Printf("🚀 Starting tournament %s", tournamentID)

	t, err := c.tournamentManager.StartTournament(tournamentID)
	if err != nil {
		log.Printf("❌ Failed to start tournament: %v", err)
		return nil, err
	}

	log.Printf("✅ Tournament %s started with %d players", tournamentID, len(t.Participants))
	return t, nil
}

// Game operations (integrating with existing WebSocket protocol)
func (c *Coordinator) ProcessPlayerAction(tournamentID string, playerID string, actionType string, amount int) (*tournament.Tournament, error) {
	log.Printf("🎯 Processing action: %s by %s in tournament %s", actionType, playerID, tournamentID)

	// Convert action type to poker action
	pokerActionType, err := c.convertActionType(actionType)
	if err != nil {
		return nil, err
	}

	action := poker.PlayerAction{
		PlayerID: playerID,
		Type:     pokerActionType,
		Amount:   amount,
	}

	t, err := c.tournamentManager.ProcessGameAction(tournamentID, action)
	if err != nil {
		log.Printf("❌ Failed to process action: %v", err)
		return nil, err
	}

	log.Printf("✅ Action processed: %s", actionType)
	return t, nil
}

func (c *Coordinator) AutoCompleteTournament(tournamentID string) (*tournament.Tournament, error) {
	log.Printf("⚡ Auto-completing tournament %s", tournamentID)

	t, err := c.tournamentManager.CompleteGame(tournamentID)
	if err != nil {
		log.Printf("❌ Failed to auto-complete tournament: %v", err)
		return nil, err
	}

	log.Printf("✅ Tournament %s completed", tournamentID)
	return t, nil
}

// State queries
func (c *Coordinator) GetTournament(tournamentID string) (*tournament.Tournament, error) {
	return c.tournamentManager.GetTournament(tournamentID)
}

func (c *Coordinator) GetActiveTournaments() ([]tournament.Tournament, error) {
	return c.tournamentManager.GetActiveTournaments()
}

// WebSocket integration
func (c *Coordinator) GetWSEventChannel() <-chan WSEvent {
	return c.wsEventChannel
}

func (c *Coordinator) broadcastTournamentEvent(event tournament.TournamentEvent) {
	// Convert tournament event to WebSocket message
	wsMessage := c.convertTournamentEventToWSMessage(event)

	// Determine the channel (table ID from tournament)
	var channel string
	if t, err := c.tournamentManager.GetTournament(event.TournamentID); err == nil {
		channel = t.TableID
	} else {
		channel = "general" // fallback
	}

	// Send to WebSocket channel
	select {
	case c.wsEventChannel <- WSEvent{Channel: channel, Data: wsMessage}:
		log.Printf("📡 Broadcasted event %s to channel %s", event.Type, channel)
	default:
		log.Printf("⚠️ WebSocket event channel full, dropping event")
	}
}

func (c *Coordinator) convertTournamentEventToWSMessage(event tournament.TournamentEvent) []byte {
	// Convert to the existing WebSocket protocol format
	message := map[string]interface{}{
		"type":    "tournament_update", // Keep consistent with existing protocol
		"version": 1,
		"payload": map[string]interface{}{
			"event_type":    event.Type,
			"tournament_id": event.TournamentID,
			"timestamp":     event.Timestamp,
			"data":          event.Data,
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Failed to marshal WebSocket message: %v", err)
		return []byte("{\"error\":\"failed to marshal message\"}")
	}

	return data
}

func (c *Coordinator) convertActionType(actionType string) (poker.ActionType, error) {
	switch actionType {
	case "fold":
		return poker.ActionFold, nil
	case "check":
		return poker.ActionCheck, nil
	case "call":
		return poker.ActionCall, nil
	case "bet":
		return poker.ActionBet, nil
	case "raise":
		return poker.ActionRaise, nil
	case "all_in":
		return poker.ActionAllIn, nil
	default:
		return "", fmt.Errorf("unsupported action type: %s", actionType)
	}
}

// Blockchain integration helpers (for future implementation)
func (c *Coordinator) GetTournamentForBlockchain(tournamentID string) (*tournament.BlockchainTournament, error) {
	t, err := c.tournamentManager.GetTournament(tournamentID)
	if err != nil {
		return nil, err
	}

	// Convert to blockchain format
	var participants []string
	var winners []string
	var prizeAmounts []int

	for _, p := range t.Participants {
		participants = append(participants, p.WalletAddr)
	}

	for _, w := range t.Winners {
		winners = append(winners, w.WalletAddr)
		prizeAmounts = append(prizeAmounts, w.PrizeAmount)
	}

	return &tournament.BlockchainTournament{
		ID:            t.ID,
		Organizer:     t.Organizer,
		Participants:  participants,
		EntryFee:      t.EntryFee,
		EscrowBalance: t.EscrowBalance,
		IsCompleted:   t.IsComplete(),
		Winners:       winners,
		PrizeAmounts:  prizeAmounts,
	}, nil
}

// Integration with existing manager.go for backward compatibility
func (c *Coordinator) LegacyJoin(tableID, playerName string) *LegacyTableState {
	// For demo purposes, automatically create a tournament if none exists
	// In production, this would be handled differently

	tournaments, _ := c.GetActiveTournaments()
	var targetTournament *tournament.Tournament

	// Find existing tournament for this table
	for _, t := range tournaments {
		if t.TableID == tableID && t.IsJoinable() {
			targetTournament = &t
			break
		}
	}

	// Create tournament if none exists
	if targetTournament == nil {
		config := tournament.DefaultSitAndGoConfig()
		t, err := c.CreateTournament(tableID, "system", config)
		if err != nil {
			log.Printf("❌ Failed to create auto tournament: %v", err)
			return &LegacyTableState{Host: playerName, Players: []LegacyPlayer{{Name: playerName}}}
		}
		targetTournament = t
	}

	// Join tournament
	playerID := fmt.Sprintf("player_%s_%d", playerName, len(targetTournament.Participants))
	_, err := c.JoinTournament(targetTournament.ID, playerID, playerName, "mock_wallet")
	if err != nil {
		log.Printf("❌ Failed to join auto tournament: %v", err)
	}

	// Convert to legacy format for backward compatibility
	return c.convertToLegacyState(targetTournament)
}

func (c *Coordinator) LegacyBet(tableID, playerName string, amount int) (*LegacyTableState, error) {
	// Find active tournament for this table
	tournaments, _ := c.GetActiveTournaments()
	var targetTournament *tournament.Tournament

	for _, t := range tournaments {
		if t.TableID == tableID && t.IsActive() {
			targetTournament = &t
			break
		}
	}

	if targetTournament == nil {
		return nil, fmt.Errorf("no active tournament for table %s", tableID)
	}

	// Find player ID by name
	var playerID string
	for _, p := range targetTournament.Participants {
		if p.PlayerName == playerName {
			playerID = p.PlayerID
			break
		}
	}

	if playerID == "" {
		return nil, fmt.Errorf("player %s not found in tournament", playerName)
	}

	// Process action
	updatedTournament, err := c.ProcessPlayerAction(targetTournament.ID, playerID, "call", amount)
	if err != nil {
		return nil, err
	}

	return c.convertToLegacyState(updatedTournament), nil
}

// Legacy types for backward compatibility with existing WebSocket protocol
type LegacyTableState struct {
	Host      string         `json:"host"`
	Players   []LegacyPlayer `json:"players"`
	Pot       int            `json:"pot"`
	TurnIndex int            `json:"turn_index"`
}

type LegacyPlayer struct {
	Name string `json:"name"`
}

func (c *Coordinator) convertToLegacyState(t *tournament.Tournament) *LegacyTableState {
	var players []LegacyPlayer
	for _, p := range t.Participants {
		players = append(players, LegacyPlayer{Name: p.PlayerName})
	}

	pot := 0
	turnIndex := 0

	if t.CurrentGame != nil {
		pot = t.CurrentGame.Pot
		turnIndex = t.CurrentGame.ActionPos
	}

	host := "system"
	if len(t.Participants) > 0 {
		host = t.Participants[0].PlayerName
	}

	return &LegacyTableState{
		Host:      host,
		Players:   players,
		Pot:       pot,
		TurnIndex: turnIndex,
	}
}
