# Word Count em Go — Teste Automático com Apoio de IA

Atividade individual cujo objetivo é implementar, em Go, uma função de
contagem de frequência de palavras (`CountWords`) e validá-la por meio de
testes automáticos, usando uma ferramenta de IA como apoio em todas as
etapas: compreensão da especificação, implementação, criação de testes,
revisão crítica e verificação de cobertura.

O processo completo de uso da IA — prompts utilizados, respostas obtidas,
sugestões aceitas/rejeitadas e erros encontrados — está documentado em
[`PROMPT.md`](./PROMPT.md).

## A função `CountWords`

```go
func CountWords(text string) map[string]int
```

Recebe um texto e devolve um `map[string]int` com a frequência de cada
palavra, aplicando as seguintes regras de normalização:

1. **Minúsculas:** todo o texto é convertido com `strings.ToLower` antes de
   ser dividido em palavras.
2. **Remoção de pontuação simples:** a separação usa `strings.FieldsFunc`
   com um delimitador que trata como separador qualquer caractere que
   **não** seja letra nem número (`unicode.IsLetter` / `unicode.IsNumber`).
   Isso cobre vírgula, ponto, ponto e vírgula, interrogação, exclamação,
   aspas, parênteses, travessões, reticências etc., sem precisar de uma
   lista explícita de símbolos.
3. **Descarte de palavras curtas:** palavras com menos de 3 caracteres são
   ignoradas. A contagem de caracteres usa `utf8.RuneCountInString`, e não
   `len()`, para que letras acentuadas (contando como 1 caractere cada, e
   não 2 bytes) sejam avaliadas corretamente — por isso `"útil"` é mantida
   (4 caracteres) e `"é"` é descartada (1 caractere).
4. **Contagem:** cada palavra restante incrementa sua entrada no mapa de
   resultado.

O código também remove o BOM (Byte Order Mark) UTF-8 do início do texto,
caso ele exista, para que não vire uma "palavra" espúria.

## Como executar os testes

Requer Go 1.22 ou superior.

```bash
git clone <URL-DO-REPOSITORIO>
cd <pasta-do-repositorio>
go test -v ./...
```

### Resultado obtido ao executar `go test`

```
=== RUN   TestCountWords_CasoMinimo
--- PASS: TestCountWords_CasoMinimo (0.00s)
=== RUN   TestCountWords_TextoVazio
--- PASS: TestCountWords_TextoVazio (0.00s)
=== RUN   TestCountWords_ApenasPalavrasCurtas
--- PASS: TestCountWords_ApenasPalavrasCurtas (0.00s)
=== RUN   TestCountWords_RepeticoesComVariacoes
--- PASS: TestCountWords_RepeticoesComVariacoes (0.00s)
=== RUN   TestCountWords_PontuacaoDiversaEEspacamento
--- PASS: TestCountWords_PontuacaoDiversaEEspacamento (0.00s)
PASS
coverage: 100.0% of statements
ok      wordcount       0.003s  coverage: 100.0% of statements
```

## Casos de teste implementados

| Teste | O que valida |
| --- | --- |
| `TestCountWords_CasoMinimo` | **Caso mínimo obrigatório** da atividade: texto com maiúsculas/minúsculas misturadas, pontuação simples e acentos. Compara o mapa produzido com o mapa esperado exato (`reflect.DeepEqual`) e ainda confere, de forma explícita, que as palavras curtas (`a`, `é`, `go`, `ia`) não aparecem no resultado. |
| `TestCountWords_TextoVazio` | Texto vazio (`""`) e texto contendo só espaços em branco devem retornar um mapa vazio, sem panics. |
| `TestCountWords_ApenasPalavrasCurtas` | Texto composto inteiramente por palavras com menos de 3 caracteres deve resultar em mapa vazio. |
| `TestCountWords_RepeticoesComVariacoes` | A mesma palavra escrita com combinações diferentes de maiúsculas/minúsculas e cercada por pontuações distintas (`Gato`, `gato,`, `GATO.`, `Gato?`) deve ser contada como uma única chave no mapa. |
| `TestCountWords_PontuacaoDiversaEEspacamento` | Diferentes símbolos de pontuação (`...`, `---`, `;;;`, `:::`, `,,,`) e espaçamento irregular (múltiplos espaços, quebras de linha, tabulações) devem ser tratados corretamente como separadores de palavras. |

Todos os testes comparam o mapa **completo** produzido pela função com o
mapa esperado (via `reflect.DeepEqual`), e não apenas uma amostra do
resultado — nenhum teste se limita a imprimir o resultado na tela.

## Estrutura do repositório

| Arquivo | Conteúdo |
| --- | --- |
| `go.mod` | Definição do módulo Go. |
| `wordcount.go` | Implementação da função `CountWords`. |
| `wordcount_test.go` | Testes automáticos (`go test`). |
| `README.md` | Este arquivo. |
| `PROMPT.md` | Registro do uso da IA ao longo da atividade. |
