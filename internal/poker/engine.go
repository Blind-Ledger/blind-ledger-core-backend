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

// SidePot representa un pot lateral para all-ins múltiples
type SidePot struct {
	Amount           int   `json:"amount"`             // Cantidad total en este side pot
	EligiblePlayers  []int `json:"eligible_players"`   // Índices de jugadores elegibles para este pot
	MaxBetLevel      int   `json:"max_bet_level"`      // Nivel máximo de apuesta para este pot
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
	IsReady      bool      `json:"is_ready"`         // Nuevo: ¿Está listo para jugar?
	IsHost       bool      `json:"is_host"`          // Nuevo: ¿Es el host de la mesa?
	IsAllIn      bool      `json:"is_all_in"`        // Nuevo: ¿Está en all-in?
	IsConnected  bool      `json:"is_connected"`     // Nuevo: ¿Está conectado?
	LastSeenTime time.Time `json:"-"`                // Último momento visto (no enviar en JSON)
}

// PokerTable representa el estado completo de una mesa de poker
type PokerTable struct {
	ID               string        `json:"id"`
	Players          []PokerPlayer `json:"players"`
	CommunityCards   []Card        `json:"community_cards"`
	Pot              int           `json:"pot"`              // Pot principal (legacy, para compatibilidad)
	SidePots         []SidePot     `json:"side_pots"`        // Sistema de side pots para all-ins múltiples
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
	AutoRestart      bool          `json:"auto_restart"`      // Si las manos se reinician automáticamente
	ShowdownEndTime  time.Time     `json:"-"`                 // Tiempo cuando terminó el showdown
	RestartDelay     time.Duration `json:"-"`                 // Retraso antes del auto-restart (ej: 5 segundos)
	
	// Configuración de Buy-in
	BuyInAmount      int           `json:"buy_in_amount"`     // Cantidad estándar de buy-in
	MinBuyIn         int           `json:"min_buy_in"`        // Buy-in mínimo permitido
	MaxBuyIn         int           `json:"max_buy_in"`        // Buy-in máximo permitido
	IsCashGame       bool          `json:"is_cash_game"`      // true = cash game, false = torneo
}

