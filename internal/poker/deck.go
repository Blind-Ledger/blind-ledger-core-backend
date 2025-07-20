package poker

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"sort"
)

type DeckManager struct{}

func NewDeckManager() *DeckManager {
	return &DeckManager{}
}

// CreateStandardDeck creates a full 52-card deck
func (dm *DeckManager) CreateStandardDeck() []Card {
	var deck []Card

	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	ranks := []Rank{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}

	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}

	return deck
}

// Shuffle returns a shuffled deck using deterministic seed for auditability
func (dm *DeckManager) Shuffle(seed string) []Card {
	deck := dm.CreateStandardDeck()

	// Convert seed to deterministic random source
	hash := sha256.Sum256([]byte(seed))
	seedInt := int64(binary.BigEndian.Uint64(hash[:8]))
	rng := rand.New(rand.NewSource(seedInt))

	// Fisher-Yates shuffle
	for i := len(deck) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}

	return deck
}

// CardToString converts card to readable string (for debugging)
func (c Card) String() string {
	return string(c.Rank) + string(c.Suit)[0:1]
}

// GetRankValue returns numeric value for rank (for hand evaluation)
func (r Rank) Value() int {
	switch r {
	case Two:
		return 2
	case Three:
		return 3
	case Four:
		return 4
	case Five:
		return 5
	case Six:
		return 6
	case Seven:
		return 7
	case Eight:
		return 8
	case Nine:
		return 9
	case Ten:
		return 10
	case Jack:
		return 11
	case Queen:
		return 12
	case King:
		return 13
	case Ace:
		return 14
	default:
		return 0
	}
}

// IsValidCard checks if card is valid
func (c Card) IsValid() bool {
	validSuits := map[Suit]bool{Hearts: true, Diamonds: true, Clubs: true, Spades: true}
	validRanks := map[Rank]bool{
		Two: true, Three: true, Four: true, Five: true, Six: true,
		Seven: true, Eight: true, Nine: true, Ten: true,
		Jack: true, Queen: true, King: true, Ace: true,
	}

	return validSuits[c.Suit] && validRanks[c.Rank]
}

// SortCards sorts cards by rank (useful for hand evaluation)
func SortCards(cards []Card) []Card {
	sorted := make([]Card, len(cards))
	copy(sorted, cards)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Rank.Value() > sorted[j].Rank.Value()
	})

	return sorted
}

// GroupByRank groups cards by rank (for hand evaluation)
func GroupByRank(cards []Card) map[Rank][]Card {
	groups := make(map[Rank][]Card)

	for _, card := range cards {
		groups[card.Rank] = append(groups[card.Rank], card)
	}

	return groups
}

// GroupBySuit groups cards by suit (for flush detection)
func GroupBySuit(cards []Card) map[Suit][]Card {
	groups := make(map[Suit][]Card)

	for _, card := range cards {
		groups[card.Suit] = append(groups[card.Suit], card)
	}

	return groups
}

// HasStraight checks if cards contain a straight
func HasStraight(cards []Card) (bool, []Card) {
	if len(cards) < 5 {
		return false, nil
	}

	// Get unique ranks and sort them
	ranks := make(map[int]Card)
	for _, card := range cards {
		ranks[card.Rank.Value()] = card
	}

	var sortedValues []int
	for value := range ranks {
		sortedValues = append(sortedValues, value)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sortedValues)))

	// Check for 5 consecutive cards
	for i := 0; i <= len(sortedValues)-5; i++ {
		straight := true
		var straightCards []Card

		for j := 0; j < 5; j++ {
			expectedValue := sortedValues[i] - j
			if card, exists := ranks[expectedValue]; exists {
				straightCards = append(straightCards, card)
			} else {
				straight = false
				break
			}
		}

		if straight {
			return true, straightCards
		}
	}

	// Check for wheel straight (A-2-3-4-5)
	wheelValues := []int{14, 5, 4, 3, 2} // Ace low
	wheelStraight := true
	var wheelCards []Card

	for _, value := range wheelValues {
		if card, exists := ranks[value]; exists {
			wheelCards = append(wheelCards, card)
		} else {
			wheelStraight = false
			break
		}
	}

	if wheelStraight {
		return true, wheelCards
	}

	return false, nil
}

// RemoveCards removes specified cards from deck (utility function)
func RemoveCards(deck []Card, toRemove []Card) []Card {
	var result []Card

	removeMap := make(map[Card]bool)
	for _, card := range toRemove {
		removeMap[card] = true
	}

	for _, card := range deck {
		if !removeMap[card] {
			result = append(result, card)
		}
	}

	return result
}
