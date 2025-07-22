package poker

import (
	"testing"
)

// TestSidePotsBasic prueba el sistema básico de side pots
func TestSidePotsBasic(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("test_side_pots")

	// Agregar 3 jugadores con diferentes stacks
	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 100, IsActive: true, CurrentBet: 100, IsAllIn: true},  // 100 all-in
		{ID: "bob", Name: "Bob", Stack: 0, IsActive: true, CurrentBet: 500, IsAllIn: true},       // 500 all-in
		{ID: "carol", Name: "Carol", Stack: 500, IsActive: true, CurrentBet: 500, IsAllIn: false}, // 500 call
	}

	// Crear side pots
	engine.createSidePots(table)

	// Verificar que se crearon los side pots correctos
	if len(table.SidePots) == 0 {
		t.Errorf("Expected side pots to be created, got %d", len(table.SidePots))
		return
	}

	t.Logf("Side pots created: %d", len(table.SidePots))
	
	totalPot := 0
	for i, sidePot := range table.SidePots {
		t.Logf("Side pot %d: Amount=%d, MaxBetLevel=%d, EligiblePlayers=%v", 
			i+1, sidePot.Amount, sidePot.MaxBetLevel, sidePot.EligiblePlayers)
		totalPot += sidePot.Amount
	}

	expectedTotal := 100 + 500 + 500 // Alice + Bob + Carol
	if totalPot != expectedTotal {
		t.Errorf("Expected total pot %d, got %d", expectedTotal, totalPot)
	}

	// Verificar pot principal
	if table.Pot != expectedTotal {
		t.Errorf("Expected main pot %d, got %d", expectedTotal, table.Pot)
	}
}

// TestSidePotsDistribution prueba la distribución de side pots
func TestSidePotsDistribution(t *testing.T) {
	engine := NewPokerEngine()
	table := engine.CreateTable("test_distribution")

	// Agregar 3 jugadores
	table.Players = []PokerPlayer{
		{ID: "alice", Name: "Alice", Stack: 0, IsActive: true, CurrentBet: 100, IsAllIn: true, 
		 Cards: []Card{{Suit: "hearts", Rank: "A"}, {Suit: "spades", Rank: "K"}}},
		{ID: "bob", Name: "Bob", Stack: 0, IsActive: true, CurrentBet: 500, IsAllIn: true,
		 Cards: []Card{{Suit: "diamonds", Rank: "2"}, {Suit: "clubs", Rank: "3"}}},
		{ID: "carol", Name: "Carol", Stack: 500, IsActive: true, CurrentBet: 500, IsAllIn: false,
		 Cards: []Card{{Suit: "hearts", Rank: "Q"}, {Suit: "spades", Rank: "J"}}},
	}

	// Agregar cartas comunitarias
	table.CommunityCards = []Card{
		{Suit: "hearts", Rank: "10"},
		{Suit: "hearts", Rank: "9"},
		{Suit: "hearts", Rank: "8"},
		{Suit: "clubs", Rank: "7"},
		{Suit: "diamonds", Rank: "6"},
	}

	initialStacks := make([]int, len(table.Players))
	for i, player := range table.Players {
		initialStacks[i] = player.Stack
	}

	// Distribuir side pots
	engine.distributeSidePots(table)

	// Verificar que se distribuyeron las fichas
	totalDistributed := 0
	for i, player := range table.Players {
		gain := player.Stack - initialStacks[i]
		if gain > 0 {
			t.Logf("Player %s gained %d chips", player.Name, gain)
			totalDistributed += gain
		}
	}

	expectedDistribution := 1100 // Total de apuestas
	if totalDistributed != expectedDistribution {
		t.Errorf("Expected %d chips distributed, got %d", expectedDistribution, totalDistributed)
	}

	// Verificar que todos los side pots están vacíos
	for i, sidePot := range table.SidePots {
		if sidePot.Amount != 0 {
			t.Errorf("Side pot %d still has %d chips after distribution", i+1, sidePot.Amount)
		}
	}
}

// TestGetSortedBetLevels prueba la función de niveles de apuesta ordenados
func TestGetSortedBetLevels(t *testing.T) {
	engine := NewPokerEngine()
	table := &PokerTable{
		Players: []PokerPlayer{
			{CurrentBet: 100, IsActive: true, HasFolded: false},
			{CurrentBet: 500, IsActive: true, HasFolded: false},
			{CurrentBet: 300, IsActive: true, HasFolded: false},
			{CurrentBet: 100, IsActive: true, HasFolded: false}, // Duplicado
		},
	}

	activePlayers := []int{0, 1, 2, 3}
	betLevels := engine.getSortedBetLevels(table, activePlayers)

	expected := []int{100, 300, 500}
	if len(betLevels) != len(expected) {
		t.Errorf("Expected %d bet levels, got %d", len(expected), len(betLevels))
	}

	for i, level := range betLevels {
		if level != expected[i] {
			t.Errorf("Expected bet level %d at index %d, got %d", expected[i], i, level)
		}
	}
}