// TableConfig representa la configuración para crear una mesa personalizada
type TableConfig struct {
	SmallBlind   int           `json:"small_blind"`   // Blind pequeño
	BigBlind     int           `json:"big_blind"`     // Blind grande
	BuyInAmount  int           `json:"buy_in_amount"` // Cantidad estándar de buy-in
	MinBuyIn     int           `json:"min_buy_in"`    // Buy-in mínimo permitido
	MaxBuyIn     int           `json:"max_buy_in"`    // Buy-in máximo permitido
	IsCashGame   bool          `json:"is_cash_game"`  // true = cash game, false = torneo
	AutoRestart  bool          `json:"auto_restart"`  // Si las manos se reinician automáticamente
	RestartDelay time.Duration `json:"restart_delay"` // Retraso antes del auto-restart
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

// CreateTable crea una nueva mesa de poker con configuración estándar
func (pe *PokerEngine) CreateTable(tableID string) *PokerTable {
	table := &PokerTable{
		ID:             tableID,
		Players:        make([]PokerPlayer, 0, 10), // Soportar hasta 10 jugadores
		CommunityCards: make([]Card, 0, 5),
		Pot:            0,
		SidePots:       make([]SidePot, 0),         // Inicializar sistema de side pots
		CurrentPlayer:  0,
		Phase:          "waiting",
		Deck:           pe.createShuffledDeck(),
		StartTime:      time.Now(),
		SmallBlind:     10,  // Default blinds
		BigBlind:       20,
		DealerPosition: 0,
		AutoRestart:    true,              // Por defecto auto-restart habilitado
		RestartDelay:   5 * time.Second,   // 5 segundos de delay por defecto
		
		// Configuración de Buy-in por defecto
		BuyInAmount:    1000,              // Buy-in estándar de 1000
		MinBuyIn:       500,               // Mínimo 500 (50BB)
		MaxBuyIn:       2000,              // Máximo 2000 (100BB)
		IsCashGame:     true,              // Por defecto cash game
	}
	pe.tables[tableID] = table
	return table
}

// CreateTableWithConfig crea una mesa con configuración personalizada
func (pe *PokerEngine) CreateTableWithConfig(tableID string, config TableConfig) *PokerTable {
	table := &PokerTable{
		ID:             tableID,
		Players:        make([]PokerPlayer, 0, 10),
		CommunityCards: make([]Card, 0, 5),
		Pot:            0,
		SidePots:       make([]SidePot, 0),
		CurrentPlayer:  0,
		Phase:          "waiting",
		Deck:           pe.createShuffledDeck(),
		StartTime:      time.Now(),
		SmallBlind:     config.SmallBlind,
		BigBlind:       config.BigBlind,
		DealerPosition: 0,
		AutoRestart:    config.AutoRestart,
		RestartDelay:   config.RestartDelay,
		
		// Configuración de Buy-in personalizada
		BuyInAmount:    config.BuyInAmount,
		MinBuyIn:       config.MinBuyIn,
		MaxBuyIn:       config.MaxBuyIn,
		IsCashGame:     config.IsCashGame,
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
		ID:           playerID,
		Name:         playerName,
		Stack:        1000, // Stack inicial
		Cards:        make([]Card, 0, 2),
		Position:     len(table.Players),
		IsActive:     true,
		HasFolded:    false,
		CurrentBet:   0,
		IsReady:      false, // Por defecto no está listo
		IsHost:       isHost, // Primer jugador es host
		IsConnected:  true,   // Conectado al agregarlo
		LastSeenTime: time.Now(),
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

// AddPlayerWithBuyIn agrega un jugador a la mesa con un buy-in personalizado
func (pe *PokerEngine) AddPlayerWithBuyIn(tableID, playerID, playerName string, buyInAmount int) (*PokerTable, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		table = pe.CreateTable(tableID)
	}

	// Validar buy-in amount
	if buyInAmount < table.MinBuyIn {
		return table, fmt.Errorf("buy-in amount %d is below minimum %d", buyInAmount, table.MinBuyIn)
	}
	if buyInAmount > table.MaxBuyIn {
		return table, fmt.Errorf("buy-in amount %d is above maximum %d", buyInAmount, table.MaxBuyIn)
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
	
	// Agregar jugador con buy-in personalizado
	player := PokerPlayer{
		ID:           playerID,
		Name:         playerName,
		Stack:        buyInAmount, // Stack inicial basado en buy-in
		Cards:        make([]Card, 0, 2),
		Position:     len(table.Players),
		IsActive:     true,
		HasFolded:    false,
		CurrentBet:   0,
		IsReady:      false,
		IsHost:       isHost,
		IsConnected:  true,
		LastSeenTime: time.Now(),
	}

	table.Players = append(table.Players, player)

	// Cambiar fase si está "waiting" → "lobby"
	if table.Phase == "waiting" && len(table.Players) >= 1 {
		table.Phase = "lobby"
	} else if table.Phase != "waiting" && table.Phase != "lobby" {
		// Si hay una mano en progreso, el jugador debe esperar a la siguiente mano
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
	table.SidePots = make([]SidePot, 0) // Reiniciar side pots para nueva mano
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
		table.Players[i].IsAllIn = false // Reiniciar estado de all-in
		// Reactivar todos los jugadores que tienen fichas (incluyendo los que llegaron durante la mano anterior)
		table.Players[i].IsActive = table.Players[i].Stack > 0 && table.Players[i].IsConnected
		
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
		player.IsAllIn = true // Marcar como all-in
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
	// Crear side pots al final de cada ronda de apuestas
	pe.createSidePots(table)
	
	// Resetear las apuestas para la nueva ronda (pero mantener side pots)
	for i := range table.Players {
		table.Players[i].CurrentBet = 0
		// Solo reactivar jugadores que no están en all-in
		if table.Players[i].IsActive && !table.Players[i].HasFolded && !table.Players[i].IsAllIn {
			table.PlayersToAct[i] = true
		} else {
			table.PlayersToAct[i] = false
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

	// Crear side pots si hay all-ins múltiples
	pe.createSidePots(table)

	// Distribuir side pots a los ganadores correspondientes
	pe.distributeSidePots(table)

	// Registrar tiempo de finalización del showdown
	table.ShowdownEndTime = time.Now()

	// Programar auto-restart si está habilitado y hay suficientes jugadores
	if table.AutoRestart && pe.hasEnoughActivePlayers(table) {
		// Usar goroutine para no bloquear el flujo actual
		go pe.scheduleAutoRestart(table.ID)
	}
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

// ====== SISTEMA DE SIDE POTS PARA ALL-INS MÚLTIPLES ======

// createSidePots crea los side pots basados en los all-ins y apuestas de los jugadores
func (pe *PokerEngine) createSidePots(table *PokerTable) {
	// Limpiar side pots existentes
	table.SidePots = make([]SidePot, 0)
	
	// Obtener jugadores que no han foldeado
	activePlayers := make([]int, 0)
	for i, player := range table.Players {
		if player.IsActive && !player.HasFolded {
			activePlayers = append(activePlayers, i)
		}
	}
	
	if len(activePlayers) == 0 {
		return
	}
	
	// Crear slice de niveles de apuesta únicos y ordenarlos
	betLevels := pe.getSortedBetLevels(table, activePlayers)
	
	if len(betLevels) == 0 {
		return
	}
	
	// Crear side pots para cada nivel
	for i, betLevel := range betLevels {
		sidePot := SidePot{
			Amount:          0,
			MaxBetLevel:     betLevel,
			EligiblePlayers: make([]int, 0),
		}
		
		// Calcular la cantidad y jugadores elegibles para este side pot
		prevLevel := 0
		if i > 0 {
			prevLevel = betLevels[i-1]
		}
		
		levelContribution := betLevel - prevLevel
		
		// Agregar jugadores elegibles y calcular cantidad
		for _, playerIndex := range activePlayers {
			player := table.Players[playerIndex]
			if player.CurrentBet >= betLevel {
				sidePot.EligiblePlayers = append(sidePot.EligiblePlayers, playerIndex)
				sidePot.Amount += levelContribution
			}
		}
		
		// Solo agregar el side pot si tiene participantes y cantidad
		if len(sidePot.EligiblePlayers) > 0 && sidePot.Amount > 0 {
			table.SidePots = append(table.SidePots, sidePot)
		}
	}
	
	// Actualizar pot principal para compatibilidad (suma de todos los side pots)
	table.Pot = pe.getTotalPot(table)
}

// getSortedBetLevels obtiene y ordena los niveles de apuesta únicos
func (pe *PokerEngine) getSortedBetLevels(table *PokerTable, activePlayers []int) []int {
	betLevelMap := make(map[int]bool)
	
	for _, playerIndex := range activePlayers {
		bet := table.Players[playerIndex].CurrentBet
		if bet > 0 {
			betLevelMap[bet] = true
		}
	}
	
	// Convertir a slice y ordenar
	betLevels := make([]int, 0, len(betLevelMap))
	for level := range betLevelMap {
		betLevels = append(betLevels, level)
	}
	
	// Ordenar de menor a mayor
	for i := 0; i < len(betLevels)-1; i++ {
		for j := i + 1; j < len(betLevels); j++ {
			if betLevels[i] > betLevels[j] {
				betLevels[i], betLevels[j] = betLevels[j], betLevels[i]
			}
		}
	}
	
	return betLevels
}

// getTotalPot calcula el pot total sumando todos los side pots
func (pe *PokerEngine) getTotalPot(table *PokerTable) int {
	total := 0
	for _, sidePot := range table.SidePots {
		total += sidePot.Amount
	}
	return total
}

// distributeSidePots distribuye los side pots a los ganadores correspondientes
func (pe *PokerEngine) distributeSidePots(table *PokerTable) {
	if len(table.SidePots) == 0 {
		pe.createSidePots(table)
	}
	
	// Evaluar manos para todos los jugadores activos
	playerHands := make(map[int]*HandEvaluation)
	for i, player := range table.Players {
		if player.IsActive && !player.HasFolded {
			// Evaluar mano usando las 2 cartas del jugador + 5 comunitarias
			if len(player.Cards) >= 2 && len(table.CommunityCards) >= 5 {
				handResult := EvaluateHand(player.Cards, table.CommunityCards)
				playerHands[i] = &handResult
			}
		}
	}
	
	// Distribuir cada side pot por separado
	for sidePotIndex, sidePot := range table.SidePots {
		if sidePot.Amount <= 0 || len(sidePot.EligiblePlayers) == 0 {
			continue
		}
		
		// Encontrar ganadores entre jugadores elegibles para este side pot
		winners := pe.findWinnersInSidePot(sidePot, playerHands)
		
		if len(winners) > 0 {
			// Dividir el side pot entre los ganadores
			potPerWinner := sidePot.Amount / len(winners)
			remainder := sidePot.Amount % len(winners)
			
			for i, winnerIndex := range winners {
				table.Players[winnerIndex].Stack += potPerWinner
				// Dar el resto al primer ganador
				if i == 0 {
					table.Players[winnerIndex].Stack += remainder
				}
			}
			
			// Marcar side pot como distribuido
			table.SidePots[sidePotIndex].Amount = 0
		}
	}
	
	// Actualizar pot principal
	table.Pot = 0
}

// findWinnersInSidePot encuentra los ganadores de un side pot específico
func (pe *PokerEngine) findWinnersInSidePot(sidePot SidePot, playerHands map[int]*HandEvaluation) []int {
	if len(sidePot.EligiblePlayers) == 0 {
		return []int{}
	}
	
	// Si solo hay un jugador elegible, es el ganador automático
	if len(sidePot.EligiblePlayers) == 1 {
		return sidePot.EligiblePlayers
	}
	
	// Encontrar la mejor mano entre los jugadores elegibles
	var bestHand *HandEvaluation
	winners := make([]int, 0)
	
	for _, playerIndex := range sidePot.EligiblePlayers {
		hand, exists := playerHands[playerIndex]
		if !exists {
			continue
		}
		
		if bestHand == nil {
			bestHand = hand
			winners = []int{playerIndex}
		} else {
			// Comparar por valor - mayor valor es mejor mano
			if hand.Value > bestHand.Value {
				// Nueva mejor mano
				bestHand = hand
				winners = []int{playerIndex}
			} else if hand.Value == bestHand.Value {
				// Empate - agregar a ganadores
				winners = append(winners, playerIndex)
			}
		}
	}
	
	return winners
}

// ====== SISTEMA DE AUTO-RESTART DE MANOS ======

// scheduleAutoRestart programa el reinicio automático de una mano después del delay configurado
func (pe *PokerEngine) scheduleAutoRestart(tableID string) {
	table, exists := pe.tables[tableID]
	if !exists {
		return
	}

	// Esperar el delay configurado
	time.Sleep(table.RestartDelay)

	// Verificar que la mesa aún esté en showdown y tenga jugadores suficientes
	table, exists = pe.tables[tableID]
	if !exists || table.Phase != "showdown" || !pe.hasEnoughActivePlayers(table) {
		return
	}

	// Verificar que no hayan pasado demasiado tiempo (evitar restart si los jugadores se han ido)
	if time.Since(table.ShowdownEndTime) > table.RestartDelay*2 {
		return
	}

	// Reiniciar la mano automáticamente
	pe.startHand(table)
}

// SetAutoRestart configura el auto-restart para una mesa
func (pe *PokerEngine) SetAutoRestart(tableID string, enabled bool, delay time.Duration) error {
	table, exists := pe.tables[tableID]
	if !exists {
		return fmt.Errorf("table not found")
	}

	table.AutoRestart = enabled
	if delay > 0 {
		table.RestartDelay = delay
	}

	return nil
}

// GetAutoRestartStatus obtiene el estado del auto-restart para una mesa
func (pe *PokerEngine) GetAutoRestartStatus(tableID string) (bool, time.Duration, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		return false, 0, fmt.Errorf("table not found")
	}

	return table.AutoRestart, table.RestartDelay, nil
}

// ForceRestartHand fuerza el reinicio de una mano (para testing o administración)
func (pe *PokerEngine) ForceRestartHand(tableID string) error {
	table, exists := pe.tables[tableID]
	if !exists {
		return fmt.Errorf("table not found")
	}

	if table.Phase != "showdown" {
		return fmt.Errorf("can only restart from showdown phase")
	}

	if !pe.hasEnoughActivePlayers(table) {
		return fmt.Errorf("not enough active players to restart")
	}

	pe.startHand(table)
	return nil
}

// ====== CONFIGURACIÓN DE BUY-IN ======

// GetTableConfig retorna la configuración de buy-in de una mesa
func (pe *PokerEngine) GetTableConfig(tableID string) (*TableConfig, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		return nil, fmt.Errorf("table not found")
	}

	config := &TableConfig{
		SmallBlind:   table.SmallBlind,
		BigBlind:     table.BigBlind,
		BuyInAmount:  table.BuyInAmount,
		MinBuyIn:     table.MinBuyIn,
		MaxBuyIn:     table.MaxBuyIn,
		IsCashGame:   table.IsCashGame,
		AutoRestart:  table.AutoRestart,
		RestartDelay: table.RestartDelay,
	}

	return config, nil
}

// ValidateBuyIn valida si un monto de buy-in es válido para una mesa
func (pe *PokerEngine) ValidateBuyIn(tableID string, buyInAmount int) error {
	table, exists := pe.tables[tableID]
	if !exists {
		return fmt.Errorf("table not found")
	}

	if buyInAmount < table.MinBuyIn {
		return fmt.Errorf("buy-in amount %d is below minimum %d", buyInAmount, table.MinBuyIn)
	}
	if buyInAmount > table.MaxBuyIn {
		return fmt.Errorf("buy-in amount %d is above maximum %d", buyInAmount, table.MaxBuyIn)
	}

	return nil
}

// UpdateTableConfig actualiza la configuración de buy-in de una mesa (solo para host)
func (pe *PokerEngine) UpdateTableConfig(tableID string, config TableConfig) error {
	table, exists := pe.tables[tableID]
	if !exists {
		return fmt.Errorf("table not found")
	}

	// Solo permitir cambios en lobby
	if table.Phase != "lobby" && table.Phase != "waiting" {
		return fmt.Errorf("can only update configuration in lobby or waiting phase")
	}

	// Validar que MinBuyIn <= BuyInAmount <= MaxBuyIn
	if config.MinBuyIn > config.BuyInAmount || config.BuyInAmount > config.MaxBuyIn {
		return fmt.Errorf("invalid buy-in configuration: MinBuyIn (%d) <= BuyInAmount (%d) <= MaxBuyIn (%d)", 
			config.MinBuyIn, config.BuyInAmount, config.MaxBuyIn)
	}

	// Actualizar configuración
	table.SmallBlind = config.SmallBlind
	table.BigBlind = config.BigBlind
	table.BuyInAmount = config.BuyInAmount
	table.MinBuyIn = config.MinBuyIn
	table.MaxBuyIn = config.MaxBuyIn
	table.IsCashGame = config.IsCashGame
	table.AutoRestart = config.AutoRestart
	table.RestartDelay = config.RestartDelay

	return nil
}

// ====== MANEJO BÁSICO DE DESCONEXIONES ======

// SetPlayerConnected actualiza el estado de conexión de un jugador
func (pe *PokerEngine) SetPlayerConnected(tableID, playerID string, connected bool) error {
	table, exists := pe.tables[tableID]
	if !exists {
		return fmt.Errorf("table not found")
	}

	// Encontrar jugador
	for i := range table.Players {
		if table.Players[i].ID == playerID {
			table.Players[i].IsConnected = connected
			table.Players[i].LastSeenTime = time.Now()
			
			// Si se desconectó durante el juego, puede afectar el flujo
			if !connected && table.Phase != "waiting" && table.Phase != "lobby" {
				// Fold automático si era su turno
				if table.CurrentPlayer == i && table.Phase != "showdown" {
					table.Players[i].HasFolded = true
					table.PlayersToAct[i] = false
					pe.nextPlayer(table)
				}
			}
			return nil
		}
	}

	return fmt.Errorf("player not found")
}

// GetDisconnectedPlayers obtiene jugadores desconectados por más de X tiempo
func (pe *PokerEngine) GetDisconnectedPlayers(tableID string, timeout time.Duration) ([]string, error) {
	table, exists := pe.tables[tableID]
	if !exists {
		return nil, fmt.Errorf("table not found")
	}

	var disconnected []string
	cutoff := time.Now().Add(-timeout)

	for _, player := range table.Players {
		if !player.IsConnected || player.LastSeenTime.Before(cutoff) {
			disconnected = append(disconnected, player.Name)
		}
	}

	return disconnected, nil
}

// HeartbeatPlayer actualiza el último momento visto de un jugador
func (pe *PokerEngine) HeartbeatPlayer(tableID, playerID string) error {
	table, exists := pe.tables[tableID]
	if !exists {
		return fmt.Errorf("table not found")
	}

	for i := range table.Players {
		if table.Players[i].ID == playerID {
			table.Players[i].LastSeenTime = time.Now()
			return nil
		}
	}

	return fmt.Errorf("player not found")
}
