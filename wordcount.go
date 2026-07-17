package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// CountWords recebe um texto e retorna um map[string]int com a frequência
// de cada palavra, aplicando as seguintes regras de normalização:
//
//  1. todas as palavras são convertidas para minúsculas;
//  2. pontuação simples é removida (qualquer caractere que não seja letra
//     ou número é tratado como separador de palavras);
//  3. palavras com menos de 3 caracteres (contando runas, não bytes, para
//     preservar corretamente acentos) são ignoradas;
//  4. a contagem final considera apenas as palavras restantes.
func CountWords(text string) map[string]int {
	// Remove o BOM (Byte Order Mark) UTF-8, caso o texto comece com ele.
	text = strings.TrimPrefix(text, "\uFEFF")

	text = strings.ToLower(text)

	// FieldsFunc separa o texto em palavras usando como delimitador
	// qualquer rune que não seja letra nem número. Isso cobre pontuação
	// simples: . , ; : ! ? " ' ( ) [ ] - etc.
	rawWords := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	counts := make(map[string]int)
	for _, w := range rawWords {
		if utf8.RuneCountInString(w) < 3 {
			continue
		}
		counts[w]++
	}

	return counts
}
