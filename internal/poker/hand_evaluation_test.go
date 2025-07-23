package poker

import (
	"testing"
)

// TestHandEvaluationTableDriven verifica todas las combinaciones críticas de manos
func TestHandEvaluationTableDriven(t *testing.T) {
	tests := []struct {
		name         string
		playerCards  []Card
		communityCards []Card
		expectedRank HandRank
		description  string
	}{
		// Royal Flush
		{
			name: "Royal Flush Hearts",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "hearts", Rank: "K"},
			},
			communityCards: []Card{
				{Suit: "hearts", Rank: "Q"},
				{Suit: "hearts", Rank: "J"},
				{Suit: "hearts", Rank: "10"},
				{Suit: "clubs", Rank: "2"},
				{Suit: "spades", Rank: "3"},
			},
			expectedRank: RoyalFlush,
			description: "A-K-Q-J-10 all hearts",
		},
		
		// Straight Flush
		{
			name: "Straight Flush King High",
			playerCards: []Card{
				{Suit: "spades", Rank: "K"},
				{Suit: "spades", Rank: "Q"},
			},
			communityCards: []Card{
				{Suit: "spades", Rank: "J"},
				{Suit: "spades", Rank: "10"},
				{Suit: "spades", Rank: "9"},
				{Suit: "hearts", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedRank: StraightFlush,
			description: "K-Q-J-10-9 all spades",
		},
		
		// Four of a Kind
		{
			name: "Four Aces",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "A"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "A"},
				{Suit: "hearts", Rank: "K"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedRank: FourOfAKind,
			description: "Four Aces",
		},
		
		// Full House
		{
			name: "Full House Aces over Kings",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "A"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "K"},
				{Suit: "hearts", Rank: "K"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedRank: FullHouse,
			description: "AAA KK",
		},
		
		// Flush
		{
			name: "Ace High Flush",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "hearts", Rank: "Q"},
			},
			communityCards: []Card{
				{Suit: "hearts", Rank: "9"},
				{Suit: "hearts", Rank: "7"},
				{Suit: "hearts", Rank: "3"},
				{Suit: "spades", Rank: "K"},
				{Suit: "clubs", Rank: "2"},
			},
			expectedRank: Flush,
			description: "A-Q-9-7-3 all hearts",
		},
		
		// Straight
		{
			name: "Ace High Straight",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "K"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "Q"},
				{Suit: "clubs", Rank: "J"},
				{Suit: "hearts", Rank: "10"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedRank: Straight,
			description: "A-K-Q-J-10 rainbow",
		},
		
		// Wheel (A-2-3-4-5 straight)
		{
			name: "Wheel Straight",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "2"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "3"},
				{Suit: "clubs", Rank: "4"},
				{Suit: "hearts", Rank: "5"},
				{Suit: "spades", Rank: "K"},
				{Suit: "clubs", Rank: "Q"},
			},
			expectedRank: Straight,
			description: "A-2-3-4-5 straight (wheel)",
		},
		
		// Three of a Kind
		{
			name: "Trip Aces",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "A"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "K"},
				{Suit: "hearts", Rank: "Q"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedRank: ThreeOfAKind,
			description: "AAA with K-Q kickers",
		},
		
		// Two Pair
		{
			name: "Aces and Kings",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "K"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "K"},
				{Suit: "hearts", Rank: "Q"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedRank: TwoPair,
			description: "AA KK with Q kicker",
		},
		
		// One Pair
		{
			name: "Pair of Aces",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "K"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "Q"},
				{Suit: "hearts", Rank: "J"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedRank: OnePair,
			description: "AA with K-Q-J kickers",
		},
		
		// High Card
		{
			name: "Ace High",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "K"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "Q"},
				{Suit: "clubs", Rank: "J"},
				{Suit: "hearts", Rank: "9"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedRank: HighCard,
			description: "A-K-Q-J-9 high",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluation := EvaluateHand(tt.playerCards, tt.communityCards)
			
			if evaluation.Rank != tt.expectedRank {
				t.Errorf("Expected rank %v, got %v for %s", 
					tt.expectedRank, evaluation.Rank, tt.description)
			}
			
			// Verificar que el nombre de la mano sea correcto
			expectedNames := map[HandRank]string{
				HighCard:      "High Card",
				OnePair:       "One Pair",
				TwoPair:       "Two Pair", 
				ThreeOfAKind:  "Three of a Kind",
				Straight:      "Straight",
				Flush:         "Flush",
				FullHouse:     "Full House",
				FourOfAKind:   "Four of a Kind",
				StraightFlush: "Straight Flush",
				RoyalFlush:    "Royal Flush",
			}
			
			if evaluation.RankName != expectedNames[tt.expectedRank] {
				t.Errorf("Expected rank name %s, got %s", 
					expectedNames[tt.expectedRank], evaluation.RankName)
			}
		})
	}
}

