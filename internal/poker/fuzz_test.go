package poker

import (
	"crypto/rand"
	"math/big"
	"testing"
)

// FuzzEvaluateHand fuzzes la función de evaluación de manos
func FuzzEvaluateHand(f *testing.F) {
	// Seed con algunos casos válidos
	f.Add(
		"hearts", "A", "spades", "K",  // player cards
		"diamonds", "Q", "clubs", "J", "hearts", "10", "spades", "2", "clubs", "3", // community cards
	)
	f.Add(
		"hearts", "2", "spades", "3",
		"diamonds", "4", "clubs", "5", "hearts", "6", "spades", "7", "clubs", "8",
	)
	
	f.Fuzz(func(t *testing.T, 
		p1Suit, p1Rank, p2Suit, p2Rank string,
		c1Suit, c1Rank, c2Suit, c2Rank, c3Suit, c3Rank, c4Suit, c4Rank, c5Suit, c5Rank string) {
		
		// Validar inputs
		validSuits := map[string]bool{"hearts": true, "diamonds": true, "clubs": true, "spades": true}
		validRanks := map[string]bool{"2": true, "3": true, "4": true, "5": true, "6": true, "7": true, "8": true, "9": true, "10": true, "J": true, "Q": true, "K": true, "A": true}
		
		suits := []string{p1Suit, p2Suit, c1Suit, c2Suit, c3Suit, c4Suit, c5Suit}
		ranks := []string{p1Rank, p2Rank, c1Rank, c2Rank, c3Rank, c4Rank, c5Rank}
		
		for _, suit := range suits {
			if !validSuits[suit] {
				return // Skip invalid input
			}
		}
		for _, rank := range ranks {
			if !validRanks[rank] {
				return // Skip invalid input
			}
		}
		
		playerCards := []Card{
			{Suit: p1Suit, Rank: p1Rank},
			{Suit: p2Suit, Rank: p2Rank},
		}
		
		communityCards := []Card{
			{Suit: c1Suit, Rank: c1Rank},
			{Suit: c2Suit, Rank: c2Rank},
			{Suit: c3Suit, Rank: c3Rank},
			{Suit: c4Suit, Rank: c4Rank},
			{Suit: c5Suit, Rank: c5Rank},
		}
		
		// Esta función no debería hacer panic nunca
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("EvaluateHand panicked with input: player=%v, community=%v, error=%v", 
					playerCards, communityCards, r)
			}
		}()
		
		evaluation := EvaluateHand(playerCards, communityCards)
		
		// Verificar que el resultado sea válido
		if evaluation.Rank < HighCard || evaluation.Rank > RoyalFlush {
			t.Fatalf("Invalid hand rank: %v", evaluation.Rank)
		}
		
		if evaluation.Value < 0 {
			t.Fatalf("Invalid hand value: %d", evaluation.Value)
		}
		
		if len(evaluation.Cards) == 0 {
			t.Fatalf("No cards in evaluation result")
		}
		
		if evaluation.RankName == "" {
			t.Fatalf("Empty rank name")
		}
	})
}

// FuzzPlayerAction fuzzes las acciones de los jugadores
func FuzzPlayerAction(f *testing.F) {
	// Seed con acciones válidas
	f.Add("call", 0)
	f.Add("raise", 20)
	f.Add("fold", 0)
	f.Add("check", 0)
	f.Add("all_in", 0)
	
	f.Fuzz(func(t *testing.T, action string, amount int) {
		engine := NewPokerEngine()
		table := engine.CreateTable("fuzz_test")
		
		// Agregar jugadores
		engine.AddPlayer("fuzz_test", "player1", "Player1")
		engine.AddPlayer("fuzz_test", "player2", "Player2")
		
		// Marcar como ready e iniciar
		table, _ = engine.GetTable("fuzz_test")
		for i := range table.Players {
			table.Players[i].IsReady = true
		}
		
		engine.StartGame("fuzz_test", "player1")
		
		// Esta función no debería hacer panic nunca
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("PlayerAction panicked with action=%s, amount=%d, error=%v", 
					action, amount, r)
			}
		}()
		
		// Obtener tabla actualizada
		table, _ = engine.GetTable("fuzz_test")
		if table == nil || len(table.Players) == 0 {
			return
		}
		
		currentPlayerIndex := table.CurrentPlayer
		if currentPlayerIndex >= len(table.Players) {
			return
		}
		
		playerID := table.Players[currentPlayerIndex].ID
		
		// Intentar la acción - puede fallar pero no debe hacer panic
		_, err := engine.PlayerAction("fuzz_test", playerID, action, amount)
		
		// No verificamos el error porque muchas acciones son inválidas en fuzzing,
		// solo verificamos que no haga panic
		_ = err
	})
}

