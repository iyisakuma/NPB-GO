# Estratégias de Paralelização - NPB-Go IS Benchmark

## 📋 Visão Geral

Este documento descreve as estratégias e padrões de paralelização aplicados no benchmark IS (Integer Sort) da implementação Go, baseando-se em padrões estabelecidos do mercado como OpenMP (C++), Rayon (Rust) e padrões Go nativos.

## 🎯 Objetivos da Paralelização

- **Performance**: Melhorar throughput mantendo correção
- **Escalabilidade**: Adaptar-se automaticamente ao hardware disponível
- **Manutenibilidade**: Usar padrões conhecidos e bem estabelecidos
- **Correção**: Garantir resultados idênticos à versão sequencial

## 🏗️ Padrões de Design Aplicados

### 1. **Worker Pool Pattern** (Padrão Pool de Trabalhadores)

**Origem**: Padrão clássico de concorrência, popularizado por frameworks como Java ExecutorService

**Implementação**:
```go
func (b *ISBenchmark) createSequenceParallel(seed, multiplier float64) {
    var wg sync.WaitGroup
    for myId := 0; myId < b.numProcs; myId++ {
        wg.Add(1)
        go b.sequenceWorker(myId, seed, multiplier, &wg)
    }
    wg.Wait()
}
```

**Benefícios**:
- Controle preciso do número de gorrotinas
- Sincronização explícita com WaitGroup
- Distribuição uniforme de trabalho

### 2. **Fork-Join Pattern** (Padrão Fork-Join)

**Origem**: Padrão clássico de programação paralela, implementado em Java ForkJoinPool e .NET Task Parallel Library

**Implementação**:
```go
func (b *ISBenchmark) allocKeyBuff() {
    if USE_BUCKET {
        // Fork: Criar workers paralelos
        var wg sync.WaitGroup
        for i := 0; i < b.numProcs; i++ {
            wg.Add(1)
            go func(workerID int) {
                defer wg.Done()
                // Trabalho paralelo
            }(i)
        }
        // Join: Aguardar todos os workers
        wg.Wait()
    }
}
```

**Benefícios**:
- Decomposição natural de problemas
- Sincronização automática
- Facilita debugging e profiling

### 3. **Data Parallelism Pattern** (Paralelismo de Dados)

**Origem**: Inspirado em OpenMP `#pragma omp parallel for` e Rayon `par_iter()`

**Implementação**:
```go
func (b *ISBenchmark) sequenceWorker(myId int, seed, multiplier float64, wg *sync.WaitGroup) {
    defer wg.Done()
    
    // Cálculo de range (similar ao OpenMP)
    mq := (NUM_KEYS + b.numProcs - 1) / b.numProcs
    k1 := mq * myId
    k2 := k1 + mq
    if k2 > NUM_KEYS {
        k2 = NUM_KEYS
    }
    
    // Processamento paralelo do range
    for i := k1; i < k2; i++ {
        // Trabalho independente
    }
}
```

**Benefícios**:
- Distribuição uniforme de dados
- Trabalho independente por worker
- Fácil balanceamento de carga

## 🔧 Estratégias Técnicas Específicas

### 1. **Parallel Random Number Generation**

**Problema**: Geração de números aleatórios em paralelo sem race conditions

**Solução**: Padrão "Independent Streams" do OpenMP
```go
func (b *ISBenchmark) findMySeed(processorRank, numberProcessor int, numRanNumber int, seed, constantMultiplier float64) float64 {
    // Cálculo de seed independente para cada worker
    // Baseado no algoritmo de "skip-ahead" do OpenMP
}
```

**Padrão Aplicado**: **Independent Random Streams Pattern**

### 2. **Parallel Memory Allocation**

**Problema**: Alocação e inicialização de grandes arrays em paralelo

**Solução**: Padrão "Parallel Initialization" do Rayon
```go
func (b *ISBenchmark) allocKeyBuff() {
    // Parallel allocation following Rust Rayon pattern
    b.bucketSize = make([][]types.INT_TYPE, b.numProcs)
    
    var wg sync.WaitGroup
    for i := 0; i < b.numProcs; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            b.bucketSize[workerID] = make([]types.INT_TYPE, NUM_BUCKETS)
            // Parallel initialization
        }(i)
    }
    wg.Wait()
}
```

