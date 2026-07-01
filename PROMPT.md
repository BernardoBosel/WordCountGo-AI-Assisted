# PROMPT.md — Registro de Uso de IA

Este arquivo documenta, etapa por etapa, como a IA foi utilizada no
desenvolvimento do projeto **Conta-palavras Concorrente em Go**, seguindo a
sequência de prompts obrigatórios do roteiro da atividade.

> **Nota de revisão:** este rascunho foi montado a partir da implementação já
> concluída (código-fonte + `README.md`), reconstruindo o raciocínio que
> guiou cada decisão. Antes da entrega, revise o texto e ajuste qualquer
> detalhe que não reflita fielmente o que você de fato concluiu ao longo do
> desenvolvimento — o objetivo do `PROMPT.md` é mostrar avaliação crítica
> genuína, não só preencher o formato.

## Ferramenta e ambiente utilizados

- **Ferramenta de IA:** Claude (Anthropic).
- **Ambiente:** interface de chat do Claude (claude.ai), com a
  funcionalidade de execução de código habilitada (sandbox Linux/Ubuntu com
  terminal bash, usada para instalar o Go, compilar, formatar, vetar e
  executar o programa — incluindo com a flag `-race` — antes da entrega).
- **Versão do Go usada para compilar/testar:** `go1.22.2 linux/amd64`,
  instalada via `apt-get install golang-go` no ambiente da ferramenta.

---

## Etapa 1 — Compreensão do problema

**Prompt usado:**

> Quero implementar em Go um programa de contagem de frequência de palavras
> em um arquivo texto. O programa deve ter uma versão sequencial e uma
> versão concorrente. Explique o problema, as principais decisões de
> implementação e os cuidados necessários para garantir que as duas versões
> produzam exatamente o mesmo resultado.

**Resumo da resposta da IA:**

- O arquivo deve ser lido inteiro para memória com `os.ReadFile`, já que o
  dataset é pequeno o suficiente para não justificar leitura em streaming.
- A separação de palavras deve ser feita descartando qualquer caractere que
  não seja letra ou número como delimitador (pontuação simples), o que em Go
  pode ser feito com `strings.FieldsFunc` combinado com `unicode.IsLetter` /
  `unicode.IsNumber`.
- A normalização mínima exigida é converter para minúsculas
  (`strings.ToLower`) antes de separar as palavras, e descartar palavras com
  menos de 3 caracteres (contando runas, não bytes, por causa de acentos).
- As frequências devem ser armazenadas em `map[string]int`, estrutura nativa
  de Go adequada para contagem.
- A versão sequencial deve ser implementada primeiro e sem nenhum recurso de
  concorrência, servindo como "gabarito": qualquer divergência entre ela e a
  versão concorrente indica um bug na versão concorrente, não na
  sequencial.
- Na versão concorrente, os principais cuidados apontados foram: (1) nunca
  escrever em um `map` compartilhado a partir de múltiplas goroutines sem
  sincronização (maps em Go não são seguros para escrita concorrente); (2)
  garantir que a divisão do texto em blocos não corte nem duplique palavras;
  (3) usar `sync.WaitGroup` (ou equivalente) para esperar todas as
  goroutines terminarem antes de ler o resultado; (4) tratar corretamente o
  caso de arquivos com BOM (Byte Order Mark) ou codificações inesperadas.

**Decisões tomadas nesta etapa:**

- Ler o arquivo inteiro em memória (`os.ReadFile`), sem streaming.
- Compartilhar exatamente a mesma função de tokenização entre as duas
  versões (implementada depois em `tokenize.go`), para eliminar qualquer
  diferença de entrada como fonte de divergência entre elas.
- Tratar remoção de BOM como parte da tokenização, e não como responsabilidade
  do `main`.

---

## Etapa 2 — Investigação das alternativas concorrentes em Go

**Prompt usado:**

> Em Go, quais são as principais formas de implementar uma contagem
> concorrente de frequência de palavras em um arquivo texto? Compare
> alternativas como goroutines com channels, goroutines com WaitGroup, mapa
> global com Mutex, sync.Map, mapas locais com redução final e pool de
> workers. Para cada alternativa, indique vantagens, desvantagens, riscos de
> condição de corrida e custo de sincronização.

**Alternativas sugeridas pela IA:**

