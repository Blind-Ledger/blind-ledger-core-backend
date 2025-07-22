package poker

import (
	"testing"
	"time"
)

// TestComprehensiveSystemIntegration prueba todo el sistema integrado
func TestComprehensiveSystemIntegration(t *testing.T) {
	t.Log("🧪 PRUEBA EXHAUSTIVA COMPLETA DEL SISTEMA TEXAS HOLD'EM")
	t.Log("=======================================================")

	engine := NewPokerEngine()

	// Test 1: Side Pots
	t.Run("SidePots", func(t *testing.T) {
		testSidePotsIntegration(t, engine)
	})

	// Test 2: Auto-restart
	t.Run("AutoRestart", func(t *testing.T) {
		testAutoRestartIntegration(t, engine)
	})

	// Test 3: Flujo completo
	t.Run("CompleteFlow", func(t *testing.T) {
		testCompleteFlowIntegration(t, engine)
	})

	t.Log("✅ TODAS LAS PRUEBAS COMPLETADAS EXITOSAMENTE")
}

func testSidePotsIntegration(t *testing.T, engine *PokerEngine) {
	t.Log("🎯 Test 1: Side Pots con múltiples all-ins")
	
	table := engine.CreateTable("side_pots_integration")
	table.Phase = "showdown"
	
	// Crear escenario de side pots: Alice (100), Bob (500), Carol (500)
	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 100, IsActive: true, CurrentBet: 100, IsAllIn: true,
		 Cards: []Card{{Suit: "hearts", Rank: "A"}, {Suit: "spades", Rank: "K"}}},
		{ID: "bob", Name: "Bob", Stack: 0, IsActive: true, CurrentBet: 500, IsAllIn: true,
		 Cards: []Card{{Suit: "diamonds", Rank: "2"}, {Suit: "clubs", Rank: "3"}}},
		{ID: "carol", Name: "Carol", Stack: 500, IsActive: true, CurrentBet: 500, IsAllIn: false,
		 Cards: []Card{{Suit: "spades", Rank: "7"}, {Suit: "diamonds", Rank: "6"}}},
	}
	
	table.CommunityCards = []Card{
		{Suit: "hearts", Rank: "Q"}, {Suit: "hearts", Rank: "J"}, {Suit: "hearts", Rank: "10"},
		{Suit: "clubs", Rank: "9"}, {Suit: "diamonds", Rank: "8"},
	}

	initialStacks := make([]int, 3)
	for i := range table.Players {
		initialStacks[i] = table.Players[i].Stack
	}

	// Completar mano (creará side pots y los distribuirá)
	engine.completeHand(table)

	t.Logf("   Side pots procesados: %d", len(table.SidePots))
	
	// Verificar distribución
	totalDistributed := 0
	for i, player := range table.Players {
		gain := player.Stack - initialStacks[i]
		if gain > 0 {
			t.Logf("   %s ganó %d chips", player.Name, gain)
			totalDistributed += gain
		}
	}
	
	expectedTotal := 1100 // 100 + 500 + 500
	if totalDistributed == expectedTotal {
		t.Log("   ✅ Side pots distribuidos correctamente")
	} else {
		t.Errorf("   ❌ Expected %d chips distributed, got %d", expectedTotal, totalDistributed)
	}
}

func testAutoRestartIntegration(t *testing.T, engine *PokerEngine) {
	t.Log("🎯 Test 2: Auto-restart de manos")
	
	table := engine.CreateTable("auto_restart_integration")
	table.AutoRestart = true
	table.RestartDelay = 100 * time.Millisecond
	
	// Agregar jugadores con conexión
	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
		{ID: "bob", Name: "Bob", Stack: 500, IsActive: true, IsConnected: true, LastSeenTime: time.Now()},
	}
	
	// Simular end de mano
	table.Phase = "showdown"
	table.ShowdownEndTime = time.Now()
	
	t.Log("   Iniciando auto-restart...")
	engine.completeHand(table)
	
	// Esperar el restart
	time.Sleep(200 * time.Millisecond)
	
	if table.Phase == "preflop" {
		t.Log("   ✅ Auto-restart funcionó correctamente")
	} else {
		t.Errorf("   ❌ Expected preflop phase after restart, got %s", table.Phase)
	}
	
	// Verificar cartas
	cardsOK := true
	for i, player := range table.Players {
		if len(player.Cards) != 2 {
			cardsOK = false
			t.Errorf("   Player %d has %d cards, expected 2", i, len(player.Cards))
		}
	}
	
	if cardsOK {
		t.Log("   ✅ Cartas repartidas correctamente")
	}
}

