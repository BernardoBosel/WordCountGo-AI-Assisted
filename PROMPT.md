# PROMPT.md — Registro de Uso de IA

Este arquivo documenta como a IA foi utilizada no desenvolvimento da função
`CountWords` e de seus testes automáticos, seguindo a sequência obrigatória
de prompts do roteiro da atividade (seção 6 do enunciado).

> **Nota:** os seis prompts obrigatórios foram enviados nesta ordem, na
> mesma conversa, à ferramenta indicada abaixo. As respostas foram
> resumidas aqui em vez de coladas na íntegra; o raciocínio, as decisões e
> o erro relatado na Etapa 3 refletem o que de fato aconteceu durante essa
> conversa. **Revise este texto antes de entregar** e ajuste qualquer trecho
> que não reflita o seu próprio julgamento — o objetivo do arquivo é
> mostrar avaliação crítica genuína, não só preencher o formato.

## Ferramenta e ambiente utilizados

- **Ferramenta de IA:** Claude (Anthropic), via interface de chat.
- **Ambiente:** interface de chat do Claude, com execução de código
  habilitada (sandbox Linux/Ubuntu com terminal bash), usada para instalar
  o Go, formatar (`gofmt`), vetar (`go vet`) e executar os testes
  (`go test -v -cover`) antes de considerar a atividade concluída.
- **Versão do Go usada para compilar/testar:** `go1.22.2 linux/amd64`.
- **Contexto adicional fornecido à IA:** um repositório anterior do mesmo
  autor (`WordCountGo-AI-Assisted`), que já continha uma função de
  tokenização (`tokenize.go`) com regras semelhantes (minúsculas, remoção
  de pontuação simples via `unicode.IsLetter`/`unicode.IsNumber`, e filtro
  de palavras curtas por contagem de *runas*). Essa lógica foi reaproveitada
  como ponto de partida para `CountWords`, em vez de pedir à IA para
  reinventar a normalização do zero.

---

## Etapa 1 — Compreensão do problema

**Prompt usado (obrigatório):**

> Quero implementar em Go uma função `CountWords(text string) map[string]int`.
> A função deve contar a frequência das palavras em um texto, converter
> palavras para minúsculas, remover pontuação simples e ignorar palavras com
> menos de 3 caracteres. Explique a especificação e os principais cuidados
> de implementação.

**Resumo da resposta da IA:**

- Minúsculas devem ser aplicadas com `strings.ToLower` **antes** de dividir
  o texto em palavras, para não depender de comparações case-insensitive
  espalhadas pelo código.
- "Pontuação simples" pode ser tratada de forma genérica com
  `strings.FieldsFunc`, usando como separador qualquer caractere que não
  seja letra nem número (`unicode.IsLetter` / `unicode.IsNumber`), em vez
  de montar uma lista fixa de símbolos — isso cobre vírgula, ponto,
  interrogação, exclamação, aspas, parênteses etc. automaticamente.
- O corte de palavras com menos de 3 caracteres precisa contar **runas**
  (`utf8.RuneCountInString`), não bytes (`len`), porque caracteres
  acentuados como `á`, `é`, `í`, `ó`, `ú` ocupam 2 bytes em UTF-8 mas devem
  contar como 1 caractere.
- É preciso cuidado com o BOM (Byte Order Mark, `\uFEFF`) caso o texto de
  entrada venha de um arquivo salvo com esse marcador — sem tratamento, ele
  pode grudar na primeira palavra ou virar um token vazio.
- O mapa de resultado deve conter **apenas** as palavras que sobrevivem aos
  filtros; palavras descartadas não devem aparecer nem com contagem zero.

**Cuidados identificados:**

- Diferenciar contagem por *runas* de contagem por *bytes* é essencial em
  português por causa dos acentos — um erro comum seria usar `len(w) < 3`.
- A ordem das operações importa: normalizar para minúsculas antes de
  comparar/agrupar, e filtrar por tamanho **depois** de já ter separado a
  pontuação (senão `"casa."` teria 5 caracteres e não seria filtrada mesmo
  que devesse).

**Decisões tomadas:**

- Reaproveitar a abordagem de tokenização já usada em `tokenize.go` do
  projeto anterior (mesma lógica, mesmas garantias), em vez de escrever uma
  versão nova do zero, já que ela já resolve exatamente os pontos acima.
