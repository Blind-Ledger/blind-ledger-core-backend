package poker

import (
	"sort"
)

type HandEvaluator struct{}

func NewHandEvaluator() *HandEvaluator {
	return &HandEvaluator{}
}

// EvaluateHand finds the best 5-card hand from hole cards + community cards
func (he *HandEvaluator) EvaluateHand(holeCards []Card, communityCards []Card) HandResult {
	// Combine all available cards
	allCards := append(holeCards, communityCards...)

	if len(allCards) < 5 {
		// Not enough cards for a valid hand
		return HandResult{
			Type:    HighCard,
			Cards:   allCards,
			Rank:    0,
			Kickers: []Rank{},
		}
	}

	// Try all possible 5-card combinations and find the best
	bestHand := HandResult{Type: HighCard, Rank: 0}

	combinations := he.generateCombinations(allCards, 5)
	for _, combo := range combinations {
		hand := he.evaluateFiveCards(combo)
		if he.compareHandResults(hand, bestHand) > 0 {
			bestHand = hand
		}
	}

	return bestHand
}

// CompareHands returns 1 if hand1 > hand2, -1 if hand1 < hand2, 0 if equal
func (he *HandEvaluator) CompareHands(hand1, hand2 HandResult) int {
	return he.compareHandResults(hand1, hand2)
}

// evaluateFiveCards evaluates exactly 5 cards and returns the hand type
func (he *HandEvaluator) evaluateFiveCards(cards []Card) HandResult {
	if len(cards) != 5 {
		return HandResult{Type: HighCard, Cards: cards, Rank: 0}
	}

	sorted := SortCards(cards)

	// Check for flush
	isFlush := he.isFlush(sorted)

	// Check for straight
	isStraight, straightCards := he.isStraight(sorted)

	// Straight Flush
	if isFlush && isStraight {
		if he.isRoyalFlush(straightCards) {
			return HandResult{
				Type:    RoyalFlush,
				Cards:   straightCards,
				Rank:    1000, // Highest possible rank
				Kickers: []Rank{},
			}
		}
		return HandResult{
			Type:    StraightFlush,
			Cards:   straightCards,
			Rank:    800 + straightCards[0].Rank.Value(), // High card of straight
			Kickers: []Rank{},
		}
	}

	// Group by rank for pair-based hands
	rankGroups := GroupByRank(sorted)
	groupSizes := he.getGroupSizes(rankGroups)

	// Four of a Kind
	if len(groupSizes) == 2 && groupSizes[0] == 4 {
		quad, kicker := he.getQuadsAndKicker(rankGroups)
		return HandResult{
			Type:    FourOfAKind,
			Cards:   sorted,
			Rank:    700 + quad.Value(),
			Kickers: []Rank{kicker},
		}
	}

	// Full House
	if len(groupSizes) == 2 && groupSizes[0] == 3 && groupSizes[1] == 2 {
		trips, pair := he.getTripsAndPair(rankGroups)
		return HandResult{
			Type:    FullHouse,
			Cards:   sorted,
			Rank:    600 + trips.Value(),
			Kickers: []Rank{pair},
		}
	}

	// Flush
	if isFlush {
		kickers := he.getRanksDescending(sorted)
		return HandResult{
			Type:    Flush,
			Cards:   sorted,
			Rank:    500 + kickers[0].Value(),
			Kickers: kickers[1:],
		}
	}

	// Straight
	if isStraight {
		return HandResult{
			Type:    Straight,
			Cards:   straightCards,
			Rank:    400 + straightCards[0].Rank.Value(),
			Kickers: []Rank{},
		}
	}

	// Three of a Kind
	if len(groupSizes) == 3 && groupSizes[0] == 3 {
		trips, kickers := he.getTripsAndKickers(rankGroups)
		return HandResult{
			Type:    ThreeOfAKind,
			Cards:   sorted,
			Rank:    300 + trips.Value(),
			Kickers: kickers,
		}
	}

	// Two Pair
	if len(groupSizes) == 3 && groupSizes[0] == 2 && groupSizes[1] == 2 {
		highPair, lowPair, kicker := he.getTwoPairAndKicker(rankGroups)
		return HandResult{
			Type:    TwoPair,
			Cards:   sorted,
			Rank:    200 + highPair.Value(),
			Kickers: []Rank{lowPair, kicker},
		}
	}

	// One Pair
	if len(groupSizes) == 4 && groupSizes[0] == 2 {
		pair, kickers := he.getPairAndKickers(rankGroups)
		return HandResult{
			Type:    Pair,
			Cards:   sorted,
			Rank:    100 + pair.Value(),
			Kickers: kickers,
		}
	}

	// High Card
	kickers := he.getRanksDescending(sorted)
	return HandResult{
		Type:    HighCard,
		Cards:   sorted,
		Rank:    kickers[0].Value(),
		Kickers: kickers[1:],
	}
}

// Helper methods for hand evaluation

func (he *HandEvaluator) isFlush(cards []Card) bool {
	if len(cards) != 5 {
		return false
	}

	suit := cards[0].Suit
	for _, card := range cards[1:] {
		if card.Suit != suit {
			return false
		}
	}
	return true
}

