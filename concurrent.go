package main

import "sync"

// concurrentCount realiza a contagem de frequência de palavras usando
// o padrão "fan-out / fan-in":
//
//  1. (fan-out) a fatia de palavras é dividida em "numWorkers" blocos
//     contíguos, de tamanho aproximadamente igual;
//  2. cada bloco é processado por uma goroutine independente, que
//     monta seu PRÓPRIO mapa local de contagens (sem nenhum
//     compartilhamento de memória entre goroutines, logo sem
//     necessidade de mutex nessa etapa);
//  3. (fan-in) cada goroutine envia seu mapa local para um channel;
//  4. depois que todas as goroutines terminam (sync.WaitGroup), a
//     goroutine principal lê todos os mapas parciais do channel e os
//     combina (reduce) em um único mapa final, somando as
//     contagens de cada palavra.
//
// Como a etapa de combinação (passo 4) é feita inteiramente pela
// goroutine principal, depois que todas as goroutines de contagem já
// terminaram (wg.Wait()), não há acesso concorrente ao mapa final e,
// portanto, não é necessário usar mutex nem channels não teriam
// problema de corrida de dados (data race) nessa etapa.
func concurrentCount(words []string, numWorkers int) map[string]int {
	n := len(words)

	if numWorkers <= 0 {
		numWorkers = 1
	}
	if numWorkers > n && n > 0 {
		numWorkers = n
	}
	if n == 0 {
		return make(map[string]int)
	}

	chunkSize := (n + numWorkers - 1) / numWorkers // arredonda para cima
	partials := make(chan map[string]int, numWorkers)
	var wg sync.WaitGroup

	for start := 0; start < n; start += chunkSize {
		end := start + chunkSize
		if end > n {
			end = n
		}

		wg.Add(1)
		go func(chunk []string) {
			defer wg.Done()

			local := make(map[string]int)
			for _, w := range chunk {
				local[w]++
			}
			partials <- local
		}(words[start:end])
	}

	// Espera todas as goroutines de contagem terminarem e então fecha
	// o channel, sinalizando que não há mais mapas parciais a serem
	// recebidos.
	wg.Wait()
	close(partials)

	// Combinação (reduce) sequencial dos mapas parciais.
	final := make(map[string]int, n/2)
	for local := range partials {
		for word, count := range local {
			final[word] += count
		}
	}
	return final
}
