// package game

// import (
// 	"fmt"
// 	"sync"
// )

// // Tipos de accion que puede tomar un jugador
// type ActionType string

// const (
// 	Fold  ActionType = "fold"
// 	Check ActionType = "check"
// 	Call  ActionType = "call"
// 	Bet   ActionType = "bet"
// 	Raise ActionType = "raise"
// 	AllIn ActionType = "allin"
// )

// // PlayerState guarda el estado de cada jugador en la ronda
// type PlayerState struct {
// 	Name         string `json:"name"`
// 	Stack        int    `json:"stack"`  // fichas restantes
// 	BetThisRound int    `json:"bet"`    // cuanto ha apostado en la ronda
// 	Active       bool   `json:"active"` // true si no ha foldeado
// }

// // TableState extiende el anterior para reflejar una ronda de apuestas
// type TableState struct {
// 	Host          string        `json:"host"`          // quien reparte
// 	Players       []PlayerState `json:"players"`       // estado de cada jugador
// 	Pot           int           `json:"pot"`           // bote acumulado
// 	CurrentBet    int           `json:"currentBet"`    // apuesta minima de la ronda
// 	DealerIndex   int           `json:"dealerIndex"`   // quien reparte
// 	TurnIndex     int           `json:"turnIndex"`     // indice del jugador actual
// 	RoundComplete bool          `json:"roundComplete"` // true si la ronda de apuestas ha terminado
// }

// // Engine coordina el flujo de cada ronda
// type Engine struct {
// 	mu     sync.Mutex
// 	tables map[string]*TableState
// }

// // NewEngine crea un motor limpio
// func NewEngine() *Engine {
// 	return &Engine{tables: make(map[string]*TableState)}
// }

// // InitTable prepara la mesa con blinds y orden inicial
// func (e *Engine) InitTable(tableID string, players []string, startingStack, smallBlind, bigBlind int) *TableState {
// 	e.mu.Lock()
// 	defer e.mu.Unlock()
// 	// Crear estados de jugadores
// 	states := make([]PlayerState, len(players))
// 	for i, n := range players {
// 		states[i] = PlayerState{Name: n, Stack: startingStack, BetThisRound: 0, Active: true}
// 	}
// 	// Deduce blinds
// 	states[1].Stack -= smallBlind
// 	states[1].BetThisRound = smallBlind
// 	states[2].Stack -= bigBlind
// 	states[2].BetThisRound = bigBlind
// 	pot := smallBlind + bigBlind
// 	ts := &TableState{
// 		Host:        players[0],
// 		Players:     states,
// 		Pot:         pot,
// 		CurrentBet:  bigBlind,
// 		DealerIndex: 0,
// 		TurnIndex:   3 % len(players), // primer actor tras big blind
// 	}
// 	e.tables[tableID] = ts
// 	return ts
// }

// // Act permite a un jugador ejecutar una accion en su turno
// func (e *Engine) Act(tableID, player string, action ActionType, amount int) (*TableState, error) {
// 	e.mu.Lock()
// 	defer e.mu.Unlock()
// 	ts, ok := e.tables[tableID]
// 	if !ok {
// 		return nil, fmt.Errorf("mesa %s no existe", tableID)
// 	}
// 	// Verificar turno
// 	if ts.Players[ts.TurnIndex].Name != player {
// 		return ts, fmt.Errorf("no es tu turno de %s", ts.Players[ts.TurnIndex].Name)
// 	}
// 	ps := &ts.Players[ts.TurnIndex]
// 	switch action {
// 	case Fold:
// 		ps.Active = false
// 	case Check:
// 		if ps.BetThisRound != ts.CurrentBet {
// 			return ts, fmt.Errorf("no puedes check, debes call o raise")
// 		}
// 	case Call:
// 		toCall := ts.CurrentBet - ps.BetThisRound
// 		if ps.Stack < toCall {
// 			return ts, fmt.Errorf("stack insuficiente para call")
// 		}
// 		ps.Stack -= toCall
// 		ps.BetThisRound += toCall
// 		ts.Pot += toCall
// 	case Bet, Raise:
// 		if amount <= ts.CurrentBet {
// 			return ts, fmt.Errorf("la apuesta debe ser mayor que currentBet")
// 		}
// 		if ps.Stack < amount {
// 			return ts, fmt.Errorf("stack insuficiente para bet/raise")
// 		}
// 		ps.Stack -= amount
// 		ps.BetThisRound += amount
// 		ts.Pot += amount
// 		ts.CurrentBet = amount
// 	case AllIn:
// 		amt := ps.Stack
// 		ps.Stack = 0
// 		ps.BetThisRound += amt
// 		ts.Pot += amt
// 		if amt > ts.CurrentBet {
// 			ts.CurrentBet = amt
// 		}
// 	default:
// 		return ts, fmt.Errorf("accion desconocida %s", action)

// 		// Avanzar al siguiente jugador activo
// 		n := len(ts.Players)
// 		for i := 1; i <= n; i++ {
// 			idx := (ts.TurnIndex + i) % n
// 			if ts.Players[idx].Active && ts.Players[idx].Stack > 0 {
// 				ts.TurnIndex = idx
// 				break
// 			}
// 		}
// 		// Si vuelta completa sin cambios, marcar ronda completa
// 		// (puedes refinar esta logica)
// 		// ...
// 	}
// 	return ts, nil
// }