- Consolidar tudo em uma única função `CountWords`, sem função de
  tokenização separada, já que a atividade pede apenas uma função pública.

---

## Etapa 2 — Implementação da função

**Prompt usado (obrigatório):**

> Implemente em Go a função `CountWords(text string) map[string]int`. A
> função deve converter palavras para minúsculas, remover pontuação simples,
> ignorar palavras com menos de 3 caracteres e retornar um `map[string]int`
> com a frequência das palavras.

**Código gerado:** ver [`wordcount.go`](./wordcount.go).

**Verificação feita após receber o código:**

| Pergunta | Resultado |
| --- | --- |
| O código compila? | Sim (`go vet ./...` sem erros). |
| A função tem a assinatura esperada? | Sim, `func CountWords(text string) map[string]int`. |
| A normalização está correta? | Sim, `strings.ToLower` aplicado antes da separação em palavras. |
| Palavras curtas são ignoradas? | Sim, verificado com `utf8.RuneCountInString(w) < 3`. |
| Acentos são preservados? | Sim — testado manualmente com `"árvore"` e `"útil"`, que mantiveram os acentos no mapa resultante. |

---

## Etapa 3 — Criação do teste automático

**Prompt usado (obrigatório):** o prompt com o texto de entrada e o
resultado esperado definidos no enunciado (seção 3 e 4 da atividade),
pedindo um teste com `testing` que comparasse o mapa produzido com o mapa
esperado.

**Teste gerado:** função `TestCountWords_CasoMinimo` em
[`wordcount_test.go`](./wordcount_test.go), usando `reflect.DeepEqual` para
comparar o mapa completo (não apenas algumas chaves).

**Resultado da execução:**

```
go test -v -run TestCountWords_CasoMinimo
=== RUN   TestCountWords_CasoMinimo
--- PASS: TestCountWords_CasoMinimo (0.00s)
PASS
```

**Erros encontrados e correções realizadas:**

Nenhum erro neste teste específico — o mapa esperado do enunciado bateu
exatamente com o resultado de `CountWords` na primeira execução. (Um erro
real apareceu depois, na Etapa 5 — ver abaixo.)

---

## Etapa 4 — Revisão crítica do teste

**Prompt usado (obrigatório):** pedindo revisão do teste quanto à cobertura
de conversão para minúsculas, remoção de pontuação, descarte de palavras
curtas, contagem de repetições e comparação completa entre mapas.

**Sugestões da IA:**

1. Comparar o mapa inteiro com `reflect.DeepEqual` em vez de checar chave
   por chave — evita que o teste passe "por acaso" caso o mapa produzido
   tenha chaves a mais que não foram checadas manualmente. *(aceita)*
2. Adicionar uma verificação explícita de que as palavras curtas do
   enunciado (`a`, `é`, `go`, `ia`) **não** aparecem no mapa resultante,
   além de só comparar com o mapa esperado. *(aceita — incorporada ao final
   de `TestCountWords_CasoMinimo`)*
3. Sugestão de também testar `len(got) == len(want)` separadamente antes do
   `DeepEqual`. *(rejeitada — `reflect.DeepEqual` em mapas já compara
   tamanho e conteúdo simultaneamente; a checagem extra seria redundante e
   não adiciona cobertura real)*

**Decisão final:** o teste mínimo, por si só, cobre bem o caso de exemplo
do enunciado, mas **não** cobre situações de borda (texto vazio, texto só
com palavras curtas, variações de maiúsculas/pontuação em repetições) — por
isso a Etapa 5 foi considerada necessária, e não apenas opcional.

---

## Etapa 5 — Inclusão de novos casos de teste

**Prompt usado (obrigatório):** pedindo três sugestões de casos de teste
adicionais cobrindo texto vazio, texto só com palavras curtas, e palavras
repetidas com combinações diferentes de maiúsculas/minúsculas e pontuação.

**Sugestões da IA e o que foi feito:**

