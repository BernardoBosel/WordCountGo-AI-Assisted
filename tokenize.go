package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// tokenize converte o texto para minúsculas, remove pontuação simples
// (qualquer caractere que não seja letra ou número é tratado como
// separador) e descarta palavras com menos de 3 caracteres.
//
// Esta função é usada tanto pela versão sequencial quanto pela versão
// concorrente, garantindo que ambas partam exatamente do mesmo
// conjunto de palavras (a divisão de trabalho concorrente acontece
// apenas na fase de CONTAGEM, não na fase de tokenização).
func tokenize(text string) []string {
	// Remove o BOM (Byte Order Mark) UTF-8, caso exista no início do arquivo.
	text = strings.TrimPrefix(text, "\uFEFF")

	text = strings.ToLower(text)

	rawWords := strings.FieldsFunc(text, func(r rune) bool {
		// Tudo que não for letra ou número é considerado separador
		// (pontuação simples: . , ; : ! ? " ' ( ) [ ] _ - etc.)
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	words := make([]string, 0, len(rawWords))
	for _, w := range rawWords {
		if utf8.RuneCountInString(w) >= 3 {
			words = append(words, w)
		}
	}
	return words
}
