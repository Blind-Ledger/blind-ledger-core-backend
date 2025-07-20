package tournament

import (
	"fmt"
	"sync"
	"time"

	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/poker"
)

type tournamentManager struct {
	mu          sync.RWMutex
	tournaments map[string]*Tournament
	pokerEngine poker.Engine

	// Event broadcasting
	eventHandlers []func(TournamentEvent)
}

func NewManager(pokerEngine poker.Engine) Manager {
	return &tournamentManager{
		tournaments:   make(map[string]*Tournament),
		pokerEngine:   pokerEngine,
		eventHandlers: make([]func(TournamentEvent), 0),
	}
}

func (tm *tournamentManager) CreateTournament(tableID string, organizer string, config TournamentConfig) (*Tournament, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Validate configuration
	if err := tm.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Generate unique tournament ID
	tournamentID := fmt.Sprintf("tournament_%s_%d", tableID, time.Now().Unix())

	tournament := &Tournament{
		ID:              tournamentID,
		TableID:         tableID,
		Name:            config.Name,
		Organizer:       organizer,
		MaxPlayers:      config.MaxPlayers,
		EntryFee:        config.EntryFee,
		PrizeStructure:  config.PrizeStructure,
		Status:          StatusRegistering,
		Participants:    make([]Participant, 0, config.MaxPlayers),
		StartTime:       time.Now(),
		RegistrationEnd: time.Now().Add(config.RegistrationDuration),
		EscrowBalance:   0,
	}

	tm.tournaments[tournamentID] = tournament

	// Broadcast event
	tm.broadcastEvent(TournamentEvent{
		Type:         EventTournamentStarted,
		TournamentID: tournamentID,
		Timestamp:    time.Now(),
		Data:         tournament,
	})

	return tournament, nil
}

func (tm *tournamentManager) JoinTournament(tournamentID string, participant Participant) (*Tournament, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return nil, fmt.Errorf("tournament not found: %s", tournamentID)
	}

	// Validate join conditions
	if !tournament.IsJoinable() {
		return nil, fmt.Errorf("tournament not joinable: status=%s, players=%d/%d",
			tournament.Status, len(tournament.Participants), tournament.MaxPlayers)
	}

	// Check if player already joined
	for _, p := range tournament.Participants {
		if p.PlayerID == participant.PlayerID {
			return nil, fmt.Errorf("player already in tournament: %s", participant.PlayerID)
		}
	}

	// Add participant
	participant.JoinedAt = time.Now()
	participant.IsActive = true
	tournament.Participants = append(tournament.Participants, participant)
	tournament.EscrowBalance += tournament.EntryFee

	// Check if tournament is ready to start
	if len(tournament.Participants) == tournament.MaxPlayers {
		tournament.Status = StatusWaiting
	}

	// Broadcast event
	tm.broadcastEvent(TournamentEvent{
		Type:         EventPlayerJoined,
		TournamentID: tournamentID,
		Timestamp:    time.Now(),
		Data: map[string]interface{}{
			"participant": participant,
			"tournament":  tournament,
		},
	})

	return tournament, nil
}

func (tm *tournamentManager) StartTournament(tournamentID string) (*Tournament, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return nil, fmt.Errorf("tournament not found: %s", tournamentID)
	}

	if !tournament.CanStart() {
		return nil, fmt.Errorf("tournament cannot start: status=%s, players=%d",
			tournament.Status, len(tournament.Participants))
	}

	// Update tournament status
	tournament.Status = StatusInProgress
	now := time.Now()
	tournament.ActualStart = &now

	// Create initial poker game
	playerNames := make([]string, len(tournament.Participants))
	for i, p := range tournament.Participants {
		playerNames[i] = p.PlayerName
	}

	blinds := poker.Blinds{
		Small: 10, // TODO: Get from tournament config
		Big:   20,
	}

	game, err := tm.pokerEngine.NewGame(tournament.TableID, playerNames, blinds)
	if err != nil {
		return nil, fmt.Errorf("failed to create poker game: %w", err)
	}

	tournament.CurrentGame = game

	// Broadcast event
	tm.broadcastEvent(TournamentEvent{
		Type:         EventGameStarted,
		TournamentID: tournamentID,
		Timestamp:    time.Now(),
		Data:         tournament,
	})

	return tournament, nil
}