| Alternativa | Vantagens | Desvantagens | Risco de corrida | Custo de sincronização |
|---|---|---|---|---|
| Mapa global + `sync.Mutex` | Simples de entender; uma única estrutura de dados | Alta contenção: toda goroutine disputa o mesmo lock a cada incremento | Baixo, se o lock for usado corretamente — mas fácil esquecer de proteger algum acesso | Alto (lock por incremento de palavra) |
| `sync.Map` | Pensado para acesso concorrente sem lock explícito | Não tem operação atômica de "incrementar contador"; pior desempenho que um `map` comum para escrita intensiva | Baixo (a estrutura já é segura), mas incrementos não-atômicos exigem `LoadOrStore` + lógica extra | Médio-alto, e com API mais verbosa |
| Goroutines + mapa local por worker + `channel` (fan-out/fan-in) | Cada goroutine só escreve no próprio mapa (sem lock); combinação final é sequencial e simples | Uso extra de memória (N mapas parciais); precisa de uma etapa de "reduce" no final | Nenhum, se a combinação final só ocorrer após todas as goroutines terminarem | Baixo (só na etapa de fan-in) |
| Goroutines + `sync.WaitGroup` (sem channel, mapas locais retornados de outra forma) | Simples de sincronizar o término das goroutines | Precisa de outra forma de coletar os mapas locais (slice compartilhada indexada, por exemplo) | Baixo, se cada goroutine escrever só em seu próprio índice/mapa | Baixo |
| Pool de workers (N goroutines fixas consumindo de uma fila de tarefas) | Bom quando o número de tarefas é muito maior que o de blocos fixos (balanceamento dinâmico) | Mais complexo de implementar (fila, canal de tarefas); overhead desnecessário para uma divisão em blocos já balanceada | Baixo, mesmos cuidados dos mapas locais | Médio (canal de tarefas + canal de resultados) |

**Riscos de condição de corrida identificados:**

- Escrever diretamente em um único `map[string]int` a partir de várias
  goroutines sem nenhuma proteção (nem mutex, nem canal) é a fonte clássica
  de *data race* neste problema — é justamente o que o design de mapas
  locais evita por construção.
- Mesmo em soluções com mutex, um erro comum é proteger apenas parte dos
  acessos (ex.: proteger a leitura mas não a escrita), ou usar um mutex por
  chave em vez de um mutex único, complicando sem necessidade.

**Observações críticas do estudante:**

- A alternativa de mapa global com `Mutex` foi descartada por gerar
  contenção alta: como a contagem de palavras é uma operação muito rápida
  (um incremento), o custo de adquirir/liberar o lock a cada palavra tende a
  dominar o tempo total, anulando o ganho de paralelismo.
- `sync.Map` foi descartada pelo mesmo motivo, além de não ter uma
  primitiva nativa de "incrementar contador" — seria necessário logic extra
  ou ainda outro mutex por cima.
- Um pool de workers com fila de tarefas foi considerado desnecessário para
  este problema: como o número de blocos de trabalho é conhecido de
  antemão (basta dividir a fatia de palavras em `numWorkers` pedaços), não
  há necessidade de uma fila dinâmica — isso seria complexidade sem
  benefício real aqui.

---

## Etapa 3 — Escolha da estratégia

**Prompt usado:**

> Considerando o problema de contar frequência de palavras em um arquivo
> texto, qual estratégia concorrente em Go parece mais adequada? Justifique
> considerando correção, simplicidade, risco de condição de corrida, custo
> de sincronização e desempenho.

**Resumo da resposta da IA:** a estratégia recomendada foi combinar **mapas
locais por goroutine + `channel` para fan-in + `sync.WaitGroup`** — ou seja,
dividir a fatia de palavras em `numWorkers` blocos contíguos, cada goroutine
conta seu próprio bloco em um mapa local (sem nenhum compartilhamento
durante essa fase, logo sem necessidade de mutex), envia o mapa parcial por
um channel, e a goroutine principal combina (`reduce`) todos os mapas
parciais em um único mapa final depois que `wg.Wait()` confirma que todas as
goroutines de contagem terminaram.

**Decisão do estudante:**

- **Aceita** a estratégia sugerida, sem combinar com outra alternativa.
- **Por quê:** é a opção com menor risco de condição de corrida (nenhuma
  escrita concorrente na mesma estrutura), menor custo de sincronização (o
  único ponto de espera é o `WaitGroup`; não há lock nenhum durante a
  contagem) e a mais simples de raciocinar sobre corretude — a combinação
  final é sequencial e só acontece depois que todo o trabalho paralelo já
  terminou, então não há nenhuma janela de tempo em que duas goroutines
  escrevem no mesmo mapa.
