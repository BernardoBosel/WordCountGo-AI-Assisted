package main

import (
	"reflect"
	"testing"
)

// TestCountWords_CasoMinimo valida o caso de teste mínimo obrigatório da
// atividade: um texto com maiúsculas/minúsculas misturadas, pontuação
// simples, acentos e palavras curtas que devem ser descartadas.
func TestCountWords_CasoMinimo(t *testing.T) {
	text := `Casa, casa! A casa é azul.
Árvore; árvore? verde.
Go go Go. IA é útil, mas IA erra.`

	got := CountWords(text)

	want := map[string]int{
		"casa":   3,
		"árvore": 2,
		"azul":   1,
		"verde":  1,
		"útil":   1,
		"mas":    1,
		"erra":   1,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CountWords() = %#v, esperado %#v", got, want)
	}

	// Confere explicitamente que as palavras curtas foram descartadas,
	// mesmo que por acaso o mapa resultante tivesse o tamanho certo.
	for _, palavraCurta := range []string{"a", "é", "go", "ia"} {
		if _, existe := got[palavraCurta]; existe {
			t.Errorf("palavra curta %q não deveria aparecer no resultado", palavraCurta)
		}
	}
}

// TestCountWords_TextoVazio garante que um texto vazio (ou só com espaços)
// retorna um mapa vazio, sem panics e sem entradas espúrias.
func TestCountWords_TextoVazio(t *testing.T) {
	got := CountWords("")
	if len(got) != 0 {
		t.Fatalf("CountWords(\"\") = %#v, esperado mapa vazio", got)
	}

	got = CountWords("   \n\t  ")
	if len(got) != 0 {
		t.Fatalf("CountWords(espaços em branco) = %#v, esperado mapa vazio", got)
	}
}

// TestCountWords_ApenasPalavrasCurtas garante que um texto composto
// inteiramente por palavras com menos de 3 caracteres resulta em um mapa
// vazio, e não em um mapa com contagens incorretas.
func TestCountWords_ApenasPalavrasCurtas(t *testing.T) {
	text := "eu tu tu é a ia ou vs ok no lá tá dá"

	got := CountWords(text)

	if len(got) != 0 {
		t.Fatalf("CountWords(%q) = %#v, esperado mapa vazio (todas as palavras têm menos de 3 caracteres)", text, got)
	}
}

// TestCountWords_RepeticoesComVariacoes garante que a mesma palavra escrita
// com combinações diferentes de maiúsculas/minúsculas e cercada por
// pontuações distintas é contada como uma única palavra.
func TestCountWords_RepeticoesComVariacoes(t *testing.T) {
	text := "Gato gato GATO gato! Gato? GATO. gato, Cachorro cachorro CACHORRO"

	got := CountWords(text)

	want := map[string]int{
		"gato":     7,
		"cachorro": 3,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CountWords(%q) = %#v, esperado %#v", text, got, want)
	}
}

// TestCountWords_PontuacaoDiversaEEspacamento garante que diferentes tipos
// de pontuação e espaçamento irregular (múltiplos espaços, quebras de
// linha, tabulações) são tratados corretamente como separadores.
func TestCountWords_PontuacaoDiversaEEspacamento(t *testing.T) {
	text := "sol...lua---estrela;;;céu:::mar,,,  sol\n\nlua\t\testrela"

	got := CountWords(text)

	want := map[string]int{
		"sol":     2,
		"lua":     2,
		"estrela": 2,
		"céu":     1,
		"mar":     1,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CountWords(%q) = %#v, esperado %#v", text, got, want)
	}
}
