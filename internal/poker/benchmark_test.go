package poker

import (
	"strconv"
	"testing"
)

// BenchmarkEngineOperations benchmarks operaciones principales del engine
func BenchmarkEngineOperations(b *testing.B) {
	engine := NewPokerEngine()
	
	b.Run("CreateTable", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tableID := "bench_table_" + strconv.Itoa(i)
			engine.CreateTable(tableID)
		}
	})
	
	b.Run("AddPlayer", func(b *testing.B) {
		table := engine.CreateTable("bench_add_player")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			playerID := "player_" + strconv.Itoa(i)
			engine.AddPlayer("bench_add_player", playerID, "Player"+strconv.Itoa(i))
			// Reset after 10 players to avoid table full errors
			if i%10 == 9 {
				table.Players = table.Players[:0]
			}
		}
	})
	
	b.Run("GetTable", func(b *testing.B) {
		engine.CreateTable("bench_get_table")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = engine.GetTable("bench_get_table")
		}
	})
}

// BenchmarkHandEvaluation benchmarks evaluación de manos
func BenchmarkHandEvaluation(b *testing.B) {
	// Setup cartas de prueba
	playerCards := []Card{
		{Suit: "hearts", Rank: "A"},
		{Suit: "spades", Rank: "K"},
	}
	communityCards := []Card{
		{Suit: "diamonds", Rank: "Q"},
		{Suit: "clubs", Rank: "J"},
		{Suit: "hearts", Rank: "10"},
		{Suit: "spades", Rank: "9"},
		{Suit: "clubs", Rank: "8"},
	}
	
	b.Run("EvaluateHand", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = EvaluateHand(playerCards, communityCards)
		}
	})
	
	b.Run("CompareHands", func(b *testing.B) {
		hand1 := EvaluateHand(playerCards, communityCards)
		hand2 := EvaluateHand([]Card{{Suit: "hearts", Rank: "2"}, {Suit: "spades", Rank: "3"}}, communityCards)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = CompareHands(hand1, hand2)
		}
	})
	
	b.Run("DetermineWinners", func(b *testing.B) {
		// Setup table con jugadores
		engine := NewPokerEngine()
		table := engine.CreateTable("bench_winners")
		engine.AddPlayer("bench_winners", "p1", "Player1")
		engine.AddPlayer("bench_winners", "p2", "Player2")
		
		table.Players[0].Cards = playerCards
		table.Players[1].Cards = []Card{{Suit: "hearts", Rank: "2"}, {Suit: "spades", Rank: "3"}}
		table.CommunityCards = communityCards
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = DetermineWinners(table)
		}
	})
}

// BenchmarkGameFlow benchmarks flujo completo de juego
func BenchmarkGameFlow(b *testing.B) {
	b.Run("CompleteHand", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine := NewPokerEngine()
			table := engine.CreateTable("bench_hand")
			
			// Agregar jugadores
			engine.AddPlayer("bench_hand", "p1", "Player1")
			engine.AddPlayer("bench_hand", "p2", "Player2")
			
			// Marcar ready e iniciar
			table.Players[0].IsReady = true
			table.Players[1].IsReady = true
			engine.StartGame("bench_hand", "p1")
			
			// Simular acciones rápidas hasta showdown
			for table.Phase != "showdown" {
				currentPlayer := table.Players[table.CurrentPlayer]
				if table.CurrentBet > currentPlayer.CurrentBet {
					engine.PlayerAction("bench_hand", currentPlayer.ID, "call", 0)
				} else {
					engine.PlayerAction("bench_hand", currentPlayer.ID, "check", 0)
				}
			}
		}
	})
}

// BenchmarkSidePots benchmarks sistema de side pots
func BenchmarkSidePots(b *testing.B) {
	b.Run("CreateSidePots", func(b *testing.B) {
		engine := NewPokerEngine()
		table := engine.CreateTable("bench_sidepots")
		
		// Setup jugadores con diferentes apuestas
		table.Players = []PokerPlayer{
			{ID: "p1", IsActive: true, CurrentBet: 100},
			{ID: "p2", IsActive: true, CurrentBet: 300},
			{ID: "p3", IsActive: true, CurrentBet: 500},
			{ID: "p4", IsActive: true, CurrentBet: 500},
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine.createSidePots(table)
		}
	})
	
	b.Run("DistributeSidePots", func(b *testing.B) {
		engine := NewPokerEngine()
		table := engine.CreateTable("bench_distribute")
		
		// Setup con side pots y manos
		table.Players = []PokerPlayer{
			{ID: "p1", IsActive: true, Cards: []Card{{Suit: "hearts", Rank: "A"}, {Suit: "spades", Rank: "A"}}},
			{ID: "p2", IsActive: true, Cards: []Card{{Suit: "hearts", Rank: "K"}, {Suit: "spades", Rank: "K"}}},
		}
		table.CommunityCards = []Card{
			{Suit: "diamonds", Rank: "Q"},
			{Suit: "clubs", Rank: "J"},
			{Suit: "hearts", Rank: "10"},
			{Suit: "spades", Rank: "9"},
			{Suit: "clubs", Rank: "8"},
		}
		table.SidePots = []SidePot{
			{Amount: 1000, EligiblePlayers: []int{0, 1}},
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Reset state for each iteration
			table.Players[0].Stack = 1000
			table.Players[1].Stack = 1000
			table.SidePots[0].Amount = 1000
			
			engine.distributeSidePots(table)
		}
	})
}

// BenchmarkConcurrency benchmarks operaciones concurrentes
func BenchmarkConcurrency(b *testing.B) {
	b.Run("ConcurrentTableAccess", func(b *testing.B) {
		engine := NewPokerEngine()
		engine.CreateTable("concurrent_test")
		
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				// Alternar entre operaciones de lectura y escritura
				if i%2 == 0 {
					_, _ = engine.GetTable("concurrent_test")
				} else {
					playerID := "p" + strconv.Itoa(i)
					_, _ = engine.AddPlayer("concurrent_test", playerID, "Player")
				}
				i++
			}
		})
	})
}

// BenchmarkMemoryUsage benchmarks uso de memoria
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("LargeTournament", func(b *testing.B) {
		engine := NewPokerEngine()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Crear 100 mesas con 10 jugadores cada una
			for tableNum := 0; tableNum < 100; tableNum++ {
				tableID := "tournament_" + strconv.Itoa(i) + "_" + strconv.Itoa(tableNum)
				table := engine.CreateTable(tableID)
				
				for playerNum := 0; playerNum < 10; playerNum++ {
					playerID := "p" + strconv.Itoa(playerNum)
					engine.AddPlayer(tableID, playerID, "Player"+strconv.Itoa(playerNum))
				}
				
				// Simular cartas repartidas
				for j := range table.Players {
					table.Players[j].Cards = []Card{
						{Suit: "hearts", Rank: "A"},
						{Suit: "spades", Rank: "K"},
					}
				}
				table.CommunityCards = []Card{
					{Suit: "diamonds", Rank: "Q"},
					{Suit: "clubs", Rank: "J"},
					{Suit: "hearts", Rank: "10"},
					{Suit: "spades", Rank: "9"},
					{Suit: "clubs", Rank: "8"},
				}
			}
		}
	})
}