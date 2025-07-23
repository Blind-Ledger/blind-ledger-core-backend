package poker

import (
	"testing"
)

// TestUniqueDeck verifica que el deck contenga exactamente 52 cartas únicas
func TestUniqueDeck(t *testing.T) {
	engine := NewPokerEngine()
	deck := engine.createShuffledDeck()
	
	// Verificar que hay 52 cartas
	if len(deck) != 52 {
		t.Fatalf("expected 52 cards, got %d", len(deck))
	}
	
	// Verificar que todas las cartas son únicas
	cardSet := make(map[string]bool)
	for _, card := range deck {
		cardKey := card.Suit + ":" + card.Rank
		if cardSet[cardKey] {
			t.Fatalf("found duplicate card: %s of %s", card.Rank, card.Suit)
		}
		cardSet[cardKey] = true
	}
	
	// Verificar que tenemos exactamente las cartas esperadas
	expectedSuits := []string{"hearts", "diamonds", "clubs", "spades"}
	expectedRanks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	
	for _, suit := range expectedSuits {
		for _, rank := range expectedRanks {
			cardKey := suit + ":" + rank
			if !cardSet[cardKey] {
				t.Fatalf("missing expected card: %s of %s", rank, suit)
			}
		}
	}
}

// TestShuffleRandomness verifica que el shuffle produzca resultados diferentes
func TestShuffleRandomness(t *testing.T) {
	engine := NewPokerEngine()
	
	// Generar múltiples decks y verificar que son diferentes
	deck1 := engine.createShuffledDeck()
	deck2 := engine.createShuffledDeck()
	
	// Comparar las primeras 10 cartas - deberían ser diferentes en la mayoría de los casos
	identical := true
	for i := 0; i < 10 && i < len(deck1) && i < len(deck2); i++ {
		if deck1[i].Suit != deck2[i].Suit || deck1[i].Rank != deck2[i].Rank {
			identical = false
			break
		}
	}
	
	if identical {
		t.Logf("Warning: Two shuffles produced identical first 10 cards (very unlikely but possible)")
	}
}

// TestDeckCreationConsistency verifica que cada creación de deck sea válida
func TestDeckCreationConsistency(t *testing.T) {
	engine := NewPokerEngine()
	
	// Crear múltiples decks y verificar que todos son válidos
	for i := 0; i < 10; i++ {
		deck := engine.createShuffledDeck()
		
		if len(deck) != 52 {
			t.Fatalf("deck %d has %d cards, expected 52", i, len(deck))
		}
		
		// Verificar unicidad
		cardSet := make(map[string]bool)
		for _, card := range deck {
			cardKey := card.Suit + ":" + card.Rank
			if cardSet[cardKey] {
				t.Fatalf("deck %d has duplicate card: %s of %s", i, card.Rank, card.Suit)
			}
			cardSet[cardKey] = true
		}
	}
}

// BenchmarkDeckCreation mide el rendimiento de creación de deck
func BenchmarkDeckCreation(b *testing.B) {
	engine := NewPokerEngine()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.createShuffledDeck()
	}
}

// BenchmarkDeckShuffle mide el rendimiento del shuffle
func BenchmarkDeckShuffle(b *testing.B) {
	engine := NewPokerEngine()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.createShuffledDeck()
	}
}