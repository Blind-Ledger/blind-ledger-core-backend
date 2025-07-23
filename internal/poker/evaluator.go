package poker

import (
	"sort"
)

// HandRank representa la fuerza de una mano
type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

// HandEvaluation contiene el resultado de evaluar una mano
type HandEvaluation struct {
	Rank     HandRank `json:"rank"`
	Value    int      `json:"value"`    // Valor numérico para comparación
	Cards    []Card   `json:"cards"`    // Las 5 mejores cartas
	RankName string   `json:"rank_name"`
}

// CardValue convierte rank de carta a valor numérico
func CardValue(rank string) int {
	switch rank {
	case "2":
		return 2
	case "3":
		return 3
	case "4":
		return 4
	case "5":
		return 5
	case "6":
		return 6
	case "7":
		return 7
	case "8":
		return 8
	case "9":
		return 9
	case "10":
		return 10
	case "J":
		return 11
	case "Q":
		return 12
	case "K":
		return 13
	case "A":
		return 14
	default:
		return 0
	}
}

// EvaluateHand evalúa la mejor mano de 5 cartas de las 7 disponibles
func EvaluateHand(playerCards []Card, communityCards []Card) HandEvaluation {
	allCards := append(playerCards, communityCards...)
	
	// Generar todas las combinaciones posibles de 5 cartas
	bestHand := HandEvaluation{Rank: HighCard, Value: 0}
	
	// Si hay menos de 5 cartas, evaluar lo que hay
	if len(allCards) < 5 {
		return evaluateFiveCards(allCards)
	}
	
	// Generar combinaciones de 5 cartas
	combinations := generateCombinations(allCards, 5)
	
	for _, combo := range combinations {
		evaluation := evaluateFiveCards(combo)
		if evaluation.Value > bestHand.Value {
			bestHand = evaluation
		}
	}
	
	return bestHand
}

// evaluateFiveCards evalúa exactamente 5 cartas
func evaluateFiveCards(cards []Card) HandEvaluation {
	if len(cards) == 0 {
		return HandEvaluation{Rank: HighCard, Value: 0, RankName: "High Card"}
	}
	
	// Copiar y ordenar cartas por valor
	sortedCards := make([]Card, len(cards))
	copy(sortedCards, cards)
	sort.Slice(sortedCards, func(i, j int) bool {
		return CardValue(sortedCards[i].Rank) > CardValue(sortedCards[j].Rank)
	})
	
	// Contar ranks y suits
	rankCounts := make(map[string]int)
	suitCounts := make(map[string]int)
	
	for _, card := range sortedCards {
		rankCounts[card.Rank]++
		suitCounts[card.Suit]++
	}
	
	// Verificar flush - necesitamos exactamente 5 cartas del mismo palo
	isFlush := false
	flushSuit := ""
	for suit, count := range suitCounts {
		if count >= 5 {
			isFlush = true
			flushSuit = suit
			break
		}
	}
	
	// Verificar straight
	isStraight, straightHigh := checkStraight(sortedCards)
	
	// Evaluar combinaciones
	if isFlush && isStraight {
		if straightHigh == 14 { // A-K-Q-J-10
			return HandEvaluation{
				Rank:     RoyalFlush,
				Value:    1000000,
				Cards:    sortedCards,
				RankName: "Royal Flush",
			}
		}
		return HandEvaluation{
			Rank:     StraightFlush,
			Value:    900000 + straightHigh,
			Cards:    sortedCards,
			RankName: "Straight Flush",
		}
	}
	
	// Buscar grupos de ranks
	var pairs, threes, fours []string
	for rank, count := range rankCounts {
		switch count {
		case 2:
			pairs = append(pairs, rank)
		case 3:
			threes = append(threes, rank)
		case 4:
			fours = append(fours, rank)
		}
	}
	
	// Four of a kind
	if len(fours) > 0 {
		fourValue := CardValue(fours[0])
		return HandEvaluation{
			Rank:     FourOfAKind,
			Value:    800000 + fourValue*1000,
			Cards:    sortedCards,
			RankName: "Four of a Kind",
		}
	}
	
	// Full house
	if len(threes) > 0 && len(pairs) > 0 {
		threeValue := CardValue(threes[0])
		pairValue := CardValue(pairs[0])
		return HandEvaluation{
			Rank:     FullHouse,
			Value:    700000 + threeValue*1000 + pairValue,
			Cards:    sortedCards,
			RankName: "Full House",
		}
	}
	
	// Flush
	if isFlush {
		// Obtener las 5 cartas más altas del palo del flush
		flushCards := make([]Card, 0, 5)
		for _, card := range sortedCards {
			if card.Suit == flushSuit && len(flushCards) < 5 {
				flushCards = append(flushCards, card)
			}
		}
		
		if len(flushCards) >= 5 {
			highCard := CardValue(flushCards[0].Rank)
			return HandEvaluation{
				Rank:     Flush,
				Value:    600000 + highCard*1000,
				Cards:    flushCards[:5], // Solo las 5 mejores cartas del flush
				RankName: "Flush",
			}
		}
	}
	
	// Straight
	if isStraight {
		return HandEvaluation{
			Rank:     Straight,
			Value:    500000 + straightHigh,
			Cards:    sortedCards,
			RankName: "Straight",
		}
	}
	
	// Three of a kind
	if len(threes) > 0 {
		threeValue := CardValue(threes[0])
		return HandEvaluation{
			Rank:     ThreeOfAKind,
			Value:    400000 + threeValue*1000,
			Cards:    sortedCards,
			RankName: "Three of a Kind",
		}
	}
	
	// Two pair
	if len(pairs) >= 2 {
		// Ordenar pairs por valor
		sort.Slice(pairs, func(i, j int) bool {
			return CardValue(pairs[i]) > CardValue(pairs[j])
		})
		highPair := CardValue(pairs[0])
		lowPair := CardValue(pairs[1])
		return HandEvaluation{
			Rank:     TwoPair,
			Value:    300000 + highPair*1000 + lowPair*100,
			Cards:    sortedCards,
			RankName: "Two Pair",
		}
	}
	
	// One pair
	if len(pairs) > 0 {
		pairValue := CardValue(pairs[0])
		
		// Calcular kickers (las 3 cartas más altas que no sean el par)
		kickers := make([]int, 0, 3)
		for _, card := range sortedCards {
			if card.Rank != pairs[0] && len(kickers) < 3 {
				kickers = append(kickers, CardValue(card.Rank))
			}
		}
		
		// Incluir kickers en el valor
		kickerValue := 0
		if len(kickers) > 0 {
			kickerValue += kickers[0] * 100 // Primer kicker
		}
		if len(kickers) > 1 {
			kickerValue += kickers[1] * 10 // Segundo kicker
		}
		if len(kickers) > 2 {
			kickerValue += kickers[2] // Tercer kicker
		}
		
		return HandEvaluation{
			Rank:     OnePair,
			Value:    200000 + pairValue*1000 + kickerValue,
			Cards:    sortedCards,
			RankName: "One Pair",
		}
	}
	
	// High card
	highCard := CardValue(sortedCards[0].Rank)
	return HandEvaluation{
		Rank:     HighCard,
		Value:    100000 + highCard*1000,
		Cards:    sortedCards,
		RankName: "High Card",
	}
}