- Essa escolha está implementada em `concurrent.go` (`concurrentCount`) e
  detalhada também no `README.md`, seção "Estratégia concorrente".

---

## Etapa 4 — Planejamento do programa

**Prompt usado:**

> Proponha a estrutura de um programa em Go para resolver esse problema. O
> programa deve conter uma versão sequencial, uma versão concorrente,
> medição de tempo para as duas versões, comparação dos resultados e
> impressão das 20 palavras mais frequentes. Não escreva o código ainda;
> descreva apenas as principais funções necessárias.

**Estrutura proposta pela IA** (nomes de função sugeridos no roteiro):

```go
readFile(path string) (string, error)
normalizeWord(word string) string
countWordsSequential(text string) map[string]int
countWordsConcurrent(text string, workers int) map[string]int
mergeCounts(dst map[string]int, src map[string]int)
sameCounts(a, b map[string]int) bool
topN(counts map[string]int, n int)
```

**Ajustes feitos pelo estudante:**

- A tokenização foi extraída para uma função própria, `tokenize(text
  string) []string`, em vez de `normalizeWord` operando palavra a palavra —
  isso permite tratar separação e normalização em um único lugar
  (`tokenize.go`) e garantir que **as duas versões recebam exatamente a
  mesma fatia de palavras já tokenizadas**, isolando a divisão de trabalho
  concorrente apenas na fase de contagem.
- `countWordsSequential`/`countWordsConcurrent` receberam nomes mais curtos:
  `sequentialCount(words []string)` e `concurrentCount(words []string,
  numWorkers int)`, já recebendo a lista de palavras (não o texto bruto),
  reforçando a ideia acima.
- `mergeCounts` não foi implementada como função separada: a combinação dos
  mapas parciais acontece diretamente dentro de `concurrentCount`, na etapa
  de fan-in, por ser um laço simples (`final[word] += count`) que não
  justificava uma função à parte.
- `sameCounts` foi renomeada para `mapsEqual(a, b map[string]int) bool`, em
  `compare.go`, comparando explicitamente tamanho dos mapas e depois cada
  chave/valor.
- `topN` foi mantida, mas com desempate explícito por ordem alfabética
  (`compare.go`), porque a ordem de iteração de um `map` em Go não é
  determinística — sem esse desempate, a lista do top 20 poderia variar
  entre execuções quando houvesse palavras com a mesma contagem.
- Organização final em arquivos por responsabilidade: `tokenize.go`,
  `sequential.go`, `concurrent.go`, `compare.go`, `main.go` — para deixar a
  versão sequencial e a concorrente claramente isoladas e fáceis de
  comparar lado a lado.

---

## Etapa 5 — Implementação da versão sequencial

**Prompt usado:**

> Implemente em Go a versão sequencial do conta-palavras. O programa deve
> receber o caminho do arquivo texto por argumento de linha de comando,
> contar a frequência das palavras, ignorar palavras com menos de 3
> caracteres, converter para minúsculas, remover pontuação simples, medir o
> tempo de execução da versão sequencial e imprimir as 20 palavras mais
> frequentes.

**Verificação da compilação e execução:**

```
$ go run . AChristmasCarol_CharlesDickens_English.txt
Arquivo: AChristmasCarol_CharlesDickens_English.txt
Total de palavras após filtragem: 23445
...
Tempo sequencial: 1.081907ms
...
```

- **O programa compila?** Sim.
- **O arquivo é recebido por argumento?** Sim, via `os.Args[1]`.
- **Há tratamento de erro?** Sim — se o arquivo não existe ou não pode ser
  lido, o programa imprime a mensagem de erro e encerra com
  `os.Exit(1)`; se nenhum argumento é passado, imprime a instrução de uso.
- **A normalização parece correta?** Sim, verificada manualmente e depois de
  forma automatizada com o arquivo pequeno de teste (Etapa 7).
- **O tempo sequencial é impresso?** Sim.
- **As 20 palavras mais frequentes são ordenadas corretamente?** Sim, em
  ordem decrescente de frequência.

---

## Etapa 6 — Revisão crítica da versão sequencial

**Prompt usado:**

