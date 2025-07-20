package poker

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

type GameEngine struct {
	mu    sync.RWMutex
	games map[string]*GameState
	deck  *DeckManager
	eval  *HandEvaluator
}

func NewEngine() Engine {
	return &GameEngine{
		games: make(map[string]*GameState),
		deck:  NewDeckManager(),
		eval:  NewHandEvaluator(),
	}
}

func (e *GameEngine) NewGame(tableID string, playerNames []string, blinds Blinds) (*GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(playerNames) < 2 || len(playerNames) > 4 {
		return nil, fmt.Errorf("invalid player count: %d (must be 2-4)", len(playerNames))
	}

	// Generate unique game ID
	gameID := fmt.Sprintf("%s_%d", tableID, time.Now().Unix())

	// Create players with starting stacks
	players := make([]*Player, len(playerNames))
	startingStack := 1000 // TODO: make configurable
	
	for i, name := range playerNames {
		players[i] = &Player{
			ID:       fmt.Sprintf("player_%d", i),
			Name:     name,
			Stack:    startingStack,
			Position: i,
			Status:   StatusActive,
			Cards:    make([]Card, 0, 2),
		}
	}

	// Generate deck seed for reproducible shuffles
	seed := e.generateSeed()

	game := &GameState{
		ID:             gameID,
		TableID:        tableID,
		Players:        players,
		CommunityCards: make([]Card, 0, 5),
		Phase:          PhasePreflop,
		DealerPos:      0, // TODO: rotate for subsequent hands
		SmallBlind:     blinds.Small,
		BigBlind:       blinds.Big,
		DeckSeed:       seed,
		StartTime:      time.Now(),
		LastAction:     time.Now(),
	}

	// Deal hole cards
	if err := e.dealHoleCards(game); err != nil {
		return nil, fmt.Errorf("failed to deal cards: %w", err)
	}

	// Post blinds
	if err := e.postBlinds(game); err != nil {
		return nil, fmt.Errorf("failed to post blinds: %w", err)
	}

	// Set action position (first to act after big blind)
	game.ActionPos = (game.DealerPos + 3) % len(game.Players) // UTG in 4-max

	e.games[gameID] = game
	return game, nil
}

func (e *GameEngine) ProcessAction(gameID string, action PlayerAction) (*GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	if game.Phase == PhaseComplete {
		return game, fmt.Errorf("game already complete")
	}

	// Validate action
	if err := e.validateAction(game, action); err != nil {
		return nil, err
	}

	// Process the action
	if err := e.applyAction(game, action); err != nil {
		return nil, err
	}

	// Check if betting round is complete
	if e.isBettingRoundComplete(game) {
		if err := e.advanceToNextPhase(game); err != nil {
			return nil, err
		}
	} else {
		// Move to next active player
		e.advanceAction(game)
	}

	game.LastAction = time.Now()
	return game, nil
}

func (e *GameEngine) GetGameState(gameID string) (*GameState, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	game, exists := e.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	// Return a copy to prevent external modification
	return e.copyGameState(game), nil
}

// AutoCompleteGame - For MVP: everyone all-in preflop
func (e *GameEngine) AutoCompleteGame(gameID string) (*GameState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	game, exists := e.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	// For MVP: put everyone all-in and run to showdown
	for _, player := range game.Players {
		if player.Status == StatusActive {
			player.Status = StatusAllIn
			game.Pot += player.Stack
			player.TotalInvested += player.Stack
			player.Stack = 0
		}
	}

	// Deal all community cards
	if err := e.dealAllCommunityCards(game); err != nil {
		return nil, err
	}

	// Evaluate hands and determine winners
	if err := e.evaluateShowdown(game); err != nil {
		return nil, err
	}

	game.Phase = PhaseComplete
	return game, nil
}