// checkStraight verifica si hay una escalera
func checkStraight(sortedCards []Card) (bool, int) {
	if len(sortedCards) < 5 {
		return false, 0
	}
	
	values := make([]int, 0, len(sortedCards))
	for _, card := range sortedCards {
		value := CardValue(card.Rank)
		// Evitar duplicados
		if len(values) == 0 || values[len(values)-1] != value {
			values = append(values, value)
		}
	}
	
	// Verificar escalera normal
	for i := 0; i <= len(values)-5; i++ {
		if values[i]-values[i+4] == 4 {
			return true, values[i]
		}
	}
	
	// Verificar escalera baja A-2-3-4-5
	if len(values) >= 5 {
		hasAce := false
		has2 := false
		has3 := false
		has4 := false
		has5 := false
		
		for _, v := range values {
			switch v {
			case 14:
				hasAce = true
			case 2:
				has2 = true
			case 3:
				has3 = true
			case 4:
				has4 = true
			case 5:
				has5 = true
			}
		}
		
		if hasAce && has2 && has3 && has4 && has5 {
			return true, 5 // El 5 es la carta alta en esta escalera
		}
	}
	
	return false, 0
}

// generateCombinations genera todas las combinaciones de k elementos de un slice
func generateCombinations(cards []Card, k int) [][]Card {
	if k > len(cards) {
		return [][]Card{}
	}
	
	if k == 0 {
		return [][]Card{{}}
	}
	
	if k == len(cards) {
		return [][]Card{cards}
	}
	
	var result [][]Card
	
	// Incluir el primer elemento
	head := cards[0]
	tail := cards[1:]
	
	for _, combo := range generateCombinations(tail, k-1) {
		newCombo := make([]Card, 0, k)
		newCombo = append(newCombo, head)
		newCombo = append(newCombo, combo...)
		result = append(result, newCombo)
	}
	
	// No incluir el primer elemento
	result = append(result, generateCombinations(tail, k)...)
	
	return result
}

// CompareHands compara dos evaluaciones de manos
func CompareHands(hand1, hand2 HandEvaluation) int {
	if hand1.Value > hand2.Value {
		return 1
	} else if hand1.Value < hand2.Value {
		return -1
	}
	return 0
}

// DetermineWinners encuentra los ganadores de una mesa
func DetermineWinners(table *PokerTable) []int {
	if len(table.Players) == 0 {
		return []int{}
	}
	
	var winners []int
	bestEvaluation := HandEvaluation{Value: -1}
	
	for i, player := range table.Players {
		if player.IsActive && !player.HasFolded && len(player.Cards) > 0 {
			evaluation := EvaluateHand(player.Cards, table.CommunityCards)
			
			if evaluation.Value > bestEvaluation.Value {
				bestEvaluation = evaluation
				winners = []int{i}
			} else if evaluation.Value == bestEvaluation.Value {
				winners = append(winners, i)
			}
		}
	}
	
	return winners
}