func (he *HandEvaluator) isStraight(cards []Card) (bool, []Card) {
	if len(cards) != 5 {
		return false, nil
	}

	sorted := SortCards(cards)

	// Check for regular straight
	for i := 0; i < 4; i++ {
		if sorted[i].Rank.Value()-sorted[i+1].Rank.Value() != 1 {
			break
		}
		if i == 3 { // Made it through all 4 gaps
			return true, sorted
		}
	}

	// Check for wheel straight (A-2-3-4-5)
	if sorted[0].Rank == Ace && sorted[1].Rank == Five &&
		sorted[2].Rank == Four && sorted[3].Rank == Three && sorted[4].Rank == Two {
		// Rearrange to put Ace at the end for wheel
		wheel := []Card{sorted[1], sorted[2], sorted[3], sorted[4], sorted[0]}
		return true, wheel
	}

	return false, nil
}

func (he *HandEvaluator) isRoyalFlush(cards []Card) bool {
	if len(cards) != 5 {
		return false
	}

	sorted := SortCards(cards)
	return sorted[0].Rank == Ace && sorted[1].Rank == King &&
		sorted[2].Rank == Queen && sorted[3].Rank == Jack && sorted[4].Rank == Ten
}

func (he *HandEvaluator) getGroupSizes(rankGroups map[Rank][]Card) []int {
	var sizes []int
	for _, group := range rankGroups {
		sizes = append(sizes, len(group))
	}

	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
	return sizes
}

func (he *HandEvaluator) getQuadsAndKicker(rankGroups map[Rank][]Card) (Rank, Rank) {
	var quad, kicker Rank

	for rank, cards := range rankGroups {
		if len(cards) == 4 {
			quad = rank
		} else {
			kicker = rank
		}
	}

	return quad, kicker
}

func (he *HandEvaluator) getTripsAndPair(rankGroups map[Rank][]Card) (Rank, Rank) {
	var trips, pair Rank

	for rank, cards := range rankGroups {
		if len(cards) == 3 {
			trips = rank
		} else {
			pair = rank
		}
	}

	return trips, pair
}

func (he *HandEvaluator) getTripsAndKickers(rankGroups map[Rank][]Card) (Rank, []Rank) {
	var trips Rank
	var kickers []Rank

	for rank, cards := range rankGroups {
		if len(cards) == 3 {
			trips = rank
		} else {
			kickers = append(kickers, rank)
		}
	}

	// Sort kickers descending
	sort.Slice(kickers, func(i, j int) bool {
		return kickers[i].Value() > kickers[j].Value()
	})

	return trips, kickers
}

func (he *HandEvaluator) getTwoPairAndKicker(rankGroups map[Rank][]Card) (Rank, Rank, Rank) {
	var pairs []Rank
	var kicker Rank

	for rank, cards := range rankGroups {
		if len(cards) == 2 {
			pairs = append(pairs, rank)
		} else {
			kicker = rank
		}
	}

	// Sort pairs descending
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value() > pairs[j].Value()
	})

	return pairs[0], pairs[1], kicker
}

func (he *HandEvaluator) getPairAndKickers(rankGroups map[Rank][]Card) (Rank, []Rank) {
	var pair Rank
	var kickers []Rank

	for rank, cards := range rankGroups {
		if len(cards) == 2 {
			pair = rank
		} else {
			kickers = append(kickers, rank)
		}
	}

	// Sort kickers descending
	sort.Slice(kickers, func(i, j int) bool {
		return kickers[i].Value() > kickers[j].Value()
	})

	return pair, kickers
}

func (he *HandEvaluator) getRanksDescending(cards []Card) []Rank {
	var ranks []Rank
	for _, card := range cards {
		ranks = append(ranks, card.Rank)
	}

	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].Value() > ranks[j].Value()
	})

	return ranks
}

func (he *HandEvaluator) compareHandResults(hand1, hand2 HandResult) int {
	// First compare by rank
	if hand1.Rank > hand2.Rank {
		return 1
	}
	if hand1.Rank < hand2.Rank {
		return -1
	}

	// If ranks are equal, compare kickers
	for i := 0; i < len(hand1.Kickers) && i < len(hand2.Kickers); i++ {
		if hand1.Kickers[i].Value() > hand2.Kickers[i].Value() {
			return 1
		}
		if hand1.Kickers[i].Value() < hand2.Kickers[i].Value() {
			return -1
		}
	}

	// If we get here, hands are equal
	return 0
}

// generateCombinations generates all possible combinations of r elements from slice
func (he *HandEvaluator) generateCombinations(cards []Card, r int) [][]Card {
	var result [][]Card

	if r > len(cards) {
		return result
	}

	if r == 0 {
		return [][]Card{{}}
	}

	if r == len(cards) {
		cardsCopy := make([]Card, len(cards))
		copy(cardsCopy, cards)
		return [][]Card{cardsCopy}
	}

	// Include first element
	firstElement := cards[0]
	remaining := cards[1:]

	smallerCombos := he.generateCombinations(remaining, r-1)
	for _, combo := range smallerCombos {
		newCombo := append([]Card{firstElement}, combo...)
		result = append(result, newCombo)
	}

	// Exclude first element
	result = append(result, he.generateCombinations(remaining, r)...)

	return result
}