// Helper methods
func (e *GameEngine) generateSeed() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func (e *GameEngine) dealHoleCards(game *GameState) error {
	shuffledDeck := e.deck.Shuffle(game.DeckSeed)
	cardIndex := 0

	// Deal 2 cards to each player
	for i := 0; i < 2; i++ {
		for _, player := range game.Players {
			if player.Status == StatusActive {
				player.Cards = append(player.Cards, shuffledDeck[cardIndex])
				cardIndex++
			}
		}
	}

	game.RemainingDeck = shuffledDeck[cardIndex:]
	return nil
}

func (e *GameEngine) postBlinds(game *GameState) error {
	playerCount := len(game.Players)
	
	// Small blind
	sbPos := (game.DealerPos + 1) % playerCount
	sbPlayer := game.Players[sbPos]
	sbAmount := min(game.SmallBlind, sbPlayer.Stack)
	sbPlayer.Stack -= sbAmount
	sbPlayer.CurrentBet = sbAmount
	sbPlayer.TotalInvested = sbAmount
	game.Pot += sbAmount

	// Big blind
	bbPos := (game.DealerPos + 2) % playerCount
	bbPlayer := game.Players[bbPos]
	bbAmount := min(game.BigBlind, bbPlayer.Stack)
	bbPlayer.Stack -= bbAmount
	bbPlayer.CurrentBet = bbAmount
	bbPlayer.TotalInvested = bbAmount
	game.Pot += bbAmount
	
	game.CurrentBet = bbAmount

	return nil
}

func (e *GameEngine) validateAction(game *GameState, action PlayerAction) error {
	// Find player
	var player *Player
	for _, p := range game.Players {
		if p.ID == action.PlayerID {
			player = p
			break
		}
	}
	if player == nil {
		return fmt.Errorf("player not found: %s", action.PlayerID)
	}

	// Check if it's player's turn
	if game.Players[game.ActionPos].ID != action.PlayerID {
		return fmt.Errorf("not player's turn")
	}

	// Check player status
	if player.Status != StatusActive {
		return fmt.Errorf("player not active")
	}

	// Validate specific actions (simplified for MVP)
	switch action.Type {
	case ActionFold:
		return nil // always valid
	case ActionCall:
		needed := game.CurrentBet - player.CurrentBet
		if needed > player.Stack {
			return fmt.Errorf("insufficient chips to call")
		}
		return nil
	case ActionAllIn:
		return nil // always valid if active
	default:
		return fmt.Errorf("action not supported in MVP: %s", action.Type)
	}
}

func (e *GameEngine) applyAction(game *GameState, action PlayerAction) error {
	player := game.Players[game.ActionPos]

	switch action.Type {
	case ActionFold:
		player.Status = StatusFolded
		
	case ActionCall:
		needed := game.CurrentBet - player.CurrentBet
		amount := min(needed, player.Stack)
		player.Stack -= amount
		player.CurrentBet += amount
		player.TotalInvested += amount
		game.Pot += amount
		
		if player.Stack == 0 {
			player.Status = StatusAllIn
		}
		
	case ActionAllIn:
		amount := player.Stack
		player.Stack = 0
		player.CurrentBet += amount
		player.TotalInvested += amount
		game.Pot += amount
		player.Status = StatusAllIn
		
		// Update current bet if this all-in is a raise
		if player.CurrentBet > game.CurrentBet {
			game.CurrentBet = player.CurrentBet
		}
	}

	player.HasActed = true
	return nil
}

func (e *GameEngine) isBettingRoundComplete(game *GameState) bool {
	activePlayers := 0
	allInPlayers := 0
	
	for _, player := range game.Players {
		switch player.Status {
		case StatusActive:
			activePlayers++
			if !player.HasActed || player.CurrentBet < game.CurrentBet {
				return false
			}
		case StatusAllIn:
			allInPlayers++
		}
	}
	
	// Round complete if <= 1 active player or all have acted and matched bet
	return activePlayers <= 1 || (activePlayers > 0 && allInPlayers >= 0)
}

func (e *GameEngine) advanceAction(game *GameState) {
	for i := 1; i <= len(game.Players); i++ {
		nextPos := (game.ActionPos + i) % len(game.Players)
		if game.Players[nextPos].Status == StatusActive {
			game.ActionPos = nextPos
			return
		}
	}
}