func (tm *tournamentManager) ProcessGameAction(tournamentID string, action poker.PlayerAction) (*Tournament, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return nil, fmt.Errorf("tournament not found: %s", tournamentID)
	}

	if !tournament.IsActive() || tournament.CurrentGame == nil {
		return nil, fmt.Errorf("no active game in tournament")
	}

	// Process action through poker engine
	updatedGame, err := tm.pokerEngine.ProcessAction(tournament.CurrentGame.ID, action)
	if err != nil {
		return nil, fmt.Errorf("failed to process action: %w", err)
	}

	tournament.CurrentGame = updatedGame

	// Broadcast action event
	tm.broadcastEvent(TournamentEvent{
		Type:         EventGameAction,
		TournamentID: tournamentID,
		Timestamp:    time.Now(),
		Data: map[string]interface{}{
			"action": action,
			"game":   updatedGame,
		},
	})

	// Check if game is complete
	if updatedGame.Phase == poker.PhaseComplete {
		return tm.handleGameCompletion(tournament)
	}

	return tournament, nil
}

func (tm *tournamentManager) CompleteGame(tournamentID string) (*Tournament, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return nil, fmt.Errorf("tournament not found: %s", tournamentID)
	}

	if !tournament.IsActive() || tournament.CurrentGame == nil {
		return nil, fmt.Errorf("no active game in tournament")
	}

	// Auto-complete the current game (MVP: all-in showdown)
	completedGame, err := tm.pokerEngine.AutoCompleteGame(tournament.CurrentGame.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to complete game: %w", err)
	}

	tournament.CurrentGame = completedGame

	return tm.handleGameCompletion(tournament)
}

func (tm *tournamentManager) handleGameCompletion(tournament *Tournament) (*Tournament, error) {
	if tournament.CurrentGame == nil || tournament.CurrentGame.Phase != poker.PhaseComplete {
		return nil, fmt.Errorf("game not complete")
	}

	// Broadcast game completion
	tm.broadcastEvent(TournamentEvent{
		Type:         EventGameCompleted,
		TournamentID: tournament.ID,
		Timestamp:    time.Now(),
		Data:         tournament.CurrentGame,
	})

	// For MVP: Since we only support 4-max sit & go, one game = tournament complete
	// In full implementation, this would handle elimination and start new games

	// Determine final results from the poker game
	tournament.Results = tm.convertGameResultsToTournamentResults(tournament.CurrentGame)
	tournament.Winners = tm.calculateWinners(tournament)

	// Mark tournament as completed
	tournament.Status = StatusCompleted
	now := time.Now()
	tournament.CompletedAt = &now

	// Broadcast tournament completion
	tm.broadcastEvent(TournamentEvent{
		Type:         EventTournamentCompleted,
		TournamentID: tournament.ID,
		Timestamp:    time.Now(),
		Data:         tournament,
	})

	return tournament, nil
}

func (tm *tournamentManager) convertGameResultsToTournamentResults(game *poker.GameState) []TournamentResult {
	var results []TournamentResult

	// Create results for winners
	for i, winner := range game.Winners {
		results = append(results, TournamentResult{
			Position:    i + 1,
			PlayerID:    winner.PlayerID,
			PlayerName:  tm.getPlayerName(game, winner.PlayerID),
			PrizeAmount: winner.Amount,
		})
	}

	// Add non-winners (folded players) in order they folded
	position := len(game.Winners) + 1
	for _, player := range game.Players {
		if player.Status == poker.StatusFolded {
			results = append(results, TournamentResult{
				Position:    position,
				PlayerID:    player.ID,
				PlayerName:  player.Name,
				PrizeAmount: 0,
			})
			position++
		}
	}

	return results
}

func (tm *tournamentManager) calculateWinners(tournament *Tournament) []Winner {
	var winners []Winner
	totalPrize := tournament.GetTotalPrizePool()

	for _, prizeLevel := range tournament.PrizeStructure {
		// Find player at this position
		for _, result := range tournament.Results {
			if result.Position == prizeLevel.Position {
				prizeAmount := (totalPrize * prizeLevel.Percentage) / 100

				// Find participant details
				var participant Participant
				for _, p := range tournament.Participants {
					if p.PlayerID == result.PlayerID {
						participant = p
						break
					}
				}

				winners = append(winners, Winner{
					PlayerID:    result.PlayerID,
					PlayerName:  result.PlayerName,
					WalletAddr:  participant.WalletAddr,
					Position:    result.Position,
					PrizeAmount: prizeAmount,
				})
				break
			}
		}
	}

	return winners
}

