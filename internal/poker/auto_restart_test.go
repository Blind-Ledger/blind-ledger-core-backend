package poker

import (
	"testing"
	"time"
)

// TestAutoRestartBasic prueba el funcionamiento básico del auto-restart
func TestAutoRestartBasic(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("test_auto_restart")

	// Configurar auto-restart con delay corto para testing
	table.AutoRestart = true
	table.RestartDelay = 100 * time.Millisecond

	// Agregar jugadores suficientes con conexión
	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
		{ID: "bob", Name: "Bob", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
	}

	// Simular final de mano
	table.Phase = "showdown"
	table.ShowdownEndTime = time.Now()

	// Completar la mano (debería programar auto-restart)
	engine.completeHand(table)

	// Esperar más que el delay configurado
	time.Sleep(200 * time.Millisecond)

	// Verificar que la mano se reinició
	if table.Phase == "showdown" {
		t.Errorf("Expected hand to auto-restart, but still in showdown phase")
	}

	if table.Phase != "preflop" {
		t.Errorf("Expected phase to be preflop after restart, got %s", table.Phase)
	}

	// Verificar que los jugadores fueron reactivados correctamente
	for i, player := range table.Players {
		if !player.IsActive {
			t.Errorf("Player %d should be active after restart", i)
		}
		if player.HasFolded {
			t.Errorf("Player %d should not be folded after restart", i)
		}
		if player.IsAllIn {
			t.Errorf("Player %d should not be all-in after restart", i)
		}
	}
}

// TestAutoRestartDisabled prueba que no se reinicia cuando está deshabilitado
func TestAutoRestartDisabled(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("test_disabled")

	// Deshabilitar auto-restart
	table.AutoRestart = false
	table.RestartDelay = 100 * time.Millisecond

	// Agregar jugadores suficientes con conexión
	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
		{ID: "bob", Name: "Bob", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
	}

	// Simular final de mano
	table.Phase = "showdown"
	engine.completeHand(table)

	// Esperar más que el delay
	time.Sleep(200 * time.Millisecond)

	// Verificar que NO se reinició
	if table.Phase != "showdown" {
		t.Errorf("Expected to stay in showdown when auto-restart disabled, got %s", table.Phase)
	}
}

// TestAutoRestartInsufficientPlayers prueba que no se reinicia sin jugadores suficientes
func TestAutoRestartInsufficientPlayers(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("test_insufficient")

	table.AutoRestart = true
	table.RestartDelay = 100 * time.Millisecond

	// Solo un jugador activo (insuficiente)
	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
		{ID: "bob", Name: "Bob", Stack: 0, IsActive: false, IsConnected: false}, // Sin fichas
	}

	table.Phase = "showdown"
	engine.completeHand(table)

	time.Sleep(200 * time.Millisecond)

	// Verificar que NO se reinició por jugadores insuficientes
	if table.Phase != "showdown" {
		t.Errorf("Expected to stay in showdown with insufficient players, got %s", table.Phase)
	}
}

// TestSetAutoRestart prueba la configuración de auto-restart
func TestSetAutoRestart(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("test_config")

	// Verificar configuración inicial
	enabled, delay, err := engine.GetAutoRestartStatus(table.ID)
	if err != nil {
		t.Fatalf("Error getting auto-restart status: %v", err)
	}

	if !enabled {
		t.Errorf("Expected auto-restart to be enabled by default")
	}

	if delay != 5*time.Second {
		t.Errorf("Expected default delay to be 5s, got %v", delay)
	}

	// Cambiar configuración
	newDelay := 10 * time.Second
	err = engine.SetAutoRestart(table.ID, false, newDelay)
	if err != nil {
		t.Fatalf("Error setting auto-restart: %v", err)
	}

	// Verificar cambios
	enabled, delay, err = engine.GetAutoRestartStatus(table.ID)
	if err != nil {
		t.Fatalf("Error getting updated auto-restart status: %v", err)
	}

	if enabled {
		t.Errorf("Expected auto-restart to be disabled")
	}

	if delay != newDelay {
		t.Errorf("Expected delay to be %v, got %v", newDelay, delay)
	}
}