func (e *GameEngine) advanceToNextPhase(game *GameState) error {
	// Reset betting round state
	for _, player := range game.Players {
		player.HasActed = false
		player.CurrentBet = 0
	}
	game.CurrentBet = 0
	
	switch game.Phase {
	case PhasePreflop:
		game.Phase = PhaseFlop
		return e.dealFlop(game)
	case PhaseFlop:
		game.Phase = PhaseTurn
		return e.dealTurn(game)
	case PhaseTurn:
		game.Phase = PhaseRiver
		return e.dealRiver(game)
	case PhaseRiver:
		game.Phase = PhaseShowdown
		return e.evaluateShowdown(game)
	default:
		return fmt.Errorf("invalid phase transition from %s", game.Phase)
	}
}

func (e *GameEngine) dealFlop(game *GameState) error {
	// Burn one card, deal 3
	if len(game.RemainingDeck) < 4 {
		return fmt.Errorf("insufficient cards in deck")
	}
	
	game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
	game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1:4]...)
	game.RemainingDeck = game.RemainingDeck[4:]
	
	return nil
}

func (e *GameEngine) dealTurn(game *GameState) error {
	if len(game.RemainingDeck) < 2 {
		return fmt.Errorf("insufficient cards in deck")
	}
	
	game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
	game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1])
	game.RemainingDeck = game.RemainingDeck[2:]
	
	return nil
}

func (e *GameEngine) dealRiver(game *GameState) error {
	if len(game.RemainingDeck) < 2 {
		return fmt.Errorf("insufficient cards in deck")
	}
	
	game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
	game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1])
	game.RemainingDeck = game.RemainingDeck[2:]
	
	return nil
}

func (e *GameEngine) dealAllCommunityCards(game *GameState) error {
	// For MVP auto-complete: deal all remaining community cards at once
	if len(game.CommunityCards) == 0 {
		// Deal flop
		game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
		game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1:4]...)
		game.RemainingDeck = game.RemainingDeck[4:]
	}
	
	if len(game.CommunityCards) == 3 {
		// Deal turn
		game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
		game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1])
		game.RemainingDeck = game.RemainingDeck[2:]
	}
	
	if len(game.CommunityCards) == 4 {
		// Deal river
		game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
		game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1])
		game.RemainingDeck = game.RemainingDeck[2:]
	}
	
	return nil
}

func (e *GameEngine) evaluateShowdown(game *GameState) error {
	// Evaluate hands for all non-folded players
	var results []struct {
		player *Player
		hand   HandResult
	}
	
	for _, player := range game.Players {
		if player.Status != StatusFolded {
			hand := e.eval.EvaluateHand(player.Cards, game.CommunityCards)
			results = append(results, struct {
				player *Player
				hand   HandResult
			}{player, hand})
		}
	}
	
	// Determine winner(s) - simplified for MVP
	if len(results) == 0 {
		return fmt.Errorf("no active players for showdown")
	}
	
	// Find best hand
	bestHand := results[0].hand
	var winners []*Player
	
	for _, result := range results {
		if e.eval.CompareHands(result.hand, bestHand) > 0 {
			bestHand = result.hand
			winners = []*Player{result.player}
		} else if e.eval.CompareHands(result.hand, bestHand) == 0 {
			winners = append(winners, result.player)
		}
	}
	
	// Distribute pot (simplified - equal split for ties)
	winAmount := game.Pot / len(winners)
	
	for _, winner := range winners {
		game.Winners = append(game.Winners, Winner{
			PlayerID: winner.ID,
			Amount:   winAmount,
			Hand:     bestHand.Type,
			Cards:    bestHand.Cards,
		})
	}
	
	game.Showdown = true
	game.Phase = PhaseComplete
	
	return nil
}