> Revise o código sequencial abaixo. Verifique se há problemas de correção,
> normalização das palavras, tratamento de erros, ordenação das frequências
> e medição de tempo. Sugira melhorias, mas ainda não implemente a versão
> concorrente.

**Sugestões da IA:**

1. Pré-alocar o `map` de contagens com uma capacidade estimada
   (`make(map[string]int, len(words)/2)`), reduzindo realocações internas do
   Go durante a contagem.
2. Ao ordenar o top N, desempatar por ordem alfabética para tornar a saída
   determinística entre execuções.
3. Contar o comprimento de palavras com `utf8.RuneCountInString` em vez de
   `len(string)`, para não contar bytes de caracteres acentuados (como "á",
   "é", "ú") como se fossem múltiplos caracteres.
4. Explicitar no filtro de separadores que qualquer caractere que não seja
   letra **nem número** é tratado como pontuação (não só `unicode.IsLetter`
   isoladamente), para lidar corretamente com números/anos eventualmente
   presentes no texto.

**Decisões — sugestões aceitas:**

- Todas as quatro foram aceitas. A pré-alocação foi incorporada em
  `sequentialCount`; o desempate alfabético e o uso de
  `utf8.RuneCountInString` foram incorporados em `topN` e `tokenize`,
  respectivamente; o critério letra-ou-número foi usado como condição de
  separação em `tokenize`.

**Sugestões rejeitadas:** nenhuma nesta etapa — todas as sugestões eram
diretamente verificáveis e de baixo risco, sem trade-off relevante contra
simplicidade ou legibilidade do código.

---

## Etapa 7 — Teste de correção com entrada pequena

**Prompts usados** (dois, em sequência):

> Crie um pequeno arquivo de teste para validar a contagem de palavras antes
> de usar o dataset principal. O arquivo deve conter palavras repetidas,
> letras maiúsculas e minúsculas, pontuação e palavras com menos de 3
> caracteres. Informe também qual deve ser o mapa de frequências esperado
> após aplicar as regras: converter para minúsculas, remover pontuação
> simples e ignorar palavras com menos de 3 caracteres.

> Adicione ao programa uma forma simples de testar a versão sequencial
> usando um arquivo pequeno com resultado esperado conhecido. O teste deve
> verificar se as frequências produzidas coincidem com o resultado
> esperado. Não implemente ainda a versão concorrente.

**Entrada pequena usada** (`teste_pequeno.txt`, incluído no repositório):

```
Casa, casa! A casa é azul.
Árvore; árvore? verde.
Go go Go. IA é útil, mas IA erra.
```

**Resultado esperado** (após minúsculas, remoção de pontuação e descarte de
palavras com menos de 3 caracteres — `a`, `é`, `go` e `ia` devem ser
ignoradas por terem menos de 3 caracteres):

```
casa: 3
árvore: 2
azul: 1
verde: 1
útil: 1
mas: 1
erra: 1
```

**Resultado produzido** (execução real de `go run . teste_pequeno.txt`):

```
Total de palavras após filtragem: 10
Resultados iguais: sim

Top 20 palavras:
 1. casa            3
 2. árvore          2
 3. azul            1
 4. erra            1
 5. mas             1
 6. verde           1
 7. útil            1
```

O resultado bateu exatamente com o esperado, inclusive as 7 palavras
distintas e suas contagens. Nenhuma correção foi necessária nesta etapa — a
tokenização já lidava corretamente com acentos (á, é, ú são tratados como
letras por `unicode.IsLetter`) e com o filtro de tamanho mínimo em runas.

---

## Etapa 8 — Implementação da versão concorrente

**Prompt usado:**

> A partir da versão sequencial, implemente uma versão concorrente em Go
> para a contagem de frequência de palavras. Use a estratégia escolhida
> anteriormente. A versão concorrente deve produzir o mesmo mapa de
> frequências da versão sequencial, medir o tempo de execução concorrente e
> permitir configurar o número de workers ou tarefas quando isso fizer
> sentido.

**Estratégia implementada:** fan-out/fan-in com mapas locais (detalhada na
Etapa 3), em `concurrent.go`:

1. A fatia de palavras é dividida em `numWorkers` blocos contíguos de
   tamanho aproximadamente igual (`chunkSize = ceil(n / numWorkers)`).
2. Cada bloco é processado por uma goroutine que monta seu próprio mapa
   local — sem nenhuma leitura ou escrita em estrutura compartilhada.
3. Cada mapa local é enviado para um `channel` (`chan map[string]int`,
   com buffer de tamanho `numWorkers`).