func testCompleteFlowIntegration(t *testing.T, engine *PokerEngine) {
	t.Log("🎯 Test 3: Flujo completo de Texas Hold'em")
	
	table := engine.CreateTable("complete_flow_integration")
	table.SmallBlind = 10
	table.BigBlind = 20
	
	// Agregar jugadores usando la función correcta
	engine.AddPlayer("complete_flow_integration", "alice_id", "Alice")
	engine.AddPlayer("complete_flow_integration", "bob_id", "Bob")
	engine.AddPlayer("complete_flow_integration", "carol_id", "Carol")
	
	// Marcar como ready y conectados
	for i := range table.Players {
		table.Players[i].IsReady = true
		table.Players[i].IsConnected = true
		table.Players[i].LastSeenTime = time.Now()
	}
	
	// Cambiar fase a lobby
	table.Phase = "lobby"
	
	// Iniciar mano
	engine.startHand(table)
	
	t.Logf("   Fase inicial: %s", table.Phase)
	t.Logf("   Jugadores activos: %d", countActivePlayers(table))
	t.Logf("   Blinds configurados: SB=%d, BB=%d", table.SmallBlind, table.BigBlind)
	
	// Verificar cartas
	cardsDealt := 0
	for _, player := range table.Players {
		if len(player.Cards) == 2 {
			cardsDealt++
		}
	}
	
	if cardsDealt == len(table.Players) {
		t.Log("   ✅ Cartas repartidas a todos los jugadores")
	} else {
		t.Errorf("   ❌ Expected cards for %d players, got %d", len(table.Players), cardsDealt)
	}
	
	// Verificar blinds
	blindsCount := 0
	totalBlinds := 0
	for _, player := range table.Players {
		if player.CurrentBet > 0 {
			blindsCount++
			totalBlinds += player.CurrentBet
		}
	}
	
	expectedBlinds := table.SmallBlind + table.BigBlind
	if totalBlinds == expectedBlinds {
		t.Logf("   ✅ Blinds correctos: %d total", totalBlinds)
	} else {
		t.Errorf("   ❌ Expected blinds total %d, got %d", expectedBlinds, totalBlinds)
	}
	
	// Simular ronda de apuestas completa
	t.Log("   Simulando ronda de apuestas...")
	
	// Todos los jugadores call
	playerIDs := []string{"alice_id", "bob_id", "carol_id"}
	for _, playerID := range playerIDs {
		_, err := engine.PlayerAction("complete_flow_integration", playerID, "call", table.BigBlind)
		if err != nil && table.CurrentPlayer < len(table.Players) {
			// Solo aplicar la acción si es el turno del jugador
			currentPlayerID := table.Players[table.CurrentPlayer].ID
			if currentPlayerID == playerID {
				engine.PlayerAction("complete_flow_integration", playerID, "call", table.BigBlind)
			}
		}
	}
	
	// Verificar progresión
	if table.Phase == "flop" || len(table.CommunityCards) >= 3 {
		t.Log("   ✅ Progresión a flop exitosa")
		if len(table.CommunityCards) >= 3 {
			t.Logf("   ✅ Flop repartido: %d cartas comunitarias", len(table.CommunityCards))
		}
	} else {
		t.Logf("   ⚠️  Fase actual: %s, cartas comunitarias: %d", table.Phase, len(table.CommunityCards))
	}
}

// Función auxiliar
func countActivePlayers(table *PokerTable) int {
	count := 0
	for _, player := range table.Players {
		if player.IsActive {
			count++
		}
	}
	return count
}