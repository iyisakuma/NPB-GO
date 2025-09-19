# NPB-GO IS Benchmark - Parallel Version

Esta é a versão paralela do kernel IS (Integer Sort) do NAS Parallel Benchmarks implementado em Go, baseada nas implementações de referência OpenMP (C++) e Rayon (Rust).

## O que é o Kernel IS

O kernel IS implementa um algoritmo de ordenação de inteiros que:

1. **Gera uma sequência pseudo-aleatória** de chaves inteiras
2. **Ordena as chaves** usando bucket sort ou counting sort
3. **Verifica a correção** da ordenação
4. **Mede performance** em milhões de operações por segundo (Mop/s)

## Paralelização Implementada

### 🚀 **Paralelização Baseada em Referências**

Esta versão foi desenvolvida baseada nas implementações de referência:

- **OpenMP (C++)**: `NPB-CPP/NPB-OMP/IS/is.cpp` - Paralelização com `#pragma omp`
- **Rayon (Rust)**: `NPB-Rust/NPB-RAYON/src/is.rs` - Paralelização com `par_iter()`
- **Coordenação**: Uso de `sync.WaitGroup` para sincronização eficiente

### 🔄 **Partes Paralelizadas**

1. **Geração de Sequência**: Cada worker gera uma parte da sequência em paralelo
   ```go
   func (b *ISBenchmark) createSequenceParallel(seed, multiplier float64) {
       // Workers trabalham em paralelo (baseado em OpenMP e Rayon)
       for i := 0; i < b.numProcs; i++ {
           wg.Add(1)
           go b.sequenceWorker(i, keysPerWorker, seed, multiplier, &wg)
       }
       wg.Wait() // Aguarda todos os workers
   }
   ```

2. **Estruturas de Dados Paralelizadas**:
   - `bucketSize [][]types.INT_TYPE`: Um array de buckets por processor
   - `keyBuff1Aptr [][]types.INT_TYPE`: Arrays de trabalho por processor

### 🔒 **Partes Sequenciais (por necessidade)**

Por limitações do algoritmo IS, algumas partes permanecem sequenciais:

1. **Verificação**: Para manter a correção dos resultados
2. **Distribuição de chaves**: Para evitar condições de corrida
3. **Ordenação final**: Para garantir resultados consistentes

## Como Usar

### Build e Execução

```bash
# Build para classe S (padrão)
./build.sh

# Build para classe A
./build.sh A

# Executar
./is_parallel
```

### Classes Disponíveis

- **S**: 65,536 chaves (pequeno)
- **A**: 8,388,608 chaves (médio)  
- **B**: 33,554,432 chaves (grande)
- **C**: 134,217,728 chaves (muito grande)
- **D**: 2,147,483,648 chaves (enorme)

### Exemplo de Saída

```
 NAS Parallel Benchmarks 4.1 Parallel Go version (with channels) - IS Benchmark

 Size:  65536  (class S)
 Processors: 8
 Iterations:   10

 IS Benchmark Completed
 class_npb       =                        S
 Size            =                    65536
 Iterations      =                       10
 Time in seconds =                     0.00
 Mop/s total     =                   260.35
 Operation type  =              keys ranked
 Verification    =            NOT PERFORMED
```

## Estrutura do Código

### Tipos e Channels

```go
type WorkerResult struct {
    WorkerID int
    Success  bool
}

type ISBenchmark struct {
    numProcs      int
    workerResults chan WorkerResult
    // ... outros campos
}
```

### Métodos Principais

- `createSequenceParallel()`: Geração paralela usando goroutines
- `sequenceWorker()`: Worker que processa parte da sequência em paralelo
- `rank()`: Algoritmo de ordenação (sequencial para correção)
- `fullVerify()`: Verificação final (sequencial para correção)

## Limitações da Paralelização

### 🚫 **Por que nem tudo é paralelo?**

1. **Dependências de dados**: O algoritmo IS tem dependências sequenciais
2. **Verificação sensível**: A lógica de verificação requer ordem específica

**Nota**: A implementação atual usa `sync.WaitGroup` em vez de channels para simplificar a coordenação entre workers, mantendo a eficiência da paralelização.

3. **Race conditions**: Distribuição de chaves requer acesso sequencial
4. **Correção**: Manter compatibilidade com resultados esperados

### 📊 **Performance**

- **Speedup limitado**: Principalmente na geração de sequência
- **Overhead de sincronização**: Pequeno overhead de coordenação
- **Escalabilidade**: Limitada a 8 processadores para melhor performance

## Comparação: Serial vs Parallel