func (e *GameEngine) copyGameState(game *GameState) *GameState {
	// Deep copy implementation
	copy := *game
	
	// Copy slices
	copy.Players = make([]*Player, len(game.Players))
	for i, p := range game.Players {
		playerCopy := *p
		playerCopy.Cards = make([]Card, len(p.Cards))
		copy(playerCopy.Cards, p.Cards)
		copy.Players[i] = &playerCopy
	}
	
	copy.CommunityCards = make([]Card, len(game.CommunityCards))
	copy(copy.CommunityCards, game.CommunityCards)
	
	copy.Winners = make([]Winner, len(game.Winners))
	copy(copy.Winners, game.Winners)
	
	return &copy
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}BigBlind, bbPlayer.Stack)
	bbPlayer.Stack -= bbAmount
	bbPlayer.CurrentBet = bbAmount
	bbPlayer.TotalInvested = bbAmount
	game.Pot += bbAmount
	
	game.CurrentBet = bbAmount

	return nil
}

func (e *GameEngine) validateAction(game *GameState, action PlayerAction) error {
	// Find player
	var player *Player
	for _, p := range game.Players {
		if p.ID == action.PlayerID {
			player = p
			break
		}
	}
	if player == nil {
		return fmt.Errorf("player not found: %s", action.PlayerID)
	}

	// Check if it's player's turn
	if game.Players[game.ActionPos].ID != action.PlayerID {
		return fmt.Errorf("not player's turn")
	}

	// Check player status
	if player.Status != StatusActive {
		return fmt.Errorf("player not active")
	}

	// Validate specific actions (simplified for MVP)
	switch action.Type {
	case ActionFold:
		return nil // always valid
	case ActionCall:
		needed := game.CurrentBet - player.CurrentBet
		if needed > player.Stack {
			return fmt.Errorf("insufficient chips to call")
		}
		return nil
	case ActionAllIn:
		return nil // always valid if active
	default:
		return fmt.Errorf("action not supported in MVP: %s", action.Type)
	}
}

func (e *GameEngine) applyAction(game *GameState, action PlayerAction) error {
	player := game.Players[game.ActionPos]

	switch action.Type {
	case ActionFold:
		player.Status = StatusFolded
		
	case ActionCall:
		needed := game.CurrentBet - player.CurrentBet
		amount := min(needed, player.Stack)
		player.Stack -= amount
		player.CurrentBet += amount
		player.TotalInvested += amount
		game.Pot += amount
		
		if player.Stack == 0 {
			player.Status = StatusAllIn
		}
		
	case ActionAllIn:
		amount := player.Stack
		player.Stack = 0
		player.CurrentBet += amount
		player.TotalInvested += amount
		game.Pot += amount
		player.Status = StatusAllIn
		
		// Update current bet if this all-in is a raise
		if player.CurrentBet > game.CurrentBet {
			game.CurrentBet = player.CurrentBet
		}
	}

	player.HasActed = true
	return nil
}

func (e *GameEngine) isBettingRoundComplete(game *GameState) bool {
	activePlayers := 0
	allInPlayers := 0
	
	for _, player := range game.Players {
		switch player.Status {
		case StatusActive:
			activePlayers++
			if !player.HasActed || player.CurrentBet < game.CurrentBet {
				return false
			}
		case StatusAllIn:
			allInPlayers++
		}
	}
	
	// Round complete if <= 1 active player or all have acted and matched bet
	return activePlayers <= 1 || (activePlayers > 0 && allInPlayers >= 0)
}

func (e *GameEngine) advanceAction(game *GameState) {
	for i := 1; i <= len(game.Players); i++ {
		nextPos := (game.ActionPos + i) % len(game.Players)
		if game.Players[nextPos].Status == StatusActive {
			game.ActionPos = nextPos
			return
		}
	}
}

func (e *GameEngine) advanceToNextPhase(game *GameState) error {
	// Reset betting round state
	for _, player := range game.Players {
		player.HasActed = false
		player.CurrentBet = 0
	}
	game.CurrentBet = 0
	
	switch game.Phase {
	case PhasePreflop:
		game.Phase = PhaseFlop
		return e.dealFlop(game)
	case PhaseFlop:
		game.Phase = PhaseTurn
		return e.dealTurn(game)
	case PhaseTurn:
		game.Phase = PhaseRiver
		return e.dealRiver(game)
	case PhaseRiver:
		game.Phase = PhaseShowdown
		return e.evaluateShowdown(game)
	default:
		return fmt.Errorf("invalid phase transition from %s", game.Phase)
	}
}

