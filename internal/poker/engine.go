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
	IsReady    bool   `json:"is_ready"`    // Nuevo: ¿Está listo para jugar?
	IsHost     bool   `json:"is_host"`     // Nuevo: ¿Es el host de la mesa?
}

// PokerTable representa el estado completo de una mesa de poker
type PokerTable struct {
	ID               string        `json:"id"`
	Players          []PokerPlayer `json:"players"`
	CommunityCards   []Card        `json:"community_cards"`
	Pot              int           `json:"pot"`
	CurrentPlayer    int           `json:"current_player"`
	Phase            string        `json:"phase"` // lobby, preflop, flop, turn, river, showdown
	Deck             []Card        `json:"-"`     // No enviar en JSON
	StartTime        time.Time     `json:"start_time"`
	SmallBlind       int           `json:"small_blind"`
	BigBlind         int           `json:"big_blind"`
	DealerPosition   int           `json:"dealer_position"`
	CurrentBet       int           `json:"current_bet"`       // Apuesta actual más alta en esta ronda
	LastRaiser       int           `json:"last_raiser"`       // Índice del último jugador que subió
	PlayersToAct     []bool        `json:"players_to_act"`    // Qué jugadores necesitan actuar en esta ronda
	BettingComplete  bool          `json:"betting_complete"`  // Si la ronda de apuestas está completa
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

	// Determinar si es el host (primer jugador)
	isHost := len(table.Players) == 0
	
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
		IsReady:    false, // Por defecto no está listo
		IsHost:     isHost, // Primer jugador es host
	}

	table.Players = append(table.Players, player)

	// YA NO auto-start - Solo cambiar fase si está "waiting" → "lobby"
	if table.Phase == "waiting" && len(table.Players) >= 1 {
		table.Phase = "lobby" // Nuevo estado: esperando que los jugadores estén listos
	} else if table.Phase != "waiting" && table.Phase != "lobby" {
		// Si hay una mano en progreso, el jugador debe esperar a la siguiente mano
		// Marcar como inactivo hasta la siguiente mano
		table.Players[len(table.Players)-1].IsActive = false
	}

	return table, nil
}

// SetPlayerReady marca a un jugador como listo/no listo
func (pe *PokerEngine) SetPlayerReady(tableID, playerID string, ready bool) (*PokerTable, error) {
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

	// Solo se puede marcar ready en lobby
	if table.Phase != "lobby" {
		return nil, fmt.Errorf("can only set ready status in lobby")
	}

	table.Players[playerIndex].IsReady = ready
	return table, nil
}

// StartGame inicia el juego manualmente (solo por el host)
func (pe *PokerEngine) StartGame(tableID, playerID string) (*PokerTable, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		return nil, fmt.Errorf("table not found")
	}

	// Verificar que sea el host
	var isHost bool
	for _, player := range table.Players {
		if player.ID == playerID && player.IsHost {
			isHost = true
			break
		}
	}

	if !isHost {
		return nil, fmt.Errorf("only the host can start the game")
	}

	// Verificar que estemos en lobby
	if table.Phase != "lobby" {
		return nil, fmt.Errorf("game can only be started from lobby")
	}

	// Verificar que haya al menos 2 jugadores
	if len(table.Players) < 2 {
		return nil, fmt.Errorf("need at least 2 players to start")
	}

	// Verificar que todos estén listos
	for _, player := range table.Players {
		if !player.IsReady {
			return nil, fmt.Errorf("all players must be ready to start (player %s is not ready)", player.Name)
		}
	}

	// Iniciar el juego
	pe.startHand(table)
	return table, nil
}

// GetReadyStatus obtiene el estado de "ready" de todos los jugadores
func (pe *PokerEngine) GetReadyStatus(tableID string) (map[string]bool, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		return nil, fmt.Errorf("table not found")
	}

	status := make(map[string]bool)
	for _, player := range table.Players {
		status[player.Name] = player.IsReady
	}

	return status, nil
}

