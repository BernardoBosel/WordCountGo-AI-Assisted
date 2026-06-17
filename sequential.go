package main

// sequentialCount realiza a contagem de frequência de palavras de
// forma puramente sequencial, sem nenhum uso de goroutines, channels
// ou outros recursos de concorrência. Serve como referência de
// correção para a versão concorrente.
func sequentialCount(words []string) map[string]int {
	counts := make(map[string]int, len(words)/2)
	for _, w := range words {
		counts[w]++
	}
	return counts
}