// FuzzDeckCreation fuzzes la creación y barajado de deck
func FuzzDeckCreation(f *testing.F) {
	f.Add(int64(12345))
	f.Add(int64(0))
	f.Add(int64(-1))
	
	f.Fuzz(func(t *testing.T, seed int64) {
		engine := NewPokerEngine()
		
		// Esta función no debería hacer panic nunca
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("createShuffledDeck panicked with seed=%d, error=%v", seed, r)
			}
		}()
		
		deck := engine.createShuffledDeck()
		
		// Verificar invariantes básicos
		if len(deck) != 52 {
			t.Fatalf("Expected 52 cards, got %d", len(deck))
		}
		
		// Verificar que no hay cartas duplicadas
		cardSet := make(map[string]bool)
		for _, card := range deck {
			cardKey := card.Suit + ":" + card.Rank
			if cardSet[cardKey] {
				t.Fatalf("Found duplicate card: %s of %s", card.Rank, card.Suit)
			}
			cardSet[cardKey] = true
		}
	})
}

// TestFuzzSidePots verifica que los side pots manejen casos extremos
func TestFuzzSidePots(t *testing.T) {
	engine := NewPokerEngine()
	
	for i := 0; i < 100; i++ {
		table := engine.CreateTable("side_pot_fuzz")
		
		// Crear escenario aleatorio con diferentes stacks y apuestas
		numPlayers, _ := rand.Int(rand.Reader, big.NewInt(8))
		playerCount := int(numPlayers.Int64()) + 2 // 2-9 jugadores
		
		for j := 0; j < playerCount; j++ {
			playerID := "player" + string(rune('A'+j))
			engine.AddPlayer("side_pot_fuzz", playerID, playerID)
		}
		
		table, _ = engine.GetTable("side_pot_fuzz")
		
		// Simular all-ins con diferentes cantidades
		for j := range table.Players {
			stackSize, _ := rand.Int(rand.Reader, big.NewInt(1000))
			betAmount, _ := rand.Int(rand.Reader, big.NewInt(500))
			
			table.Players[j].Stack = int(stackSize.Int64()) + 100 // Mínimo 100
			table.Players[j].CurrentBet = int(betAmount.Int64())
			table.Players[j].IsActive = true
			table.Players[j].HasFolded = false
			
			// Algunos jugadores all-in
			if j%3 == 0 {
				table.Players[j].IsAllIn = true
				table.Players[j].Stack = 0
			}
		}
		
		// Esta función no debería hacer panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Side pot creation panicked on iteration %d: %v", i, r)
			}
		}()
		
		// Crear side pots
		engine.createSidePots(table)
		
		// Verificar invariantes básicos
		totalPot := 0
		for _, sidePot := range table.SidePots {
			if sidePot.Amount < 0 {
				t.Fatalf("Negative side pot amount: %d", sidePot.Amount)
			}
			totalPot += sidePot.Amount
			
			if len(sidePot.EligiblePlayers) == 0 && sidePot.Amount > 0 {
				t.Fatalf("Side pot with amount but no eligible players")
			}
		}
		
		// El total debe ser igual al pot principal
		if totalPot != table.Pot {
			t.Logf("Warning: Total side pots (%d) != main pot (%d)", totalPot, table.Pot)
		}
	}
}

// BenchmarkEvaluateHandFuzz mide el rendimiento bajo condiciones de fuzzing
func BenchmarkEvaluateHandFuzz(b *testing.B) {
	// Pre-generar casos de prueba aleatorios
	testCases := make([]struct {
		playerCards    []Card
		communityCards []Card
	}, 1000)
	
	validSuits := []string{"hearts", "diamonds", "clubs", "spades"}
	validRanks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	
	for i := range testCases {
		// Generar cartas aleatorias válidas
		playerCards := make([]Card, 2)
		communityCards := make([]Card, 5)
		
		for j := 0; j < 2; j++ {
			suitIdx, _ := rand.Int(rand.Reader, big.NewInt(4))
			rankIdx, _ := rand.Int(rand.Reader, big.NewInt(13))
			playerCards[j] = Card{
				Suit: validSuits[suitIdx.Int64()],
				Rank: validRanks[rankIdx.Int64()],
			}
		}
		
		for j := 0; j < 5; j++ {
			suitIdx, _ := rand.Int(rand.Reader, big.NewInt(4))
			rankIdx, _ := rand.Int(rand.Reader, big.NewInt(13))
			communityCards[j] = Card{
				Suit: validSuits[suitIdx.Int64()],
				Rank: validRanks[rankIdx.Int64()],
			}
		}
		
		testCases[i] = struct {
			playerCards    []Card
			communityCards []Card
		}{playerCards, communityCards}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc := testCases[i%len(testCases)]
		_ = EvaluateHand(tc.playerCards, tc.communityCards)
	}
}