// startHand inicia una nueva mano
func (pe *PokerEngine) startHand(table *PokerTable) {
	// Reiniciar deck
	table.Deck = pe.createShuffledDeck()
	table.CommunityCards = make([]Card, 0, 5)
	table.Pot = 0
	table.Phase = "preflop"
	table.CurrentBet = table.BigBlind // La apuesta inicial es el big blind
	table.LastRaiser = -1
	table.BettingComplete = false

	// Contar jugadores activos y reactivar a todos los que tienen fichas
	activePlayers := make([]int, 0)
	for i := range table.Players {
		table.Players[i].Cards = make([]Card, 0, 2)
		table.Players[i].HasFolded = false
		table.Players[i].CurrentBet = 0
		// Reactivar todos los jugadores que tienen fichas (incluyendo los que llegaron durante la mano anterior)
		table.Players[i].IsActive = table.Players[i].Stack > 0
		
		if table.Players[i].IsActive {
			activePlayers = append(activePlayers, i)
		}
	}

	// Inicializar array de jugadores que necesitan actuar
	table.PlayersToAct = make([]bool, len(table.Players))
	for _, playerIndex := range activePlayers {
		table.PlayersToAct[playerIndex] = true
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

	// Los blinds ya han "actuado" para esta ronda preflop
	// Pero el big blind puede aún hacer raise si vuelve a él
	table.PlayersToAct[sbPlayerIndex] = false // Small blind ya puso su apuesta obligatoria
	table.PlayersToAct[bbPlayerIndex] = true  // Big blind puede hacer raise cuando le toque
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

	// Procesar acción según el Texas Hold'em real
	switch action {
	case "fold":
		player.HasFolded = true
		player.IsActive = false

	case "call":
		// Call real: igualar la apuesta más alta actual
		callAmount := table.CurrentBet - player.CurrentBet
		if callAmount <= 0 {
			return nil, fmt.Errorf("no hay apuesta que igualar")
		}
		if player.Stack < callAmount {
			// All-in automático si no tiene suficientes fichas
			callAmount = player.Stack
		}
		player.Stack -= callAmount
		player.CurrentBet += callAmount
		table.Pot += callAmount

	case "check":
		// Check solo es válido si no hay apuesta que igualar
		if table.CurrentBet > player.CurrentBet {
			return nil, fmt.Errorf("no puedes hacer check, hay una apuesta que igualar")
		}
		// No hacer nada, es solo pasar el turno

	case "raise":
		// Raise: igualar la apuesta actual + subir
		callAmount := table.CurrentBet - player.CurrentBet
		totalAmount := callAmount + amount
		
		if amount <= 0 {
			return nil, fmt.Errorf("el raise debe ser positivo")
		}
		if totalAmount > player.Stack {
			return nil, fmt.Errorf("no tienes suficientes fichas para este raise")
		}
		if amount < table.BigBlind {
			return nil, fmt.Errorf("el raise mínimo es %d", table.BigBlind)
		}

		player.Stack -= totalAmount
		player.CurrentBet += totalAmount
		table.Pot += totalAmount
		table.CurrentBet = player.CurrentBet
		table.LastRaiser = playerIndex

		// Reactivar a todos los jugadores que no han foldeado para que respondan al raise
		for i := range table.Players {
			if table.Players[i].IsActive && !table.Players[i].HasFolded && i != playerIndex {
				table.PlayersToAct[i] = true
			}
		}

	case "all_in":
		// All-in: apostar todas las fichas
		amount = player.Stack
		
		player.Stack = 0
		player.CurrentBet += amount
		table.Pot += amount

		// Si el all-in es mayor que la apuesta actual, es un raise
		if player.CurrentBet > table.CurrentBet {
			table.CurrentBet = player.CurrentBet
			table.LastRaiser = playerIndex
			// Reactivar jugadores para que respondan
			for i := range table.Players {
				if table.Players[i].IsActive && !table.Players[i].HasFolded && i != playerIndex {
					table.PlayersToAct[i] = true
				}
			}
		}

	default:
		return nil, fmt.Errorf("acción inválida: %s", action)
	}

	// Marcar que este jugador ya actuó en esta ronda
	table.PlayersToAct[playerIndex] = false

	// Verificar si la ronda de apuestas terminó
	if pe.isBettingRoundComplete(table) {
		pe.advanceToNextPhase(table)
	} else {
		// Avanzar al siguiente jugador
		pe.nextPlayer(table)
	}

	// Verificar si la mano terminó completamente
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

// isBettingRoundComplete verifica si la ronda de apuestas actual ha terminado
func (pe *PokerEngine) isBettingRoundComplete(table *PokerTable) bool {
	// Contar jugadores activos que no han foldeado
	activePlayers := 0
	for i, player := range table.Players {
		if player.IsActive && !player.HasFolded {
			activePlayers++
			// Si algún jugador activo aún necesita actuar, la ronda no ha terminado
			if table.PlayersToAct[i] {
				return false
			}
		}
	}
	
	// Si solo queda 1 jugador activo, la mano termina (no solo la ronda)
	if activePlayers <= 1 {
		return true
	}

	// Verificar que todas las apuestas estén igualadas
	for _, player := range table.Players {
		if player.IsActive && !player.HasFolded {
			// Si tienen fichas y su apuesta no está igualada, la ronda no ha terminado
			if player.Stack > 0 && player.CurrentBet != table.CurrentBet {
				return false
			}
		}
	}

	return true
}

// advanceToNextPhase avanza a la siguiente fase del juego (flop, turn, river, showdown)
func (pe *PokerEngine) advanceToNextPhase(table *PokerTable) {
	// Resetear las apuestas para la nueva ronda
	for i := range table.Players {
		table.Players[i].CurrentBet = 0
		// Reactivar jugadores que siguen en la mano
		if table.Players[i].IsActive && !table.Players[i].HasFolded {
			table.PlayersToAct[i] = true
		}
	}
	
	table.CurrentBet = 0
	table.LastRaiser = -1

	switch table.Phase {
	case "preflop":
		// Repartir el flop (3 cartas)
		pe.dealFlop(table)
		table.Phase = "flop"
		// En post-flop, el primer jugador después del dealer actúa primero
		pe.setFirstPlayerPostFlop(table)

	case "flop":
		// Repartir el turn (1 carta)
		pe.dealTurn(table)
		table.Phase = "turn"
		pe.setFirstPlayerPostFlop(table)

	case "turn":
		// Repartir el river (1 carta)
		pe.dealRiver(table)
		table.Phase = "river"
		pe.setFirstPlayerPostFlop(table)

	case "river":
		// Ir al showdown
		table.Phase = "showdown"
		pe.completeHand(table)
	}
}

// dealFlop reparte las primeras 3 cartas comunitarias
func (pe *PokerEngine) dealFlop(table *PokerTable) {
	// Quemar una carta (descartar)
	if len(table.Deck) > 0 {
		table.Deck = table.Deck[1:]
	}
	
	// Repartir 3 cartas
	for i := 0; i < 3 && len(table.Deck) > 0; i++ {
		table.CommunityCards = append(table.CommunityCards, table.Deck[0])
		table.Deck = table.Deck[1:]
	}
}

// dealTurn reparte la 4ta carta comunitaria
func (pe *PokerEngine) dealTurn(table *PokerTable) {
	// Quemar una carta
	if len(table.Deck) > 0 {
		table.Deck = table.Deck[1:]
	}
	
	// Repartir 1 carta
	if len(table.Deck) > 0 {
		table.CommunityCards = append(table.CommunityCards, table.Deck[0])
		table.Deck = table.Deck[1:]
	}
}

// dealRiver reparte la 5ta carta comunitaria
func (pe *PokerEngine) dealRiver(table *PokerTable) {
	// Quemar una carta
	if len(table.Deck) > 0 {
		table.Deck = table.Deck[1:]
	}
	
	// Repartir 1 carta
	if len(table.Deck) > 0 {
		table.CommunityCards = append(table.CommunityCards, table.Deck[0])
		table.Deck = table.Deck[1:]
	}
}

// setFirstPlayerPostFlop establece el primer jugador que actúa después del flop
func (pe *PokerEngine) setFirstPlayerPostFlop(table *PokerTable) {
	// En post-flop, el primer jugador activo después del dealer actúa primero
	for i := 1; i <= len(table.Players); i++ {
		playerIndex := (table.DealerPosition + i) % len(table.Players)
		if table.Players[playerIndex].IsActive && !table.Players[playerIndex].HasFolded {
			table.CurrentPlayer = playerIndex
			return
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

	// Ya NO auto-restart - Las nuevas manos se inician manualmente
	// El juego permanece en showdown hasta que se inicie manualmente una nueva ronda
	// Esto permite que los jugadores vean los resultados y decidan cuándo continuar
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

// GetTableForPlayer obtiene el estado de la mesa filtrando cartas privadas
// Solo muestra las cartas del jugador solicitante, ocultando las de otros jugadores
func (pe *PokerEngine) GetTableForPlayer(tableID, playerID string) (*PokerTable, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		return nil, fmt.Errorf("table not found")
	}

	// Crear una copia del estado de la mesa
	filteredTable := *table
	filteredTable.Players = make([]PokerPlayer, len(table.Players))
	
	// Copiar jugadores pero filtrar cartas privadas
	for i, player := range table.Players {
		filteredTable.Players[i] = player
		
		// Solo mostrar cartas del jugador solicitante
		if player.ID != playerID {
			// Ocultar cartas de otros jugadores
			filteredTable.Players[i].Cards = make([]Card, len(player.Cards))
			// Mantener el número de cartas pero sin mostrar los valores
			for j := range player.Cards {
				filteredTable.Players[i].Cards[j] = Card{Suit: "hidden", Rank: "?"}
			}
		}
	}
	
	return &filteredTable, nil
}
