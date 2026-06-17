[README.md](https://github.com/user-attachments/files/29055457/README.md)
# Conta-palavras Concorrente em Go

Programa em Go que conta a frequência de palavras de um arquivo de texto, com
duas implementações — **sequencial** e **concorrente** — usadas para comparar
corretude e desempenho entre as duas abordagens.

Dataset usado nos testes: *A Christmas Carol*, de Charles Dickens
(`AChristmasCarol_CharlesDickens_English.txt`).

## Como executar

Requer Go 1.22 ou superior.

```bash
# clonar o repositório e entrar na pasta
git clone <URL-DO-REPOSITORIO>
cd <pasta-do-repositorio>

# executar com número de workers padrão (= número de CPUs lógicas da máquina)
go run . AChristmasCarol_CharlesDickens_English.txt

# executar informando explicitamente o número de workers da versão concorrente
go run . AChristmasCarol_CharlesDickens_English.txt 4
```

Saída típica:

```
Arquivo: AChristmasCarol_CharlesDickens_English.txt
Total de palavras após filtragem: 23445
Workers (versão concorrente): 4

Tempo sequencial: 1.55ms
Tempo concorrente: 1.83ms
Resultados iguais: sim
A versão concorrente foi 1.18x mais lenta que a sequencial

Top 20 palavras:
 1. the             1629
 2. and             1082
 ...
```

## Estrutura do código

| Arquivo         | Responsabilidade                                                   |
|-----------------|----------------------------------------------------------------------|
| `tokenize.go`   | Normalização do texto: minúsculas, remoção de pontuação simples e filtro de palavras com menos de 3 caracteres. Usada igualmente pelas duas versões, para que ambas partam exatamente da mesma lista de palavras. |
| `sequential.go` | Versão **sequencial** da contagem (`sequentialCount`).               |
| `concurrent.go` | Versão **concorrente** da contagem (`concurrentCount`).              |
| `compare.go`    | Comparação dos dois mapas de frequência (`mapsEqual`) e seleção do top N (`topN`). |
| `main.go`       | Orquestra a leitura do arquivo, tokenização, execução das duas versões, medição de tempo e impressão dos resultados. |

## Versão sequencial

`sequentialCount` percorre a lista de palavras com um único laço `for`,
incrementando um `map[string]int`. Não usa goroutines, channels, nem qualquer
outro recurso de concorrência. É a referência de corretude para a versão
concorrente.

## Estratégia concorrente

A versão concorrente usa o padrão **fan-out / fan-in** (também chamado de
*map-reduce* local):

1. **Divisão do trabalho (fan-out):** a fatia de palavras já tokenizadas é
   dividida em `numWorkers` blocos contíguos de tamanho aproximadamente igual
   (`chunkSize = ceil(n / numWorkers)`). `numWorkers` é o número de CPUs
   lógicas da máquina por padrão, ou pode ser informado como segundo
   argumento na linha de comando.
2. **Processamento paralelo:** uma goroutine é disparada para cada bloco.
   Cada goroutine monta o **seu próprio mapa local** de contagens,
   percorrendo apenas o seu bloco de palavras. Como cada goroutine só
   escreve no seu mapa local (nunca em uma estrutura compartilhada), não há
   necessidade de `sync.Mutex` nessa etapa — não existe memória compartilhada
   sendo escrita por mais de uma goroutine ao mesmo tempo.
3. **Sincronização:** um `sync.WaitGroup` garante que a goroutine principal
   espere todas as goroutines de contagem terminarem antes de seguir.
4. **Combinação dos resultados (fan-in / reduce):** cada goroutine envia seu
   mapa local para um `channel` (`chan map[string]int`). Depois que
   `wg.Wait()` retorna e o channel é fechado, a goroutine principal lê todos
   os mapas parciais do channel, um por vez, e os soma em um único mapa
   final (`final[word] += count`). Essa etapa de combinação é feita de forma
   **sequencial, por uma única goroutine**, depois que todo o trabalho
   paralelo já terminou — por isso não há corrida de dados (*data race*) ao
   escrever no mapa final, mesmo sem mutex.

Esse design foi confirmado livre de *data races* executando o programa com a
flag `-race` do Go (`go run -race . arquivo.txt`).

## Como a correção foi verificada

A função `mapsEqual` compara o **mapa completo** de frequências produzido
pela versão sequencial com o mapa completo produzido pela versão concorrente
— não apenas o top 20. A comparação verifica:

- que os dois mapas têm o mesmo número de palavras (chaves);
- que cada palavra tem exatamente a mesma contagem nos dois mapas.

O programa imprime `Resultados iguais: sim` ou `Resultados iguais: não` com
base nesse resultado. Em todas as execuções realizadas durante o
desenvolvimento (com 1, 4 e 8 workers, e também com um arquivo sintético ~50×
maior que o dataset original), o resultado foi sempre `sim`.

## Como o tempo de execução foi medido

A medição usa `time.Now()` / `time.Since()` envolvendo **apenas** a chamada à
função de contagem de cada versão (`sequentialCount` e `concurrentCount`),
de forma equivalente para as duas:

```go
startSeq := time.Now()
seqCounts := sequentialCount(words)
seqDuration := time.Since(startSeq)

startConc := time.Now()
concCounts := concurrentCount(words, numWorkers)
concDuration := time.Since(startConc)
```

A leitura do arquivo e a tokenização (que são idênticas para as duas
versões) ficam **fora** da medição, para que a comparação reflita apenas o
custo da etapa de contagem em si.

## Análise de desempenho

Os testes abaixo foram executados em uma máquina/sandbox com **apenas 1 CPU
lógica disponível** (`nproc` = 1). Nesse cenário, a versão concorrente não
tem hardware para paralelismo real — todas as goroutines competem pelo mesmo
núcleo — então o que se observa é principalmente o **custo de overhead** de
criar goroutines, channels e fazer a etapa de *reduce*:

| Workers | Tempo sequencial | Tempo concorrente | Resultado            |
|---------|-------------------|---------------------|------------------------|
| 1       | ~1.55 ms          | ~1.42–1.83 ms        | levemente mais lenta ou equivalente |
| 4       | ~1.55 ms          | ~1.83 ms             | ~1.18x mais lenta      |
| 8       | ~1.55 ms          | ~1.91 ms             | ~1.23x mais lenta      |

Para o arquivo `AChristmasCarol_CharlesDickens_English.txt` (≈23 mil palavras
após filtragem), a versão concorrente foi, na máquina de teste, **equivalente
ou um pouco mais lenta** que a sequencial: o arquivo é pequeno o suficiente
para que o custo de criar goroutines e channels supere o ganho de dividir o
trabalho, e não há núcleos extras disponíveis para paralelismo real.

Repetindo o teste com um arquivo sintético ~50× maior (≈1,17 milhão de
palavras), a versão concorrente chegou a ser **mais rápida** em algumas
execuções (ex.: ~95 ms sequencial vs. ~47 ms concorrente), mas os resultados
variaram entre execuções — efeito esperado em uma máquina de 1 núcleo, onde
não há paralelismo verdadeiro e a comparação fica sensível a fatores como
coleta de lixo (GC) e agendamento do sistema operacional.

**Conclusão:** em uma máquina com múltiplos núcleos reais, espera-se que a
versão concorrente seja consistentemente mais rápida para arquivos grandes
(o trabalho de contagem é dividido entre núcleos), mas para arquivos
pequenos o overhead de goroutines/channels pode anular ou até superar esse
ganho. Recomenda-se executar o programa na própria máquina e comparar os
tempos obtidos, especialmente com arquivos maiores, para observar o
speedup real do paralelismo.

## Dificuldades encontradas

- **Arquivo com BOM (Byte Order Mark):** o dataset começa com o caractere
  invisível `\uFEFF` (UTF-8 BOM), que precisou ser removido explicitamente
  no início da tokenização para não virar uma "palavra" espúria.
- **Definição de "pontuação simples":** optou-se por tratar como separador
  qualquer caractere que não seja letra ou número (`unicode.IsLetter` /
  `unicode.IsNumber`), via `strings.FieldsFunc`. Isso tem o efeito colateral
  de separar contrações como `don't` em `don` + `t` — mas como `t` tem menos
  de 3 caracteres, é descartada pelo filtro de tamanho mínimo, então o
  impacto prático é pequeno.
- **Determinismo do top 20:** como a iteração sobre `map` em Go não tem ordem
  garantida, foi necessário ordenar explicitamente por frequência
  (decrescente) e, em caso de empate, por ordem alfabética, para que a saída
  seja sempre a mesma entre execuções.
- **Avaliar corretude da combinação concorrente sem mutex:** garantir que a
  etapa de *merge* dos mapas parciais fosse realmente livre de *data race*
  (mesmo sem mutex) foi verificado executando o programa com
  `go run -race .`, que não reportou nenhuma corrida de dados.
- **Demonstrar ganho de desempenho real:** a máquina usada para os testes
  durante o desenvolvimento tinha apenas 1 CPU lógica, o que limita o
  paralelismo real e torna os números de speedup pouco representativos de
  uma máquina multi-core — esse ponto está documentado na seção de análise
  acima.
