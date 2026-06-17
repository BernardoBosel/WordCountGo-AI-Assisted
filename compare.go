package main

import "sort"

// mapsEqual verifica se dois mapas de frequência são exatamente
// iguais: mesmas palavras (chaves) e mesma contagem (valor) para cada
// uma delas. Compara o mapa COMPLETO, não apenas o top 20.
func mapsEqual(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for word, countA := range a {
		countB, exists := b[word]
		if !exists || countA != countB {
			return false
		}
	}
	return true
}

// wordFreq associa uma palavra à sua frequência, usado para ordenar
// o resultado final.
type wordFreq struct {
	word  string
	count int
}

// topN retorna as n palavras mais frequentes do mapa, em ordem
// decrescente de frequência. Em caso de empate na frequência, as
// palavras são ordenadas em ordem alfabética para que o resultado
// seja determinístico (mapas em Go não têm ordem garantida).
func topN(counts map[string]int, n int) []wordFreq {
	list := make([]wordFreq, 0, len(counts))
	for word, count := range counts {
		list = append(list, wordFreq{word: word, count: count})
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].count != list[j].count {
			return list[i].count > list[j].count
		}
		return list[i].word < list[j].word
	})

	if len(list) > n {
		list = list[:n]
	}
	return list
}