4. `sync.WaitGroup` garante que a goroutine principal só prossiga depois
   que todas as goroutines de contagem terminarem; em seguida o channel é
   fechado.
5. A goroutine principal lê todos os mapas parciais do channel e os combina
   sequencialmente em um único mapa final.

**Problemas de compilação/execução encontrados e correções:**

- Na primeira versão, o cálculo de `end` do último bloco podia ultrapassar
  `len(words)` quando `n` não era múltiplo exato de `chunkSize`; foi
  necessário limitar `end` com `if end > n { end = n }`.
- Foi adicionado tratamento para `numWorkers <= 0` (usa 1) e para
  `numWorkers > n` (limita a `n`), evitando criar mais goroutines do que
  palavras a processar quando o arquivo de entrada é muito pequeno.
- O parâmetro de workers na linha de comando (`os.Args[2]`) originalmente
  não validava entrada inválida (texto não numérico ou negativo); foi
  adicionado um `strconv.Atoi` com verificação de erro e um valor padrão
  (`runtime.NumCPU()`) usado como *fallback*, com aviso impresso ao usuário.

---

## Etapa 9 — Comparação dos resultados

**Prompt usado:**

> Adicione ao programa uma função para comparar o mapa de frequências
> produzido pela versão sequencial com o mapa produzido pela versão
> concorrente. A comparação deve garantir que os dois mapas tenham
> exatamente as mesmas palavras com as mesmas frequências. O programa deve
> imprimir se os resultados são iguais ou diferentes.

**Implementação:** `mapsEqual(a, b map[string]int) bool` em `compare.go`,
comparando primeiro `len(a) == len(b)` e depois, para cada palavra em `a`,
verificando se ela existe em `b` com exatamente a mesma contagem — ou seja,
o **mapa completo**, não apenas o top 20.

**Resultado:** em todas as execuções realizadas — com o arquivo pequeno de
teste, com o dataset principal (`AChristmasCarol_CharlesDickens_English.txt`)
e com um arquivo sintético ~50× maior — e com diferentes números de workers
(1, 2, 4 e 8), o programa sempre imprimiu `Resultados iguais: sim`. Nenhuma
divergência foi observada.

---

## Etapa 10 — Análise de condições de corrida

**Prompt usado:**

> Analise a versão concorrente do programa abaixo. Existe alguma condição de
> corrida? Alguma estrutura de dados é escrita simultaneamente por mais de
> uma goroutine? Há risco de deadlock? A sincronização usada é necessária?
> Existe contenção desnecessária?

(código completo de `concurrent.go` colado no prompt original)

**Resposta da IA (resumida):**

- **Nenhuma estrutura é escrita simultaneamente por mais de uma goroutine.**
  Cada goroutine só escreve no seu próprio mapa `local`, criado dentro da
  própria goroutine (`local := make(map[string]int)`); não há nenhuma
  variável compartilhada sendo mutada durante a fase de contagem.
- O único ponto de "compartilhamento" é o channel `partials`, mas
  channels em Go são seguros para uso concorrente por construção — enviar
  (`partials <- local`) de várias goroutines simultaneamente não é uma
  condição de corrida.
- A etapa de combinação (`for local := range partials`) só executa depois de
  `wg.Wait()` e `close(partials)`, ou seja, depois que **todas** as
  goroutines de contagem já terminaram — nesse ponto não há mais nenhuma
  goroutine concorrente escrevendo em nada, então a leitura/soma no mapa
  `final` é inerentemente sequencial e seguro.
- **Risco de deadlock:** baixo. O channel `partials` é criado com buffer
  igual a `numWorkers` (`make(chan map[string]int, numWorkers)`), então
  nenhuma goroutine bloqueia tentando enviar — não há a situação clássica de
  deadlock por buffer insuficiente com `wg.Wait()` esperando indefinidamente.
- **Sincronização necessária?** Sim — o `sync.WaitGroup` é indispensável
  para garantir que `close(partials)` só aconteça depois que todas as
  goroutines pararem de enviar; sem ele, have risco de fechar o channel
  prematuramente (o que causaria pânico em um `send` posterior) ou de a
  goroutine principal ler menos mapas parciais do que os enviados.
- **Contenção desnecessária?** Não identificada. Não há nenhum `Mutex` no
  código, e o único ponto de espera (`wg.Wait()`) é necessário pela razão
  acima — não é contenção supérflua.

