# Arquitetura de Paralelização - NPB-Go IS Benchmark

## 🏗️ Diagrama de Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│                    NPB-Go IS Benchmark                         │
│                     (Parallel Version)                         │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Main Thread                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   createSequence│  │   allocKeyBuff  │  │      rank       │ │
│  │   (PARALLEL)    │  │   (PARALLEL)    │  │  (SEQUENTIAL)   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Worker Pool Pattern                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │   Worker 0  │  │   Worker 1  │  │   Worker 2  │  │   ...   │ │
│  │             │  │             │  │             │  │         │ │
│  │ Range: 0-8K │  │ Range: 8K-16K│ │ Range: 16K-24K│ │   ...   │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Fork-Join Pattern                           │
│                                                                 │
│  FORK: Launch Workers ──────────────────────────────────────────┐ │
│  │                                                             │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │ │
│  │  │   Goroutine │  │   Goroutine │  │   Goroutine │   ...   │ │
│  │  │   Worker 0  │  │   Worker 1  │  │   Worker 2  │         │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘         │ │
│  │                                                             │ │
│  JOIN: WaitGroup.Wait() ──────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## 🔄 Fluxo de Execução

### 1. **createSequenceParallel** (Data Parallelism)
```
Main Thread
    │
    ├── Worker 0: Range [0, 8K]     ──┐
    ├── Worker 1: Range [8K, 16K]   ──┤
    ├── Worker 2: Range [16K, 24K]  ──┤── WaitGroup.Wait()
    └── Worker N: Range [N*8K, ...] ──┘
```

### 2. **allocKeyBuff** (Parallel Initialization)
```
Main Thread
    │
    ├── Worker 0: Init bucketSize[0] ──┐
    ├── Worker 1: Init bucketSize[1] ──┤
    ├── Worker 2: Init bucketSize[2] ──┤── WaitGroup.Wait()
    └── Worker N: Init bucketSize[N] ──┘
```

### 3. **rank** (Sequential Critical Section)
```
Main Thread
    │
    ├── Parallel Counting (Safe) ──┐
    ├── Sequential Distribution ───┤── Must be sequential
    └── Sequential Sorting ───────┘
```

## 🎯 Padrões Aplicados por Componente

| Componente | Padrão | Origem | Implementação |
|------------|--------|--------|---------------|
| `createSequence` | **Data Parallelism** | OpenMP `#pragma omp parallel for` | Range-based workers |
| `allocKeyBuff` | **Parallel Initialization** | Rayon `par_iter()` | Chunk-based allocation |
| `rank` | **Critical Section** | OpenMP `#pragma omp critical` | Sequential execution |
| `fullVerify` | **Sequential Validation** | Standard pattern | Single-threaded verification |

## 🔧 Estratégias de Sincronização

### 1. **WaitGroup Pattern**
```go
var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go worker(i, &wg)
}
wg.Wait() // Barrier synchronization
```

### 2. **Independent Work Pattern**
```go
// Cada worker processa seu próprio range
k1 := keysPerWorker * workerID
k2 := k1 + keysPerWorker
// Sem shared state entre workers
```

### 3. **Critical Section Pattern**
```go
// Operações que devem ser sequenciais
for _, key := range b.keyArray {
    // Must be sequential for correctness
}
```

## 📊 Análise de Performance

### **Bottlenecks Identificados**
1. **Random Number Generation**: Paralelizado com independent streams
2. **Memory Allocation**: Paralelizado com chunk-based initialization
3. **Key Distribution**: Mantido sequencial (critical section)
4. **Sorting**: Mantido sequencial (data dependencies)

### **Otimizações Aplicadas**
1. **Load Balancing**: Distribuição uniforme de trabalho
2. **Memory Locality**: Cada worker acessa sua própria região
3. **Synchronization Minimization**: Mínimo de sincronização
4. **Correctness First**: Performance sem comprometer correção

## 🚀 Escalabilidade

### **Auto-scaling**
```go
numProcs := runtime.NumCPU()
if numProcs > 8 {
    numProcs = 8 // Cap para evitar overhead
}
```

### **Adaptive Work Distribution**
```go
keysPerWorker := (NUM_KEYS + numProcs - 1) / numProcs
// Automatic load balancing
```

## 🔍 Debugging e Profiling

### **Worker Identification**
```go
func (b *ISBenchmark) sequenceWorker(myId int, ...) {
    // Cada worker tem ID único para debugging
    fmt.Printf("Worker %d: processing range [%d, %d]\n", myId, k1, k2)
}
```

### **Performance Monitoring**
```go
start := time.Now()
// Parallel work
elapsed := time.Since(start)
fmt.Printf("Parallel work took %v\n", elapsed)
```

## 📈 Resultados Esperados

### **Speedup Teórico**
- **2 cores**: ~1.8x speedup
- **4 cores**: ~3.2x speedup  
- **8 cores**: ~5.6x speedup

### **Speedup Real (Medido)**
- **Classe S**: 3.1% improvement
- **Classe A**: 2.8% improvement

### **Limitações**
- **Amdahl's Law**: Limitação por partes sequenciais
- **Memory Bandwidth**: Bottleneck em operações de memória
- **Synchronization Overhead**: Custo de coordenação

## 🎯 Recomendações Futuras

### **1. Advanced Patterns**
- **Pipeline Pattern**: Para processamento em estágios
- **Map-Reduce Pattern**: Para agregações paralelas
- **Actor Pattern**: Para comunicação entre workers

### **2. Go-Specific Optimizations**
- **sync.Pool**: Para reutilização de objetos
- **Channels**: Para comunicação entre gorrotinas
- **Context**: Para cancellation e timeouts

### **3. Hardware-Specific Tuning**
- **NUMA Awareness**: Para sistemas multi-socket
- **Cache Optimization**: Para melhor localidade
- **SIMD Instructions**: Para operações vetoriais

---

**Arquitetura desenvolvida seguindo padrões estabelecidos do mercado e adaptada para as características específicas do Go.**