func (tm *tournamentManager) getPlayerName(game *poker.GameState, playerID string) string {
	for _, player := range game.Players {
		if player.ID == playerID {
			return player.Name
		}
	}
	return playerID // fallback
}

func (tm *tournamentManager) CancelTournament(tournamentID string) (*Tournament, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return nil, fmt.Errorf("tournament not found: %s", tournamentID)
	}

	if tournament.Status == StatusCompleted {
		return nil, fmt.Errorf("cannot cancel completed tournament")
	}

	tournament.Status = StatusCancelled
	now := time.Now()
	tournament.CompletedAt = &now

	return tournament, nil
}

func (tm *tournamentManager) GetTournament(tournamentID string) (*Tournament, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return nil, fmt.Errorf("tournament not found: %s", tournamentID)
	}

	// Return a copy to prevent external modification
	return tm.copyTournament(tournament), nil
}

func (tm *tournamentManager) GetActiveTournaments() ([]Tournament, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var active []Tournament
	for _, tournament := range tm.tournaments {
		if tournament.Status == StatusRegistering ||
			tournament.Status == StatusWaiting ||
			tournament.Status == StatusInProgress {
			active = append(active, *tm.copyTournament(tournament))
		}
	}

	return active, nil
}

func (tm *tournamentManager) FinalizeTournament(tournamentID string) (*Tournament, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return nil, fmt.Errorf("tournament not found: %s", tournamentID)
	}

	if tournament.Status != StatusCompleted {
		return nil, fmt.Errorf("tournament not completed yet")
	}

	// This is where we would:
	// 1. Send results to smart contract
	// 2. Trigger prize distribution
	// 3. Update blockchain state

	// For now, just mark as finalized (we'll implement blockchain integration later)

	return tournament, nil
}

// Event handling
func (tm *tournamentManager) AddEventHandler(handler func(TournamentEvent)) {
	tm.eventHandlers = append(tm.eventHandlers, handler)
}

func (tm *tournamentManager) broadcastEvent(event TournamentEvent) {
	for _, handler := range tm.eventHandlers {
		go handler(event) // Run in goroutine to avoid blocking
	}
}

// Helper methods
func (tm *tournamentManager) validateConfig(config TournamentConfig) error {
	if config.MaxPlayers < 2 || config.MaxPlayers > 4 {
		return fmt.Errorf("max players must be 2-4, got %d", config.MaxPlayers)
	}

	if config.EntryFee <= 0 {
		return fmt.Errorf("entry fee must be positive, got %d", config.EntryFee)
	}

	if len(config.PrizeStructure) == 0 {
		return fmt.Errorf("prize structure cannot be empty")
	}

	// Validate prize structure totals 100%
	totalPercentage := 0
	for _, prize := range config.PrizeStructure {
		totalPercentage += prize.Percentage
	}
	if totalPercentage != 100 {
		return fmt.Errorf("prize structure must total 100%%, got %d%%", totalPercentage)
	}

	return nil
}

func (tm *tournamentManager) copyTournament(tournament *Tournament) *Tournament {
	// Deep copy tournament to prevent external modification
	copy := *tournament

	// Copy slices
	copy.Participants = make([]Participant, len(tournament.Participants))
	copy(copy.Participants, tournament.Participants)

	copy.PrizeStructure = make([]PrizeLevel, len(tournament.PrizeStructure))
	copy(copy.PrizeStructure, tournament.PrizeStructure)

	copy.Results = make([]TournamentResult, len(tournament.Results))
	copy(copy.Results, tournament.Results)

	copy.Winners = make([]Winner, len(tournament.Winners))
	copy(copy.Winners, tournament.Winners)

	// CurrentGame would need deep copy too, but for simplicity we'll reference it
	// In production, you'd want to deep copy the game state as well

	return &copy
}