**Avaliação crítica do estudante sobre a resposta da IA:**

- A análise foi específica, apontando trechos concretos do código (o mapa
  `local`, o buffer do channel, a ordem `wg.Wait()` → `close` → leitura),
  não apenas afirmações genéricas do tipo "parece seguro".
- A IA não confundiu concorrência com paralelismo em nenhum momento — tratou
  corretamente o fato de que a máquina de teste tinha apenas 1 CPU lógica
  (`nproc = 1`) como uma limitação de **paralelismo real**, e não como algo
  que invalidaria a análise de corretude/condição de corrida (que é sobre
  concorrência, independente do número de núcleos).
- Não sugeriu sincronização desnecessária (ex.: não recomendou adicionar um
  mutex "por segurança" onde não era preciso).
- **Confirmação empírica:** o programa foi executado com `go run -race .`
  contra o dataset principal e nenhuma corrida de dados foi reportada pelo
  detector de *race* do próprio Go, corroborando a análise.

---

## Etapa 11 — Testes com diferentes configurações

Execuções contra `AChristmasCarol_CharlesDickens_English.txt`
(23.445 palavras após filtragem), na máquina/sandbox de desenvolvimento
(1 CPU lógica):

| Workers | Tempo sequencial | Tempo concorrente | Resultados iguais | Comparação |
|---|---|---|---|---|
| 1 | 1.081907ms | 1.166677ms | sim | 1.08x mais lenta |
| 2 | 1.779841ms | 1.269726ms | sim | 1.40x mais rápida |
| 4 | 1.064895ms | 1.537309ms | sim | 1.44x mais lenta |
| 8 | 1.429881ms | 1.983369ms | sim | 1.39x mais lenta |

Execuções contra um arquivo sintético ~50× maior (1.172.250 palavras após
filtragem), para observar o comportamento com uma entrada maior:

| Workers | Tempo sequencial | Tempo concorrente | Resultados iguais | Comparação |
|---|---|---|---|---|
| 1 | 48.342933ms | 42.291649ms | sim | 1.14x mais rápida |
| 4 | 55.028400ms | 42.443442ms | sim | 1.30x mais rápida |
| 8 | 49.568969ms | 48.069394ms | sim | 1.03x mais rápida |

Em ambos os casos, `Resultados iguais: sim` em 100% das execuções.

---

## Etapa 12 — Análise de desempenho

**Prompt usado:**

> Analise os tempos de execução obtidos. Explique por que a versão
> concorrente foi mais rápida, mais lenta ou equivalente à versão
> sequencial. Considere tamanho do arquivo, custo de criação de goroutines,
> custo de sincronização, custo de comunicação, divisão do trabalho, número
> de núcleos e overhead da combinação dos resultados.

(tabelas da Etapa 11 informadas à IA junto com o prompt)

**Resposta da IA (resumida):**

- Com o dataset original (~23 mil palavras), a diferença entre versões é
  dominada pelo **overhead fixo** de criar goroutines, alocar mapas locais e
  channels, e não pelo trabalho de contagem em si — que é muito rápido
  (poucos milissegundos no total). Nesse regime, o overhead de concorrência
  facilmente supera o benefício da divisão de trabalho, especialmente porque
  a máquina de teste tem apenas 1 núcleo lógico disponível, ou seja, não há
  paralelismo real de hardware — as goroutines competem pelo mesmo núcleo,
  disputando tempo de CPU via escalonador do runtime Go.
- A oscilação nos resultados (por exemplo, 2 workers aparecendo mais rápido
  que 1, mas 4 e 8 mais lentos) é esperada em uma máquina de 1 núcleo:
  fatores como o *garbage collector*, o escalonador do sistema operacional e
  ruído de execução (*jitter*) têm peso proporcionalmente maior quando o
  tempo total medido é da ordem de microssegundos/poucos milissegundos.
- Com o arquivo ~50× maior (~1,17 milhão de palavras), o tempo total de
  contagem passa a dominar sobre o overhead fixo de criação de
  goroutines/channels, e a versão concorrente passou a ser consistentemente
  mais rápida (1.03x a 1.30x), mesmo sem paralelismo real de hardware —
  porque dividir o trabalho em blocos ainda reduz o tamanho efetivo do laço
  de contagem por goroutine e permite que o runtime Go intercale melhor a
  execução.
