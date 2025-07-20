package poker

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// Card representa una carta
type Card struct {
	Suit string `json:"suit"` // hearts, diamonds, clubs, spades
	Rank string `json:"rank"` // 2-10, J, Q, K, A
}

// Player representa un jugador en la mesa
type PokerPlayer struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Stack      int    `json:"stack"`
	Cards      []Card `json:"cards"`
	Position   int    `json:"position"`
	IsActive   bool   `json:"is_active"`
	HasFolded  bool   `json:"has_folded"`
	CurrentBet int    `json:"current_bet"`
}

// PokerTable representa el estado completo de una mesa de poker
type PokerTable struct {
	ID             string        `json:"id"`
	Players        []PokerPlayer `json:"players"`
	CommunityCards []Card        `json:"community_cards"`
	Pot            int           `json:"pot"`
	CurrentPlayer  int           `json:"current_player"`
	Phase          string        `json:"phase"` // preflop, flop, turn, river, showdown
	Deck           []Card        `json:"-"`     // No enviar en JSON
	StartTime      time.Time     `json:"start_time"`
	SmallBlind     int           `json:"small_blind"`
	BigBlind       int           `json:"big_blind"`
	DealerPosition int           `json:"dealer_position"`
}

// PokerEngine maneja la lógica del poker
type PokerEngine struct {
	tables map[string]*PokerTable
}

func NewPokerEngine() *PokerEngine {
	return &PokerEngine{
		tables: make(map[string]*PokerTable),
	}
}

// CreateTable crea una nueva mesa de poker
func (pe *PokerEngine) CreateTable(tableID string) *PokerTable {
	table := &PokerTable{
		ID:             tableID,
		Players:        make([]PokerPlayer, 0, 10), // Soportar hasta 10 jugadores
		CommunityCards: make([]Card, 0, 5),
		Pot:            0,
		CurrentPlayer:  0,
		Phase:          "waiting",
		Deck:           pe.createShuffledDeck(),
		StartTime:      time.Now(),
		SmallBlind:     10,  // Default blinds
		BigBlind:       20,
		DealerPosition: 0,
	}
	pe.tables[tableID] = table
	return table
}

// AddPlayer agrega un jugador a la mesa
func (pe *PokerEngine) AddPlayer(tableID, playerID, playerName string) (*PokerTable, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		table = pe.CreateTable(tableID)
	}

	// Verificar si el jugador ya existe
	for _, player := range table.Players {
		if player.ID == playerID {
			return table, fmt.Errorf("player already at table")
		}
	}

	// Verificar límite de jugadores
	if len(table.Players) >= 10 {
		return table, fmt.Errorf("table is full")
	}

	// Agregar jugador
	player := PokerPlayer{
		ID:         playerID,
		Name:       playerName,
		Stack:      1000, // Stack inicial
		Cards:      make([]Card, 0, 2),
		Position:   len(table.Players),
		IsActive:   true,
		HasFolded:  false,
		CurrentBet: 0,
	}

	table.Players = append(table.Players, player)

	// Iniciar juego si hay al menos 2 jugadores
	if len(table.Players) >= 2 && table.Phase == "waiting" {
		pe.startHand(table)
	}

	return table, nil
}

// startHand inicia una nueva mano
func (pe *PokerEngine) startHand(table *PokerTable) {
	// Reiniciar deck
	table.Deck = pe.createShuffledDeck()
	table.CommunityCards = make([]Card, 0, 5)
	table.Pot = 0
	table.Phase = "preflop"

	// Contar jugadores activos
	activePlayers := make([]int, 0)
	for i := range table.Players {
		table.Players[i].Cards = make([]Card, 0, 2)
		table.Players[i].HasFolded = false
		table.Players[i].CurrentBet = 0
		table.Players[i].IsActive = table.Players[i].Stack > 0
		
		if table.Players[i].IsActive {
			activePlayers = append(activePlayers, i)
		}
	}

	if len(activePlayers) < 2 {
		table.Phase = "waiting"
		return
	}

	// Avanzar dealer position
	table.DealerPosition = (table.DealerPosition + 1) % len(activePlayers)

	// Repartir cartas (2 por jugador)
	pe.dealCards(table)

	// Colocar blinds
	pe.postBlinds(table, activePlayers)

	// Establecer primer jugador (después del big blind)
	if len(activePlayers) > 2 {
		table.CurrentPlayer = (table.DealerPosition + 3) % len(activePlayers)
	} else {
		// Heads-up: dealer actúa primero preflop
		table.CurrentPlayer = table.DealerPosition
	}
}

// postBlinds coloca los blinds automáticamente
func (pe *PokerEngine) postBlinds(table *PokerTable, activePlayers []int) {
	if len(activePlayers) < 2 {
		return
	}

	// Small blind (izquierda del dealer)
	sbPosition := (table.DealerPosition + 1) % len(activePlayers)
	sbPlayerIndex := activePlayers[sbPosition]
	sbAmount := table.SmallBlind
	if table.Players[sbPlayerIndex].Stack < sbAmount {
		sbAmount = table.Players[sbPlayerIndex].Stack
	}
	table.Players[sbPlayerIndex].Stack -= sbAmount
	table.Players[sbPlayerIndex].CurrentBet = sbAmount
	table.Pot += sbAmount

	// Big blind (izquierda del small blind)
	bbPosition := (table.DealerPosition + 2) % len(activePlayers)
	bbPlayerIndex := activePlayers[bbPosition]
	bbAmount := table.BigBlind
	if table.Players[bbPlayerIndex].Stack < bbAmount {
		bbAmount = table.Players[bbPlayerIndex].Stack
	}
	table.Players[bbPlayerIndex].Stack -= bbAmount
	table.Players[bbPlayerIndex].CurrentBet = bbAmount
	table.Pot += bbAmount
}

