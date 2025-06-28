package game_test

import (
	"testing"

	"github.com/zkCaleb-dev/Poker-Off-Chain/internal/game"
)

func TestManager_JoinAndTurns(t *testing.T) {
	mgr := game.NewManager()
	state := mgr.Join("mesa1", "A")
	if state.Host != "A" || len(state.Players) != 1 {
		t.Fatalf("expected host A and 1 player, got host=%s players=%v", state.Host, state.Players)
	}
	state = mgr.Join("mesa1", "B")
	if len(state.Players) != 2 || state.TurnIndex != 0 {
		t.Fatalf("expected 2 players and turnIndex 0, got players=%v, turnIndex=%d", state.Players, state.TurnIndex)
	}
}

func TestManager_BetTurnAdvance(t *testing.T) {
	mgr := game.NewManager()
	mgr.Join("mesa1", "A")
	mgr.Join("mesa1", "B")

	// A apuesta correctamente
	state, err := mgr.Bet("mesa1", "A", 10)
	if err != nil {
		t.Fatalf("unexpected error when A apuesta: %v", err)
	}
	if state.Pot != 10 || state.TurnIndex != 1 {
		t.Errorf("expected pot=10 turnIndex=1, got pot=%d turnIndex=%d", state.Pot, state.TurnIndex)
	}

	// B apuesta correctamente
	state, err = mgr.Bet("mesa1", "B", 20)
	if err != nil {
		t.Fatalf("unexpected error when B apuesta: %v", err)
	}
	if state.Pot != 30 || state.TurnIndex != 0 {
		t.Errorf("expected pot=30 turnIndex=0, got pot=%d turnIndex=%d", state.Pot, state.TurnIndex)
	}

	// Out-of-turn: B intenta apostar de nuevo sin esperar turno de A
	_, err = mgr.Bet("mesa1", "B", 5)
	if err == nil {
		t.Errorf("expected error when B apuesta fuera de turno")
	}
}

func TestManager_Distribute(t *testing.T) {
	mgr := game.NewManager()
	mgr.Join("mesa1", "A")
	mgr.Bet("mesa1", "A", 50)
	state, err := mgr.Distribute("mesa1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Pot != 0 || state.TurnIndex != 0 {
		t.Errorf("expected pot=0 turnIndex=0 after distribute, got pot=%d turnIndex=%d", state.Pot, state.TurnIndex)
	}
}