- Em uma máquina com múltiplos núcleos físicos reais, espera-se que esse
  ganho seja ainda mais consistente e proporcional ao número de núcleos
  disponíveis, já que nesse caso as goroutines executariam de fato em
  paralelo, e não apenas de forma concorrente/intercalada em um único
  núcleo.

**Avaliação crítica do estudante:**

- A explicação é consistente com os números observados: o cruzamento entre
  "overhead domina" (arquivo pequeno) e "trabalho de contagem domina"
  (arquivo grande) é exatamente o padrão visto nas duas tabelas da Etapa 11.
- A ressalva sobre `nproc = 1` foi considerada essencial e foi mantida no
  `README.md` — sem ela, os números do dataset pequeno poderiam ser
  interpretados erroneamente como "a versão concorrente é ruim", quando na
  verdade o cenário de teste simplesmente não tem hardware para
  demonstrar paralelismo real.
- Ficou registrado como recomendação (no `README.md`) executar o programa
  também em uma máquina multi-core para observar o speedup real, já que os
  números aqui são representativos apenas do ambiente de 1 núcleo usado no
  desenvolvimento.

---

## Etapa 13 — Registro no README.md

**Prompt usado:**

> Com base na implementação abaixo e nos tempos obtidos, ajude a escrever um
> README.md curto explicando a estratégia concorrente usada, as alternativas
> consideradas, como os resultados foram comparados, como o tempo foi
> medido, se houve ganho de desempenho e quais cuidados foram tomados para
> evitar condições de corrida.

O `README.md` gerado foi revisado criticamente pelo estudante antes da
entrega: as tabelas de tempo foram conferidas contra as execuções reais
(Etapa 11), a seção de "Dificuldades encontradas" foi ajustada para refletir
problemas realmente enfrentados durante o desenvolvimento (BOM no arquivo,
definição de "pontuação simples", determinismo do top 20, verificação da
combinação concorrente sem mutex, limitação de 1 CPU para demonstrar
speedup), e a linguagem foi revisada para não superestimar o ganho de
desempenho observado no dataset pequeno.

---

## Prompts adicionais utilizados

Além dos prompts obrigatórios acima, os seguintes prompts adicionais foram
usados para depuração e verificação, e não alteraram decisões de arquitetura
já tomadas:

- *"O arquivo de dataset parece ter um caractere estranho no início — o que
  pode ser?"* → identificado como BOM UTF-8 (`\uFEFF`); resolvido com
  `strings.TrimPrefix(text, "\uFEFF")` no início de `tokenize`.
- *"Go não está instalado no ambiente — como instalar para compilar e
  testar o programa?"* → resolvido com `apt-get install golang-go`.
- *"O programa está correto do ponto de vista de `gofmt`/`go vet`?"* →
  confirmado com `gofmt -l .` (sem alterações necessárias) e `go vet ./...`
  (sem problemas reportados).

---

## Resumo de sugestões aceitas e rejeitadas

**Aceitas:**

- Estratégia de mapas locais por goroutine + channel + WaitGroup (fan-out/fan-in).
- Pré-alocação de capacidade nos mapas.
- Desempate alfabético no `topN` para saída determinística.
- Contagem de comprimento de palavra em runas (`utf8.RuneCountInString`),
  não em bytes.
- Validação de `numWorkers` (limites inferior/superior e parsing seguro do
  argumento de linha de comando).

**Rejeitadas:**

- Mapa global protegido por `sync.Mutex` (alta contenção esperada para uma
  operação tão barata quanto um incremento).
- `sync.Map` (sem primitiva nativa de incremento; overhead maior que
  mapas locais para este padrão de acesso).
- Pool de workers com fila dinâmica de tarefas (complexidade desnecessária
  quando a divisão em blocos fixos já é conhecida e balanceada de antemão).

## Erros/problemas gerados pela IA e correções feitas

- Primeira versão do cálculo de blocos (`chunkSize`) podia gerar um `end`
  maior que `len(words)` no último bloco — corrigido com checagem explícita
  de limite.
- Ausência inicial de tratamento para `numWorkers` inválido/negativo/maior
  que o número de palavras — corrigido com validações em `concurrentCount`
  e em `main.go`.

## Decisões técnicas tomadas ao longo da implementação

- Isolar a tokenização em uma função única e compartilhada, para que
  qualquer divergência entre versão sequencial e concorrente só possa vir
  da lógica de contagem/combinação, nunca da entrada.