// dealCards reparte cartas a los jugadores
func (pe *PokerEngine) dealCards(table *PokerTable) {
	cardIndex := 0

	// 2 cartas por jugador
	for round := 0; round < 2; round++ {
		for i := range table.Players {
			if table.Players[i].IsActive {
				table.Players[i].Cards = append(table.Players[i].Cards, table.Deck[cardIndex])
				cardIndex++
			}
		}
	}

	// Remover cartas repartidas del deck
	table.Deck = table.Deck[cardIndex:]
}

// PlayerAction procesa una acción del jugador
func (pe *PokerEngine) PlayerAction(tableID, playerID, action string, amount int) (*PokerTable, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		return nil, fmt.Errorf("table not found")
	}

	// Encontrar jugador
	playerIndex := -1
	for i, player := range table.Players {
		if player.ID == playerID {
			playerIndex = i
			break
		}
	}

	if playerIndex == -1 {
		return nil, fmt.Errorf("player not found")
	}

	// Verificar turno
	if table.CurrentPlayer != playerIndex {
		return nil, fmt.Errorf("not your turn")
	}

	player := &table.Players[playerIndex]

	// Procesar acción
	switch action {
	case "fold":
		player.HasFolded = true
		player.IsActive = false

	case "call":
		// Por simplicidad, call = 10 fichas
		callAmount := 10
		if player.Stack < callAmount {
			callAmount = player.Stack
		}
		player.Stack -= callAmount
		player.CurrentBet += callAmount
		table.Pot += callAmount

	case "raise":
		if amount <= 0 || amount > player.Stack {
			return nil, fmt.Errorf("invalid raise amount")
		}
		player.Stack -= amount
		player.CurrentBet += amount
		table.Pot += amount

	case "all_in":
		amount = player.Stack
		player.Stack = 0
		player.CurrentBet += amount
		table.Pot += amount

	default:
		return nil, fmt.Errorf("invalid action: %s", action)
	}

	// Avanzar al siguiente jugador
	pe.nextPlayer(table)

	// Verificar si la mano terminó
	if pe.isHandComplete(table) {
		pe.completeHand(table)
	}

	return table, nil
}

// nextPlayer avanza al siguiente jugador activo
func (pe *PokerEngine) nextPlayer(table *PokerTable) {
	originalPlayer := table.CurrentPlayer

	for {
		table.CurrentPlayer = (table.CurrentPlayer + 1) % len(table.Players)

		// Si volvimos al jugador original sin encontrar uno activo, la mano terminó
		if table.CurrentPlayer == originalPlayer {
			break
		}

		// Si encontramos un jugador activo, salimos
		if table.Players[table.CurrentPlayer].IsActive && !table.Players[table.CurrentPlayer].HasFolded {
			break
		}
	}
}

// isHandComplete verifica si la mano ha terminado
func (pe *PokerEngine) isHandComplete(table *PokerTable) bool {
	activePlayers := 0
	for _, player := range table.Players {
		if player.IsActive && !player.HasFolded {
			activePlayers++
		}
	}
	return activePlayers <= 1
}

// completeHand termina la mano y determina ganador
func (pe *PokerEngine) completeHand(table *PokerTable) {
	table.Phase = "showdown"

	// Determinar ganadores usando el evaluador de manos
	winners := DetermineWinners(table)
	
	if len(winners) > 0 {
		// Dividir el pot entre los ganadores
		potPerWinner := table.Pot / len(winners)
		remainder := table.Pot % len(winners)
		
		for i, winnerIndex := range winners {
			table.Players[winnerIndex].Stack += potPerWinner
			// Dar el resto al primer ganador
			if i == 0 {
				table.Players[winnerIndex].Stack += remainder
			}
		}
		
		table.Pot = 0
	}

	// Preparar próxima mano
	time.AfterFunc(3*time.Second, func() {
		if pe.hasEnoughActivePlayers(table) {
			pe.startHand(table)
		} else {
			table.Phase = "waiting"
		}
	})
}

// hasEnoughActivePlayers verifica si hay suficientes jugadores para continuar
func (pe *PokerEngine) hasEnoughActivePlayers(table *PokerTable) bool {
	activeCount := 0
	for _, player := range table.Players {
		if player.IsActive && player.Stack > 0 {
			activeCount++
		}
	}
	return activeCount >= 2
}

// createShuffledDeck crea y baraja un deck estándar
func (pe *PokerEngine) createShuffledDeck() []Card {
	suits := []string{"hearts", "diamonds", "clubs", "spades"}
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

	deck := make([]Card, 0, 52)
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}

	// Barajar usando crypto/rand para seguridad
	for i := len(deck) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		deck[i], deck[j.Int64()] = deck[j.Int64()], deck[i]
	}

	return deck
}

// GetTable obtiene el estado de una mesa
func (pe *PokerEngine) GetTable(tableID string) (*PokerTable, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		return nil, fmt.Errorf("table not found")
	}
	return table, nil
}
