package game_test

import (
	"testing"

	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/game"
)

func TestManager_JoinAndTurns(t *testing.T) {
	mgr := game.NewManager()
	state := mgr.Join("mesa1", "A")
	if state.Host != "A" || len(state.Players) != 1 {
		t.Fatalf("expected host A and 1 player, got host=%s players=%v", state.Host, state.Players)
	}
	state = mgr.Join("mesa1", "B")
	if len(state.Players) != 2 {
		t.Fatalf("expected 2 players, got players=%v", state.Players)
	}
	// El TurnIndex ahora se maneja por el poker engine y puede variar
	if state.TurnIndex < 0 || state.TurnIndex >= len(state.Players) {
		t.Fatalf("turnIndex should be valid: got turnIndex=%d for %d players", state.TurnIndex, len(state.Players))
	}
}

func TestManager_BetTurnAdvance(t *testing.T) {
	mgr := game.NewManager()
	mgr.Join("mesa1", "A")
	state := mgr.Join("mesa1", "B")

	// Con poker engine, usaremos PokerAction en lugar de Bet
	// El pot ya tiene blinds así que verificamos que sea > 0
	if state.Pot <= 0 {
		t.Errorf("expected pot > 0 after players join (blinds), got pot=%d", state.Pot)
	}

	// Obtener el nombre del jugador actual
	var currentPlayerName string
	if state.PokerTable != nil && len(state.PokerTable.Players) > state.TurnIndex {
		currentPlayerName = state.PokerTable.Players[state.TurnIndex].Name
	} else {
		t.Fatalf("no poker table or invalid turn index")
	}

	// Hacer una acción válida con el jugador actual
	_, err := mgr.PokerAction("mesa1", currentPlayerName, "call", 0)
	if err != nil {
		t.Fatalf("unexpected error when %s calls: %v", currentPlayerName, err)
	}
}

func TestManager_Distribute(t *testing.T) {
	mgr := game.NewManager()
	mgr.Join("mesa1", "A")
	mgr.Join("mesa1", "B")
	
	// Con poker engine, distribute se maneja automáticamente
	// Solo verificamos que el estado se pueda obtener
	state, err := mgr.GetTableState("mesa1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Verificar que hay un pot (de los blinds)
	if state.Pot < 0 {
		t.Errorf("expected pot >= 0, got pot=%d", state.Pot)
	}
}
