// Comando: conta-palavras
//
// Programa em Go para contagem de frequência de palavras em um
// arquivo de texto, com uma versão sequencial e uma versão
// concorrente, comparando corretude e tempo de execução entre elas.
//
// Uso:
//
//	go run . <arquivo.txt> [numWorkers]
//
// Exemplo:
//
//	go run . AChristmasCarol_CharlesDickens_English.txt 4
package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("uso: go run . <arquivo.txt> [numWorkers]")
		os.Exit(1)
	}
	path := os.Args[1]

	// Número de workers (goroutines) da versão concorrente.
	// Padrão: número de CPUs lógicas disponíveis na máquina.
	numWorkers := runtime.NumCPU()
	if len(os.Args) >= 3 {
		parsed, err := strconv.Atoi(os.Args[2])
		if err != nil || parsed <= 0 {
			fmt.Printf("aviso: parâmetro de workers inválido (%q); usando padrão %d\n", os.Args[2], numWorkers)
		} else {
			numWorkers = parsed
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("erro ao ler arquivo:", err)
		os.Exit(1)
	}

	// A tokenização (minúsculas + remoção de pontuação simples +
	// filtro de palavras com menos de 3 caracteres) é feita uma única
	// vez e o resultado é reutilizado pelas duas versões, garantindo
	// que ambas recebam exatamente a mesma entrada.
	words := tokenize(string(raw))

	fmt.Printf("Arquivo: %s\n", path)
	fmt.Printf("Total de palavras após filtragem: %d\n", len(words))
	fmt.Printf("Workers (versão concorrente): %d\n\n", numWorkers)

	// --- Versão sequencial -------------------------------------------------
	startSeq := time.Now()
	seqCounts := sequentialCount(words)
	seqDuration := time.Since(startSeq)

	// --- Versão concorrente --------------------------------------------------
	startConc := time.Now()
	concCounts := concurrentCount(words, numWorkers)
	concDuration := time.Since(startConc)

	// --- Comparação de correção ------------------------------------------
	equal := mapsEqual(seqCounts, concCounts)

	// --- Saída ---------------------------------------------------------------
	fmt.Printf("Tempo sequencial: %v\n", seqDuration)
	fmt.Printf("Tempo concorrente: %v\n", concDuration)
	if equal {
		fmt.Println("Resultados iguais: sim")
	} else {
		fmt.Println("Resultados iguais: não")
	}

	if concDuration < seqDuration {
		speedup := float64(seqDuration) / float64(concDuration)
		fmt.Printf("A versão concorrente foi %.2fx mais rápida que a sequencial\n", speedup)
	} else if concDuration > seqDuration {
		slowdown := float64(concDuration) / float64(seqDuration)
		fmt.Printf("A versão concorrente foi %.2fx mais lenta que a sequencial\n", slowdown)
	} else {
		fmt.Println("As duas versões tiveram tempo de execução equivalente")
	}

	fmt.Println("\nTop 20 palavras:")
	top := topN(seqCounts, 20)
	for i, wf := range top {
		fmt.Printf("%2d. %-15s %d\n", i+1, wf.word, wf.count)
	}
}