// TestHandComparisons verifica que las comparaciones de manos sean correctas
func TestHandComparisons(t *testing.T) {
	tests := []struct {
		name          string
		hand1Cards    []Card
		hand1Community []Card
		hand2Cards    []Card
		hand2Community []Card
		expectedWinner int // 1 si hand1 gana, -1 si hand2 gana, 0 si empate
		description   string
	}{
		{
			name: "Full House beats Flush",
			hand1Cards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "A"},
			},
			hand1Community: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "K"},
				{Suit: "hearts", Rank: "K"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			hand2Cards: []Card{
				{Suit: "hearts", Rank: "Q"},
				{Suit: "hearts", Rank: "J"},
			},
			hand2Community: []Card{
				{Suit: "hearts", Rank: "9"},
				{Suit: "hearts", Rank: "7"},
				{Suit: "hearts", Rank: "3"},
				{Suit: "spades", Rank: "K"},
				{Suit: "clubs", Rank: "2"},
			},
			expectedWinner: 1,
			description: "Full House (AAA KK) vs Flush (hearts)",
		},
		
		{
			name: "Four of a Kind beats Straight Flush",
			hand1Cards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "A"},
			},
			hand1Community: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "A"},
				{Suit: "hearts", Rank: "K"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			hand2Cards: []Card{
				{Suit: "spades", Rank: "9"},
				{Suit: "spades", Rank: "8"},
			},
			hand2Community: []Card{
				{Suit: "spades", Rank: "7"},
				{Suit: "spades", Rank: "6"},
				{Suit: "spades", Rank: "5"},
				{Suit: "hearts", Rank: "K"},
				{Suit: "clubs", Rank: "2"},
			},
			expectedWinner: -1, // Straight Flush beats Four of a Kind
			description: "Four Aces vs 9-high Straight Flush",
		},
		
		{
			name: "Higher Pair Wins",
			hand1Cards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "A"},
			},
			hand1Community: []Card{
				{Suit: "diamonds", Rank: "K"},
				{Suit: "clubs", Rank: "Q"},
				{Suit: "hearts", Rank: "J"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			hand2Cards: []Card{
				{Suit: "hearts", Rank: "K"},
				{Suit: "spades", Rank: "K"},
			},
			hand2Community: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "Q"},
				{Suit: "hearts", Rank: "J"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedWinner: 1,
			description: "Pair of Aces vs Pair of Kings",
		},
		
		{
			name: "Same Pair, Higher Kicker",
			hand1Cards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "A"},
			},
			hand1Community: []Card{
				{Suit: "diamonds", Rank: "K"},
				{Suit: "clubs", Rank: "Q"},
				{Suit: "hearts", Rank: "J"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			hand2Cards: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "A"},
			},
			hand2Community: []Card{
				{Suit: "hearts", Rank: "K"},
				{Suit: "spades", Rank: "Q"},
				{Suit: "diamonds", Rank: "10"}, // Lower kicker than J
				{Suit: "clubs", Rank: "2"},
				{Suit: "hearts", Rank: "3"},
			},
			expectedWinner: 1,
			description: "AA with J kicker vs AA with 10 kicker",
		},
		
		{
			name: "Split Pot - Identical Hands",
			hand1Cards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "K"},
			},
			hand1Community: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "K"},
				{Suit: "hearts", Rank: "Q"},
				{Suit: "spades", Rank: "J"},
				{Suit: "clubs", Rank: "10"},
			},
			hand2Cards: []Card{
				{Suit: "diamonds", Rank: "A"},
				{Suit: "clubs", Rank: "K"},
			},
			hand2Community: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "K"},
				{Suit: "diamonds", Rank: "Q"},
				{Suit: "clubs", Rank: "J"},
				{Suit: "hearts", Rank: "10"},
			},
			expectedWinner: 0,
			description: "Two Pair AA KK identical hands",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand1 := EvaluateHand(tt.hand1Cards, tt.hand1Community)
			hand2 := EvaluateHand(tt.hand2Cards, tt.hand2Community)
			
			result := CompareHands(hand1, hand2)
			
			if result != tt.expectedWinner {
				t.Errorf("Expected winner %d, got %d for %s", 
					tt.expectedWinner, result, tt.description)
				t.Logf("Hand 1: %s (value: %d)", hand1.RankName, hand1.Value)
				t.Logf("Hand 2: %s (value: %d)", hand2.RankName, hand2.Value)
			}
		})
	}
}

// TestEdgeCases verifica casos límite importantes
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		playerCards  []Card
		communityCards []Card
		expectedRank HandRank
		description  string
	}{
		{
			name: "Straight with Ace Low (Wheel)",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "2"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "3"},
				{Suit: "clubs", Rank: "4"},
				{Suit: "hearts", Rank: "5"},
				{Suit: "spades", Rank: "K"},
				{Suit: "clubs", Rank: "Q"},
			},
			expectedRank: Straight,
			description: "A-2-3-4-5 (Ace plays low)",
		},
		
		{
			name: "Full House with Trips on Board", 
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "spades", Rank: "A"},
			},
			communityCards: []Card{
				{Suit: "diamonds", Rank: "K"},
				{Suit: "clubs", Rank: "K"},
				{Suit: "hearts", Rank: "K"},
				{Suit: "spades", Rank: "2"},
				{Suit: "clubs", Rank: "3"},
			},
			expectedRank: FullHouse,
			description: "KKK AA (trips on board)",
		},
		
		{
			name: "Flush with Hearts",
			playerCards: []Card{
				{Suit: "hearts", Rank: "A"},
				{Suit: "hearts", Rank: "K"},
			},
			communityCards: []Card{
				{Suit: "hearts", Rank: "Q"},
				{Suit: "hearts", Rank: "J"},
				{Suit: "hearts", Rank: "9"},
				{Suit: "spades", Rank: "8"},
				{Suit: "clubs", Rank: "7"},
			},
			expectedRank: Flush,
			description: "Hearts flush A-K-Q-J-9",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluation := EvaluateHand(tt.playerCards, tt.communityCards)
			
			if evaluation.Rank != tt.expectedRank {
				t.Errorf("Expected rank %v, got %v for %s", 
					tt.expectedRank, evaluation.Rank, tt.description)
			}
		})
	}
}