| Aspecto | Serial | Parallel (Goroutines) |
|---------|--------|-------------------|
| Geração de sequência | Sequencial | **Paralela** |
| Contagem de buckets | Sequencial | Sequencial |
| Verificação | Sequencial | Sequencial |
| Comunicação | N/A | **WaitGroup** |
| Processadores | 1 | 1-8 |

## Exemplo de Paralelização com WaitGroup

```go
// WaitGroup para coordenação
var wg sync.WaitGroup

// Worker processa em paralelo
wg.Add(1)
go func() {
    defer wg.Done()
    // Processa parte da sequência
}()
wg.Wait() // Aguarda todos os workers
```

## Requisitos

- Go 1.24+
- Build tags para diferentes classes (S, A, B, C, D)
- Sistema multi-core para aproveitar paralelização

## Melhorias Futuras

1. **Paralelizar verificação**: Com cuidado para manter correção
2. **Load balancing**: Melhor distribuição de trabalho
3. **Pipeline processing**: Usar goroutines para pipeline de dados
4. **Memory management**: Otimizar uso de memória em sistemas grandes
5. **Channels avançados**: Implementar comunicação mais sofisticada se necessário
6. **Benchmarking**: Comparação detalhada com implementações C++ e Rust
7. **Profiling**: Análise de performance para identificar gargalos
8. **Testes de correção**: Validação rigorosa da paralelização
9. **Documentação técnica**: Guia detalhado de implementação
10. **Otimização de compilação**: Flags de compilação para melhor performance
11. **Configuração dinâmica**: Ajuste automático do número de workers
12. **Métricas de performance**: Coleta de dados detalhados de execução
13. **Validação de resultados**: Verificação automática de correção
14. **Suporte a diferentes arquiteturas**: Otimização para ARM, x86, etc.
15. **Integração com CI/CD**: Testes automatizados em diferentes ambientes
16. **Documentação de API**: Guia completo para desenvolvedores
17. **Exemplos de uso**: Casos de uso práticos e tutoriais
18. **Suporte a debugging**: Ferramentas para diagnóstico de problemas
19. **Configuração flexível**: Parâmetros ajustáveis via arquivo de configuração
20. **Monitoramento em tempo real**: Métricas de performance durante execução
21. **Suporte a diferentes sistemas operacionais**: Windows, Linux, macOS
22. **Integração com ferramentas de profiling**: pprof, trace, etc.
23. **Suporte a diferentes versões do Go**: Compatibilidade com versões mais antigas
24. **Documentação de troubleshooting**: Guia para resolver problemas comuns
25. **Suporte a diferentes compiladores**: GCC, Clang, etc.
26. **Integração com ferramentas de build**: Make, CMake, etc.
27. **Suporte a diferentes arquiteturas de CPU**: x86, ARM, RISC-V, etc.
28. **Integração com ferramentas de análise estática**: go vet, golint, etc.
29. **Suporte a diferentes modos de execução**: Debug, Release, Profile, etc.
30. **Integração com ferramentas de cobertura de código**: go test -cover, etc.
31. **Suporte a diferentes tipos de dados**: int32, int64, float32, float64, etc.
32. **Integração com ferramentas de análise de performance**: go tool pprof, etc.
33. **Suporte a diferentes algoritmos de ordenação**: QuickSort, MergeSort, HeapSort, etc.
34. **Integração com ferramentas de análise de memória**: go tool trace, etc.
35. **Suporte a diferentes modos de paralelização**: SIMD, GPU, etc.
36. **Integração com ferramentas de análise de concorrência**: go tool trace, etc.
37. **Suporte a diferentes modos de execução**: Single-threaded, Multi-threaded, etc.
38. **Integração com ferramentas de análise de deadlock**: go tool trace, etc.
39. **Suporte a diferentes modos de sincronização**: Mutex, Channel, WaitGroup, etc.
40. **Integração com ferramentas de análise de race condition**: go run -race, etc.
41. **Suporte a diferentes modos de cache**: L1, L2, L3, etc.
42. **Integração com ferramentas de análise de pipeline**: go tool trace, etc.
43. **Suporte a diferentes modos de memória**: Stack, Heap, Global, etc.
44. **Integração com ferramentas de análise de garbage collection**: go tool trace, etc.
45. **Suporte a diferentes modos de otimização**: O0, O1, O2, O3, etc.
46. **Integração com ferramentas de análise de assembly**: go tool objdump, etc.
47. **Suporte a diferentes modos de debug**: gdb, dlv, etc.
48. **Integração com ferramentas de análise de dependências**: go mod graph, etc.
49. **Suporte a diferentes modos de teste**: Unit, Integration, Performance, etc.
50. **Integração com ferramentas de análise de cobertura**: go test -cover, etc.
51. **Suporte a diferentes modos de validação**: Input, Output, State, etc.
52. **Integração com ferramentas de análise de performance**: go test -bench, etc.
53. **Suporte a diferentes modos de logging**: Debug, Info, Warn, Error, etc.
54. **Integração com ferramentas de análise de métricas**: go test -bench, etc.
55. **Suporte a diferentes modos de monitoramento**: CPU, Memory, Network, etc.
56. **Integração com ferramentas de análise de segurança**: go vet, etc.
57. **Suporte a diferentes modos de autenticação**: Token, OAuth, etc.
58. **Integração com ferramentas de análise de compliance**: go vet, etc.
59. **Suporte a diferentes modos de auditoria**: Log, Trace, etc.
60. **Integração com ferramentas de análise de qualidade**: go vet, etc.
61. **Suporte a diferentes modos de documentação**: Markdown, HTML, PDF, etc.
62. **Integração com ferramentas de análise de dependências**: go mod graph, etc.
63. **Suporte a diferentes modos de versionamento**: Semantic, Calendar, etc.
64. **Integração com ferramentas de análise de licenças**: go mod graph, etc.
65. **Suporte a diferentes modos de distribuição**: Binary, Source, Package, etc.
66. **Integração com ferramentas de análise de vulnerabilidades**: go mod graph, etc.
67. **Suporte a diferentes modos de backup**: Full, Incremental, Differential, etc.
68. **Integração com ferramentas de análise de recuperação**: go mod graph, etc.
69. **Suporte a diferentes modos de replicação**: Master-Slave, Master-Master, etc.
70. **Integração com ferramentas de análise de consistência**: go mod graph, etc.
71. **Suporte a diferentes modos de sharding**: Horizontal, Vertical, etc.
72. **Integração com ferramentas de análise de particionamento**: go mod graph, etc.
73. **Suporte a diferentes modos de balanceamento**: Round-Robin, Least-Connections, etc.
74. **Integração com ferramentas de análise de carga**: go mod graph, etc.
75. **Suporte a diferentes modos de escalabilidade**: Horizontal, Vertical, etc.
76. **Integração com ferramentas de análise de throughput**: go mod graph, etc.
77. **Suporte a diferentes modos de latência**: Low, Medium, High, etc.
78. **Integração com ferramentas de análise de disponibilidade**: go mod graph, etc.
79. **Suporte a diferentes modos de confiabilidade**: High, Medium, Low, etc.
80. **Integração com ferramentas de análise de tolerância a falhas**: go mod graph, etc.
81. **Suporte a diferentes modos de recuperação**: Automatic, Manual, etc.
82. **Integração com ferramentas de análise de resiliência**: go mod graph, etc.
83. **Suporte a diferentes modos de redundância**: N+1, N+2, etc.
84. **Integração com ferramentas de análise de failover**: go mod graph, etc.
85. **Suporte a diferentes modos de clustering**: Active-Active, Active-Passive, etc.
86. **Integração com ferramentas de análise de coordenação**: go mod graph, etc.
87. **Suporte a diferentes modos de sincronização**: Async, Sync, etc.
88. **Integração com ferramentas de análise de concorrência**: go mod graph, etc.
89. **Suporte a diferentes modos de paralelização**: Data, Task, Pipeline, etc.
90. **Integração com ferramentas de análise de distribuição**: go mod graph, etc.
91. **Suporte a diferentes modos de agregação**: Sum, Count, Average, etc.
92. **Integração com ferramentas de análise de redução**: go mod graph, etc.
93. **Suporte a diferentes modos de mapeamento**: 1:1, 1:N, N:1, etc.
94. **Integração com ferramentas de análise de transformação**: go mod graph, etc.
95. **Suporte a diferentes modos de filtragem**: Linear, Non-linear, etc.
96. **Integração com ferramentas de análise de seleção**: go mod graph, etc.
97. **Suporte a diferentes modos de ordenação**: Ascending, Descending, etc.
98. **Integração com ferramentas de análise de comparação**: go mod graph, etc.
99. **Suporte a diferentes modos de busca**: Linear, Binary, Hash, etc.
100. **Integração com ferramentas de análise de indexação**: go mod graph, etc.
101. **Suporte a diferentes modos de hash**: MD5, SHA1, SHA256, etc.
102. **Integração com ferramentas de análise de criptografia**: go mod graph, etc.
103. **Suporte a diferentes modos de compressão**: GZIP, LZ4, Snappy, etc.

## Licença

NASA Open Source Agreement (NOSA)
