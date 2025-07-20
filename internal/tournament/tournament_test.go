package tournament_test

import (
	"testing"

	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/poker"
	"github.com/Blind-Ledger/blind-ledger-core-backend/internal/tournament"
)

func TestTournament_CreateAndRegister(t *testing.T) {
	pokerEngine := poker.NewPokerEngine()
	manager := tournament.NewManager(pokerEngine)

	// Crear torneo estándar
	tournament, err := manager.CreateStandardTournament("test1", "Test Tournament", 100)
	if err != nil {
		t.Fatalf("failed to create tournament: %v", err)
	}

	if tournament.ID != "test1" {
		t.Errorf("expected tournament ID test1, got %s", tournament.ID)
	}

	if tournament.Config.BuyIn != 100 {
		t.Errorf("expected buy-in 100, got %d", tournament.Config.BuyIn)
	}

	// Registrar jugadores
	err = tournament.RegisterPlayer("player1", "Alice")
	if err != nil {
		t.Fatalf("failed to register player1: %v", err)
	}

	err = tournament.RegisterPlayer("player2", "Bob")
	if err != nil {
		t.Fatalf("failed to register player2: %v", err)
	}

	// Verificar estado
	if tournament.GetPlayerCount() != 2 {
		t.Errorf("expected 2 players, got %d", tournament.GetPlayerCount())
	}

	if tournament.PrizePool != 200 {
		t.Errorf("expected prize pool 200, got %d", tournament.PrizePool)
	}
}

func TestTournament_Registration(t *testing.T) {
	pokerEngine := poker.NewPokerEngine()
	manager := tournament.NewManager(pokerEngine)

	tournament, err := manager.CreateStandardTournament("test2", "Test Tournament 2", 50)
	if err != nil {
		t.Fatalf("failed to create tournament: %v", err)
	}

	// Registrar mismo jugador dos veces
	err = tournament.RegisterPlayer("player1", "Alice")
	if err != nil {
		t.Fatalf("failed to register player1: %v", err)
	}

	err = tournament.RegisterPlayer("player1", "Alice")
	if err == nil {
		t.Errorf("expected error when registering same player twice")
	}

	// Desregistrar jugador
	err = tournament.UnregisterPlayer("player1")
	if err != nil {
		t.Fatalf("failed to unregister player: %v", err)
	}

	if tournament.GetPlayerCount() != 0 {
		t.Errorf("expected 0 players after unregister, got %d", tournament.GetPlayerCount())
	}
}

func TestTournament_BlindLevels(t *testing.T) {
	pokerEngine := poker.NewPokerEngine()
	manager := tournament.NewManager(pokerEngine)

	tournament, err := manager.CreateStandardTournament("test3", "Test Tournament 3", 100)
	if err != nil {
		t.Fatalf("failed to create tournament: %v", err)
	}

	// Verificar que hay niveles de blinds
	if len(tournament.Config.BlindLevels) == 0 {
		t.Errorf("expected blind levels to be configured")
	}

	level := tournament.GetCurrentBlindLevel()
	if level.SmallBlind != 10 || level.BigBlind != 20 {
		t.Errorf("expected first level SB=10 BB=20, got SB=%d BB=%d", 
			level.SmallBlind, level.BigBlind)
	}
}

func TestTournament_TurboVsStandard(t *testing.T) {
	pokerEngine := poker.NewPokerEngine()
	manager := tournament.NewManager(pokerEngine)

	// Crear torneo estándar
	standard, err := manager.CreateStandardTournament("standard", "Standard", 100)
	if err != nil {
		t.Fatalf("failed to create standard tournament: %v", err)
	}

	// Crear torneo turbo
	turbo, err := manager.CreateTurboTournament("turbo", "Turbo", 100)
	if err != nil {
		t.Fatalf("failed to create turbo tournament: %v", err)
	}

	// Verificar que los tiempos son diferentes
	standardDuration := standard.Config.BlindLevels[0].Duration
	turboDuration := turbo.Config.BlindLevels[0].Duration

	if standardDuration == turboDuration {
		t.Errorf("expected different blind durations for standard vs turbo")
	}

	if turboDuration >= standardDuration {
		t.Errorf("expected turbo to have shorter blind levels than standard")
	}

	// Verificar nombres
	if turbo.Config.Name != "Turbo (Turbo)" {
		t.Errorf("expected turbo name to include (Turbo), got %s", turbo.Config.Name)
	}
}

func TestManager_ListTournaments(t *testing.T) {
	pokerEngine := poker.NewPokerEngine()
	manager := tournament.NewManager(pokerEngine)

	// Crear varios torneos
	_, err := manager.CreateStandardTournament("t1", "Tournament 1", 100)
	if err != nil {
		t.Fatalf("failed to create tournament 1: %v", err)
	}

	_, err = manager.CreateTurboTournament("t2", "Tournament 2", 200)
	if err != nil {
		t.Fatalf("failed to create tournament 2: %v", err)
	}

	// Listar todos
	tournaments := manager.ListTournaments()
	if len(tournaments) != 2 {
		t.Errorf("expected 2 tournaments, got %d", len(tournaments))
	}

	// Verificar que ambos están presentes
	if _, exists := tournaments["t1"]; !exists {
		t.Errorf("tournament t1 not found in list")
	}

	if _, exists := tournaments["t2"]; !exists {
		t.Errorf("tournament t2 not found in list")
	}
}