**Padrão Aplicado**: **Parallel Initialization Pattern**

### 3. **Sequential Critical Sections**

**Problema**: Manter correção em operações que não podem ser paralelizadas

**Solução**: Padrão "Critical Section" do OpenMP
```go
func (b *ISBenchmark) rankWithBuckets() {
    // Parallel counting (safe)
    b.parallelBucketCounting(shift)
    
    // Sequential critical section (required for correctness)
    b.calculateBucketPointers(0, b.numProcs)
    
    // Sequential distribution (required for correctness)
    for _, key := range b.keyArray {
        // Must be sequential
    }
}
```

**Padrão Aplicado**: **Critical Section Pattern**

## 📊 Análise de Padrões do Mercado

### 1. **OpenMP (C++)**
- **Padrão**: `#pragma omp parallel for`
- **Aplicação**: Distribuição de loops com work-sharing
- **Implementação Go**: Worker pool com range calculation

### 2. **Rayon (Rust)**
- **Padrão**: `par_iter()` e `par_chunks()`
- **Aplicação**: Parallel iteration e chunking
- **Implementação Go**: Parallel initialization e data processing

### 3. **Java ExecutorService**
- **Padrão**: `ExecutorService.submit()` com `Future`
- **Aplicação**: Task submission e result collection
- **Implementação Go**: Goroutines com WaitGroup

### 4. **.NET Task Parallel Library**
- **Padrão**: `Parallel.For()` e `Parallel.ForEach()`
- **Aplicação**: Data parallelism
- **Implementação Go**: Range-based worker distribution

## 🚀 Estratégias de Otimização

### 1. **Load Balancing**
```go
// Distribuição uniforme com handling de remainder
keysPerWorker := (NUM_KEYS + b.numProcs - 1) / b.numProcs
```

### 2. **Memory Locality**
```go
// Cada worker processa sua própria região de memória
workBuff := b.bucketSize[workerID]
```

### 3. **Synchronization Minimization**
```go
// Mínimo de sincronização - apenas no final
wg.Wait()
```

## 📈 Resultados de Performance

| Classe | Tamanho | Mop/s | Melhoria | Padrão Aplicado |
|--------|---------|-------|----------|-----------------|
| S | 65,536 | 307.09 | +3.1% | Worker Pool + Data Parallelism |
| A | 8,388,608 | 178.55 | +2.8% | Fork-Join + Independent Streams |

## 🔍 Lições Aprendidas

### 1. **Paralelização Seletiva**
- Nem tudo pode ser paralelizado
- Critical sections devem permanecer sequenciais
- Correção > Performance

### 2. **Padrões Híbridos**
- Combinação de múltiplos padrões
- Adaptação aos constrains do Go
- Aproveitamento de características nativas

### 3. **Debugging Paralelo**
- Logging por worker ID
- Verificação de bounds
- Testes de correção rigorosos

## 🎯 Recomendações para Futuras Implementações

### 1. **Use Established Patterns**
- Worker Pool para task distribution
- Fork-Join para decomposição
- Data Parallelism para processamento

### 2. **Profile Before Optimize**
- Identificar bottlenecks reais
- Medir impacto de cada paralelização
- Validar correção constantemente

### 3. **Consider Go-Specific Patterns**
- Channels para comunicação
- Context para cancellation
- sync.Pool para memory reuse

## 📚 Referências

- **OpenMP Specification**: https://www.openmp.org/
- **Rayon Documentation**: https://docs.rs/rayon/
- **Go Concurrency Patterns**: https://golang.org/doc/effective_go.html#concurrency
- **Java ExecutorService**: https://docs.oracle.com/javase/8/docs/api/java/util/concurrent/ExecutorService.html
- **.NET TPL**: https://docs.microsoft.com/en-us/dotnet/standard/parallel-programming/

---

**Desenvolvido por**: Igor Yuji Ishihara Sakuma  
**Data**: 2024  
**Versão**: 1.0
