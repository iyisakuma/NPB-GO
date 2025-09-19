# Quick Reference - Padrões de Paralelização

## 🚀 **Padrões Aplicados**

### **1. Worker Pool Pattern**
```go
// Origem: Java ExecutorService, .NET TPL
var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go worker(i, &wg)
}
wg.Wait()
```

### **2. Fork-Join Pattern**
```go
// Origem: Java ForkJoinPool, OpenMP parallel sections
// FORK: Launch workers
go worker1()
go worker2()
// JOIN: Wait for completion
wg.Wait()
```

### **3. Data Parallelism Pattern**
```go
// Origem: OpenMP #pragma omp parallel for, Rayon par_iter()
keysPerWorker := (totalKeys + numWorkers - 1) / numWorkers
for i := 0; i < numWorkers; i++ {
    start := i * keysPerWorker
    end := start + keysPerWorker
    go processRange(start, end)
}
```

### **4. Critical Section Pattern**
```go
// Origem: OpenMP #pragma omp critical, mutex patterns
// Sequential operations that cannot be parallelized
for _, item := range criticalData {
    // Must be sequential for correctness
}
```

## 🔧 **Padrões Específicos do Go**

### **Independent Random Streams**
```go
// Problema: Race conditions em geração de números aleatórios
func (b *ISBenchmark) findMySeed(processorRank, numberProcessor int, ...) float64 {
    // Algoritmo "skip-ahead" do OpenMP
    // Cada worker tem seu próprio stream independente
}
```

### **Parallel Initialization**
```go
// Problema: Inicialização de grandes arrays
func (b *ISBenchmark) allocKeyBuff() {
    // Chunk-based parallel initialization (Rayon pattern)
    var wg sync.WaitGroup
    for i := 0; i < b.numProcs; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            // Parallel initialization of chunk
        }(i)
    }
    wg.Wait()
}
```

### **Adaptive Load Balancing**
```go
// Problema: Distribuição uniforme de trabalho
keysPerWorker := (NUM_KEYS + b.numProcs - 1) / b.numProcs
k1 := keysPerWorker * workerID
k2 := k1 + keysPerWorker
if k2 > NUM_KEYS {
    k2 = NUM_KEYS
}
```

## 📊 **Comparação com Outras Linguagens**

| Padrão | C++ (OpenMP) | Rust (Rayon) | Go (Nativo) | Java (ExecutorService) |
|--------|--------------|--------------|-------------|----------------------|
| **Worker Pool** | `#pragma omp parallel` | `par_iter()` | `go func()` | `ExecutorService.submit()` |
| **Fork-Join** | `#pragma omp sections` | `join()` | `sync.WaitGroup` | `ForkJoinPool` |
| **Data Parallelism** | `#pragma omp parallel for` | `par_chunks()` | Range-based workers | `Parallel.For()` |
| **Critical Section** | `#pragma omp critical` | `mutex` | Sequential code | `synchronized` |

## 🎯 **Quando Usar Cada Padrão**

### **Worker Pool Pattern**
- ✅ **Use quando**: Controle preciso do número de workers
- ✅ **Use quando**: Trabalho independente entre workers
- ❌ **Não use quando**: Trabalho muito pequeno (overhead)

### **Fork-Join Pattern**
- ✅ **Use quando**: Decomposição natural de problemas
- ✅ **Use quando**: Sincronização automática necessária
- ❌ **Não use quando**: Dependências complexas entre tasks

### **Data Parallelism Pattern**
- ✅ **Use quando**: Processamento de arrays grandes
- ✅ **Use quando**: Trabalho uniforme por elemento
- ❌ **Não use quando**: Dependências entre elementos

### **Critical Section Pattern**
- ✅ **Use quando**: Operações que devem ser sequenciais
- ✅ **Use quando**: Garantia de correção é crítica
- ❌ **Não use quando**: Performance é mais importante que correção

## 🚀 **Otimizações Go-Specific**

### **sync.Pool para Memory Reuse**
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}
```

### **Channels para Communication**
```go
results := make(chan Result, numWorkers)
for i := 0; i < numWorkers; i++ {
    go worker(i, results)
}
```

### **Context para Cancellation**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

## 📈 **Métricas de Performance**

### **Speedup Esperado**
- **2 cores**: ~1.8x
- **4 cores**: ~3.2x
- **8 cores**: ~5.6x

### **Overhead de Sincronização**
- **WaitGroup**: ~1-2μs
- **Channel**: ~5-10μs
- **Mutex**: ~10-20μs

### **Memory Overhead**
- **Goroutine**: ~2KB stack
- **Channel**: ~96 bytes
- **WaitGroup**: ~12 bytes

## 🔍 **Debugging Patterns**

### **Worker Identification**
```go
func worker(workerID int, wg *sync.WaitGroup) {
    defer wg.Done()
    log.Printf("Worker %d: starting", workerID)
    // ... work ...
    log.Printf("Worker %d: completed", workerID)
}
```

### **Performance Monitoring**
```go
start := time.Now()
// ... parallel work ...
elapsed := time.Since(start)
log.Printf("Parallel work took %v", elapsed)
```

### **Error Handling**
```go
func worker(workerID int, wg *sync.WaitGroup, errChan chan error) {
    defer wg.Done()
    if err := doWork(); err != nil {
        errChan <- fmt.Errorf("worker %d: %w", workerID, err)
    }
}
```

## 🎯 **Best Practices**

### **1. Start Simple**
- Paralelizar apenas o que é seguro
- Validar correção constantemente
- Medir impacto de cada mudança

### **2. Use Established Patterns**
- Worker Pool para task distribution
- Fork-Join para decomposição
- Data Parallelism para processamento

### **3. Consider Go-Specific**
- Channels para comunicação
- Context para cancellation
- sync.Pool para memory reuse

### **4. Profile Before Optimize**
- Identificar bottlenecks reais
- Medir overhead de sincronização
- Validar escalabilidade

---

**Esta referência rápida fornece os padrões essenciais para implementação de paralelização em Go, baseados em padrões estabelecidos do mercado e adaptados para as características específicas da linguagem.**