- Medir tempo **apenas** ao redor da chamada às funções de contagem
  (`sequentialCount` / `concurrentCount`), deixando leitura de arquivo e
  tokenização fora da medição, por serem etapas idênticas para as duas
  versões e não fazerem parte do que se quer comparar.
- Comparar o mapa de frequências **completo** (não só o top 20) como
  critério de corretude, conforme exigido no roteiro.
- Verificar ausência de condição de corrida de forma empírica
  (`go run -race .`), além da análise de código.

---

## Perguntas finais para discussão

**Qual ferramenta de IA foi utilizada e em qual ambiente?**
Claude (Anthropic), via interface de chat web (claude.ai), com execução de
código habilitada em sandbox Linux/Ubuntu (`go1.22.2 linux/amd64`).

**Qual estratégia concorrente foi escolhida?**
Fan-out/fan-in: divisão da lista de palavras em `numWorkers` blocos, mapa
local de contagem por goroutine, envio dos mapas parciais por `channel`,
sincronização com `sync.WaitGroup` e combinação final sequencial.

**Quais alternativas foram consideradas?**
Mapa global com `sync.Mutex`, `sync.Map`, e pool de workers com fila
dinâmica de tarefas.

**Por que a estratégia escolhida pareceu mais adequada?**
Por eliminar por construção qualquer escrita concorrente em estrutura
compartilhada durante a fase de contagem (menor risco de condição de
corrida), por ter o menor custo de sincronização entre as alternativas
avaliadas (nenhum lock, apenas um `WaitGroup`), e por ser a mais simples de
verificar corretude.

**O resultado concorrente foi exatamente igual ao sequencial?**
Sim, em todas as execuções realizadas (arquivo de teste pequeno, dataset
principal com 1/2/4/8 workers, e arquivo sintético ~50× maior),
`mapsEqual` retornou `true` comparando os mapas completos.

**A versão concorrente melhorou o desempenho?**
Depende do tamanho da entrada. Para o dataset principal (~23 mil palavras),
não — o overhead de criar goroutines/channels superou o ganho, especialmente
na máquina de 1 núcleo usada nos testes. Para o arquivo ~50× maior (~1,17
milhão de palavras), sim — a versão concorrente foi consistentemente mais
rápida (1.03x a 1.30x), mesmo sem paralelismo real de hardware.

**Onde poderia ocorrer condição de corrida?**
Se, por erro de implementação, mais de uma goroutine escrevesse diretamente
no mesmo `map[string]int` compartilhado (por exemplo, se todas as goroutines
recebessem uma referência ao mesmo mapa `final` em vez de mapas `local`
independentes), ou se a etapa de combinação começasse a ler do channel antes
de todas as goroutines de contagem terminarem.

**A sincronização usada era necessária?**
Sim — o `sync.WaitGroup` é o único mecanismo de sincronização do programa, e
é necessário para garantir a ordem correta entre "todas as goroutines
terminaram de contar" e "o channel pode ser fechado e lido até o fim" pela
goroutine principal.

**Houve contenção?**
Não. Como cada goroutine escreve apenas em seu próprio mapa local, não há
nenhum recurso compartilhado disputado durante a fase de contagem — logo,
não há contenção de lock nem de qualquer outro tipo.

**A IA sugeriu alguma solução problemática?**
Não uma solução fundamentalmente errada, mas a primeira versão do cálculo de
limites de bloco (`chunkSize`/`end`) tinha um caso de borda não tratado
(último bloco podendo ultrapassar o tamanho da fatia), corrigido durante a
implementação (Etapa 8).

**O que precisou ser corrigido na solução sugerida pela IA?**
O tratamento de limites do último bloco de trabalho e a validação de
`numWorkers` inválido/fora de faixa, ambos detalhados na Etapa 8.

**Qual é a diferença entre simplesmente usar recursos concorrentes e
construir uma boa solução concorrente?**
Simplesmente "usar concorrência" seria, por exemplo, disparar uma goroutine
por palavra ou usar um mutex genérico "por segurança" sem analisar o custo
disso. Uma boa solução concorrente parte do problema real (aqui, uma
operação de agregação/redução) e escolhe o padrão que minimiza
sincronização e risco de corrida — nesse caso, dividir o trabalho de forma
que cada unidade de execução não precise nunca competir por um recurso
compartilhado durante a parte cara do processamento, deixando a única
combinação necessária para um momento sequencial e seguro por construção.
