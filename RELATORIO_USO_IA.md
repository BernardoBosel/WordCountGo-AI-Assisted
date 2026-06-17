# Relatório de Uso de IA

## Ferramenta utilizada

- **Ferramenta:** Claude (Anthropic)
- **Modelo:** Claude Sonnet 4.6
- **Ambiente:** Interface de chat web/app do Claude (claude.ai), com a
  funcionalidade de execução de código habilitada (sandbox Linux/Ubuntu com
  acesso a um terminal bash, usado para instalar o Go, compilar e executar
  o programa antes da entrega).
- **Versão do Go usada para compilar/testar:** `go1.22.2 linux/amd64`
  (instalada via `apt-get install golang-go` no ambiente da ferramenta).

## Como a IA foi utilizada

O enunciado completo da atividade (em português) foi fornecido à IA junto
com o arquivo de dataset (`AChristmasCarol_CharlesDickens_English.txt`) já
baixado. A partir disso, foram solicitados a implementação completa em Go
(versão sequencial + versão concorrente), a comparação de corretude entre as
duas, a medição de tempo, o `README.md` e este relatório de uso de IA.

## Passos realizados

1. **Leitura e interpretação do enunciado**, identificando os requisitos
   obrigatórios: tokenização (minúsculas, remoção de pontuação simples,
   filtro de palavras com menos de 3 caracteres), contagem sequencial,
   contagem concorrente, comparação de mapas completos, medição de tempo
   separada para cada versão e impressão do top 20.
2. **Definição da arquitetura do código**, separando o programa em arquivos
   por responsabilidade (`tokenize.go`, `sequential.go`, `concurrent.go`,
   `compare.go`, `main.go`) para deixar a versão sequencial e a versão
   concorrente claramente isoladas e fáceis de comparar.
3. **Implementação da versão sequencial** (`sequentialCount`): laço simples
   sobre a lista de palavras, incrementando um `map[string]int`.
4. **Implementação da versão concorrente** (`concurrentCount`), usando o
   padrão fan-out/fan-in: divisão da lista de palavras em N blocos,
   processamento de cada bloco em uma goroutine própria (mapa local, sem
   compartilhamento), envio dos mapas parciais por um `channel`, sincronização
   com `sync.WaitGroup` e combinação final sequencial dos mapas parciais.
5. **Implementação da comparação de corretude** (`mapsEqual`), comparando o
   mapa de frequências completo (todas as palavras, não só o top 20) entre
   as duas versões.
6. **Implementação da medição de tempo**, usando `time.Now()`/`time.Since()`
   ao redor apenas da chamada de cada função de contagem, de forma
   equivalente para as duas versões.
7. **Configuração do ambiente de testes**: como o Go não estava
   pré-instalado no sandbox da ferramenta, foi instalado via
   `apt-get install golang-go` (versão resultante: `go1.22.2`).
8. **Verificação de qualidade do código**: execução de `gofmt -l` (sem
   alterações necessárias) e `go vet ./...` (sem problemas reportados).
9. **Verificação de ausência de data races**: execução do programa com a
   flag `go run -race .`, sem nenhuma corrida de dados reportada.
10. **Execução e validação funcional**: o programa foi executado várias
    vezes contra o dataset fornecido, com diferentes números de workers
    (padrão/`NumCPU`, 4 e 8), confirmando em todas as execuções
    `Resultados iguais: sim` e a mesma lista de top 20 palavras.
11. **Teste de escala**: o dataset original foi concatenado 50 vezes para
    gerar um arquivo ~50× maior, e o programa foi executado novamente para
    observar o comportamento de tempo em uma entrada maior.
12. **Redação do `README.md`**, documentando a estratégia concorrente
    utilizada, como o trabalho foi dividido, como os resultados parciais
    foram combinados, como a correção foi verificada, como o tempo foi
    medido, a comparação de desempenho entre as versões (com os números
    reais observados nos testes) e as dificuldades encontradas.
13. **Redação deste relatório de uso de IA.**

## Observação sobre o ambiente de teste

A máquina/sandbox usada pela IA para testar e validar o código durante o
desenvolvimento possui apenas **1 CPU lógica** (`nproc` = 1), o que limita o
paralelismo real observável na versão concorrente. Essa limitação está
documentada na seção "Análise de desempenho" do `README.md`. Recomenda-se
executar o programa também na máquina pessoal usada para a entrega e, se os
números de tempo forem diferentes dos relatados aqui, atualizar a seção de
análise do `README.md` com os valores observados localmente.