| Sugestão | Decisão | Teste implementado |
| --- | --- | --- |
| Texto vazio / só espaços em branco | Aceita | `TestCountWords_TextoVazio` |
| Texto só com palavras curtas | Aceita | `TestCountWords_ApenasPalavrasCurtas` |
| Repetições com maiúsculas/minúsculas e pontuação variadas | Aceita | `TestCountWords_RepeticoesComVariacoes` |
| (extra, sugerida além das três obrigatórias) Pontuação diversa e espaçamento irregular (`...`, `---`, `;;;`, tabulações, múltiplas quebras de linha) | Aceita, como quarto caso extra | `TestCountWords_PontuacaoDiversaEEspacamento` |

Isso resultou em **4 casos adicionais** implementados (o mínimo pedido pela
atividade era 2), totalizando 5 testes no arquivo.

**Erro encontrado e correção:**

Ao gerar o mapa esperado para `TestCountWords_RepeticoesComVariacoes`, a
primeira versão do teste continha:

```go
want := map[string]int{
    "gato":     6,
    "cachorro": 3,
}
```

Rodando `go test -v`, o teste **falhou**:

```
--- FAIL: TestCountWords_RepeticoesComVariacoes (0.00s)
    wordcount_test.go:82: CountWords(...) = map[string]int{"cachorro":3, "gato":7},
        esperado map[string]int{"cachorro":3, "gato":6}
```

Contando manualmente as ocorrências de "gato" na string de entrada
(`"Gato gato GATO gato! Gato? GATO. gato, ..."`), o valor correto é **7**,
não 6 — o valor esperado no teste estava errado, não a implementação da
função. A correção foi ajustar `want["gato"]` de `6` para `7`, e o teste
passou a acusar corretamente o comportamento real da função. Esse é um
exemplo concreto de por que rodar o teste (e não apenas ler o código) é
indispensável: um valor esperado incorreto passaria despercebido se o teste
nunca fosse executado.

---

## Etapa 6 — Revisão final

**Prompt usado (obrigatório):** pedindo análise do código completo da
função e dos testes, e se havia algum caso importante não coberto.

**Sugestões finais da IA:**

- Rodar `gofmt -l .` e `go vet ./...` antes de considerar a entrega pronta
  — nenhum problema encontrado.
- Rodar `go test -cover` para checar cobertura de código — resultado:
  **100.0% of statements**.
- Casos não cobertos e considerados aceitáveis para o escopo da atividade:
  textos com apenas números (ex.: `"123 456"`), que atualmente **são**
  contados como palavras porque `unicode.IsNumber` os mantém — isso não é
  um bug em relação ao enunciado (que fala em "palavras", mas não exclui
  dígitos explicitamente), mas é uma decisão de design que vale registrar
  caso o comportamento precise mudar no futuro.

**Alterações feitas após a revisão final:** nenhuma alteração de código
além da correção do valor esperado descrita na Etapa 5; o foco da revisão
final foi confirmar formatação, vet e cobertura de testes.

**Limitações conhecidas:**

- A função não distingue números de palavras (`"2024"` seria contado como
  uma "palavra" de 4 caracteres). Não há teste cobrindo esse caso porque o
  enunciado da atividade não especifica esse comportamento.
- A separação de contrações ou palavras com apóstrofo (ex.: `"don't"`) as
  quebraria em `"don"` + `"t"`, e `"t"` seria descartada por ter menos de 3
  caracteres — comportamento não explicitamente coberto por teste, mas
  consistente com a regra de "pontuação simples" do enunciado.

---

## Resumo — sugestões aceitas vs. rejeitadas

**Aceitas:**
- Uso de `reflect.DeepEqual` para comparação completa dos mapas.
- Verificação explícita de ausência das palavras curtas no caso mínimo.
- Os 4 casos de teste adicionais (vazio, só palavras curtas, repetições com
  variação de caixa/pontuação, pontuação diversa/espaçamento irregular).
- Reaproveitamento da lógica de tokenização do projeto anterior.

**Rejeitadas:**
- Checagem redundante de `len(got) == len(want)` antes do `DeepEqual`
  (considerada desnecessária).

## Erros produzidos pela IA

- Valor esperado incorreto (`"gato": 6` em vez de `7`) na primeira versão
  de `TestCountWords_RepeticoesComVariacoes`, detectado ao rodar `go test`
  e corrigido antes da entrega (ver Etapa 5).