// TestForceRestartHand prueba el reinicio forzado de manos
func TestForceRestartHand(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("test_force")

	// Agregar jugadores
	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
		{ID: "bob", Name: "Bob", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
	}

	// Poner en showdown
	table.Phase = "showdown"

	// Forzar reinicio
	err := engine.ForceRestartHand(table.ID)
	if err != nil {
		t.Fatalf("Error forcing restart: %v", err)
	}

	// Verificar que se reinició
	if table.Phase != "preflop" {
		t.Errorf("Expected phase to be preflop after force restart, got %s", table.Phase)
	}
}

// TestForceRestartInvalidPhase prueba que no se puede forzar restart desde fases incorrectas
func TestForceRestartInvalidPhase(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("test_invalid")

	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
		{ID: "bob", Name: "Bob", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
	}

	// Intentar desde fase incorrecta
	table.Phase = "preflop"

	err := engine.ForceRestartHand(table.ID)
	if err == nil {
		t.Errorf("Expected error when trying to restart from non-showdown phase")
	}

	expectedError := "can only restart from showdown phase"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestAutoRestartWithSidePots prueba que el auto-restart funciona correctamente después de side pots
func TestAutoRestartWithSidePots(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("test_sidepots_restart")

	table.AutoRestart = true
	table.RestartDelay = 100 * time.Millisecond

	// Configurar escenario con side pots - todos los jugadores deben tener fichas para el restart
	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 600, IsActive: true, CurrentBet: 100, IsAllIn: true, IsConnected: true, LastSeenTime: time.Now(),
		 Cards: []Card{{Suit: "hearts", Rank: "A"}, {Suit: "spades", Rank: "K"}}},
		{ID: "bob", Name: "Bob", Stack: 400, IsActive: true, CurrentBet: 500, IsAllIn: false, IsConnected: true, LastSeenTime: time.Now(),
		 Cards: []Card{{Suit: "diamonds", Rank: "2"}, {Suit: "clubs", Rank: "3"}}},
		{ID: "carol", Name: "Carol", Stack: 500, IsActive: true, CurrentBet: 500, IsAllIn: false, IsConnected: true, LastSeenTime: time.Now(),
		 Cards: []Card{{Suit: "hearts", Rank: "Q"}, {Suit: "spades", Rank: "J"}}},
	}

	table.CommunityCards = []Card{
		{Suit: "hearts", Rank: "10"}, {Suit: "hearts", Rank: "9"}, {Suit: "hearts", Rank: "8"},
		{Suit: "clubs", Rank: "7"}, {Suit: "diamonds", Rank: "6"},
	}

	table.Phase = "showdown"
	table.ShowdownEndTime = time.Now()

	// Completar mano con side pots
	engine.completeHand(table)

	// Esperar auto-restart
	time.Sleep(200 * time.Millisecond)

	// Verificar que se reinició correctamente
	if table.Phase != "preflop" {
		t.Errorf("Expected preflop phase after restart, got %s", table.Phase)
	}

	// Verificar que los side pots se limpiaron
	if len(table.SidePots) != 0 {
		t.Errorf("Expected side pots to be cleared after restart, got %d", len(table.SidePots))
	}

	// Verificar que los estados de jugador se reiniciaron correctamente
	for i, player := range table.Players {
		if player.IsAllIn {
			t.Errorf("Player %d should not be all-in after restart", i)
		}
		if len(player.Cards) != 2 {
			t.Errorf("Player %d should have 2 cards after restart, got %d", i, len(player.Cards))
		}
		// Después del restart, los blinds se colocan automáticamente, así que algunos jugadores tendrán CurrentBet > 0
		// Esto es comportamiento correcto
	}

	// Verificar que al menos un jugador tiene blind
	foundBlind := false
	for _, player := range table.Players {
		if player.CurrentBet > 0 {
			foundBlind = true
			break
		}
	}
	if !foundBlind {
		t.Errorf("Expected at least one player to have posted blinds after restart")
	}
}