func (e *GameEngine) dealFlop(game *GameState) error {
	// Burn one card, deal 3
	if len(game.RemainingDeck) < 4 {
		return fmt.Errorf("insufficient cards in deck")
	}
	
	game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
	game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1:4]...)
	game.RemainingDeck = game.RemainingDeck[4:]
	
	return nil
}

func (e *GameEngine) dealTurn(game *GameState) error {
	if len(game.RemainingDeck) < 2 {
		return fmt.Errorf("insufficient cards in deck")
	}
	
	game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
	game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1])
	game.RemainingDeck = game.RemainingDeck[2:]
	
	return nil
}

func (e *GameEngine) dealRiver(game *GameState) error {
	if len(game.RemainingDeck) < 2 {
		return fmt.Errorf("insufficient cards in deck")
	}
	
	game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
	game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1])
	game.RemainingDeck = game.RemainingDeck[2:]
	
	return nil
}

func (e *GameEngine) dealAllCommunityCards(game *GameState) error {
	// For MVP auto-complete: deal all remaining community cards at once
	if len(game.CommunityCards) == 0 {
		// Deal flop
		game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
		game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1:4]...)
		game.RemainingDeck = game.RemainingDeck[4:]
	}
	
	if len(game.CommunityCards) == 3 {
		// Deal turn
		game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
		game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1])
		game.RemainingDeck = game.RemainingDeck[2:]
	}
	
	if len(game.CommunityCards) == 4 {
		// Deal river
		game.BurnCards = append(game.BurnCards, game.RemainingDeck[0])
		game.CommunityCards = append(game.CommunityCards, game.RemainingDeck[1])
		game.RemainingDeck = game.RemainingDeck[2:]
	}
	
	return nil
}

func (e *GameEngine) evaluateShowdown(game *GameState) error {
	// Evaluate hands for all non-folded players
	var results []struct {
		player *Player
		hand   HandResult
	}
	
	for _, player := range game.Players {
		if player.Status != StatusFolded {
			hand := e.eval.EvaluateHand(player.Cards, game.CommunityCards)
			results = append(results, struct {
				player *Player
				hand   HandResult
			}{player, hand})
		}
	}
	
	// Determine winner(s) - simplified for MVP
	if len(results) == 0 {
		return fmt.Errorf("no active players for showdown")
	}
	
	// Find best hand
	bestHand := results[0].hand
	var winners []*Player
	
	for _, result := range results {
		if e.eval.CompareHands(result.hand, bestHand) > 0 {
			bestHand = result.hand
			winners = []*Player{result.player}
		} else if e.eval.CompareHands(result.hand, bestHand) == 0 {
			winners = append(winners, result.player)
		}
	}
	
	// Distribute pot (simplified - equal split for ties)
	winAmount := game.Pot / len(winners)
	
	for _, winner := range winners {
		game.Winners = append(game.Winners, Winner{
			PlayerID: winner.ID,
			Amount:   winAmount,
			Hand:     bestHand.Type,
			Cards:    bestHand.Cards,
		})
	}
	
	game.Showdown = true
	game.Phase = PhaseComplete
	
	return nil
}

func (e *GameEngine) copyGameState(game *GameState) *GameState {
	// Deep copy implementation
	copy := *game
	
	// Copy slices
	copy.Players = make([]*Player, len(game.Players))
	for i, p := range game.Players {
		playerCopy := *p
		playerCopy.Cards = make([]Card, len(p.Cards))
		copy(playerCopy.Cards, p.Cards)
		copy.Players[i] = &playerCopy
	}
	
	copy.CommunityCards = make([]Card, len(game.CommunityCards))
	copy(copy.CommunityCards, game.CommunityCards)
	
	copy.Winners = make([]Winner, len(game.Winners))
	copy(copy.Winners, game.Winners)
	
	return &copy
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}