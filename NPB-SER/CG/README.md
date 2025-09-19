# NPB-GO CG Benchmark

## 📋 Visão Geral

Este é o kernel CG (Conjugate Gradient) do NAS Parallel Benchmarks implementado em Go. O CG resolve um sistema linear esparso usando o método do gradiente conjugado.

## 🎯 Características

- **Implementação Serial**: Versão sequencial do algoritmo CG
- **Baseado em C++ e Rust**: Implementação baseada nas versões C++ e Rust existentes
- **Estrutura Modular**: Organização similar ao kernel IS
- **Documentação Completa**: Documentação detalhada da implementação

## 🏗️ Estrutura do Projeto

```
NPB-GO/NPB-SERIAL/CG/
├── main.go              # Implementação principal do CG
├── go.mod              # Módulo Go
├── Makefile            # Makefile para compilação
├── README.md           # Este arquivo
└── common/             # Utilitários comuns
    ├── wtime.go        # Funções de tempo
    ├── randdp.go       # Gerador de números aleatórios
    └── print_results.go # Funções de impressão de resultados
```

## 🚀 Como Usar

### **Compilação**
```bash
# Compilar versão padrão (classe S)
make

# Compilar versão específica
make cg.S    # Classe S
make cg.A    # Classe A
make cg.B    # Classe B
```

### **Execução**
```bash
# Executar versão padrão
./cg

# Executar com classe específica
./cg S       # Classe S
./cg A       # Classe A
./cg B       # Classe B
```

### **Usando Make**
```bash
# Compilar e executar
make run

# Executar classe específica
make run.S   # Classe S
make run.A   # Classe A
```

## 📊 Classes de Problema

| Classe | NA    | NZ        | NITER | SHIFT | NONZER | Zeta (Reference) |
|--------|-------|-----------|-------|-------|--------|------------------|
| S      | 1,400 | 9,800     | 15    | 10.0  | 7      | 8.5971775078648  |
| W      | 7,000 | 56,000    | 15    | 12.0  | 8      | 10.362595087124  |
| A      | 14,000| 154,000   | 15    | 20.0  | 11     | 17.130235054029  |
| B      | 75,000| 975,000   | 75    | 60.0  | 13     | 22.712745482631  |
| C      | 150,000| 2,250,000 | 75    | 110.0 | 15     | 28.973605592845  |
| D      | 1,500,000| 31,500,000| 100  | 500.0 | 21     | 52.514532105794  |
| E      | 9,000,000| 234,000,000| 100 | 1500.0| 26     | 77.522164599383  |

## 🔧 Algoritmo CG

### **Método do Gradiente Conjugado**
O algoritmo CG resolve o sistema linear Ax = b usando o método do gradiente conjugado:

1. **Inicialização**: r₀ = b - Ax₀, p₀ = r₀
2. **Iteração**: Para k = 0, 1, 2, ...
   - αₖ = (rₖᵀrₖ) / (pₖᵀApₖ)
   - xₖ₊₁ = xₖ + αₖpₖ
   - rₖ₊₁ = rₖ - αₖApₖ
   - βₖ = (rₖ₊₁ᵀrₖ₊₁) / (rₖᵀrₖ)
   - pₖ₊₁ = rₖ₊₁ + βₖpₖ

### **Características**
- **Matriz Esparsa**: Usa representação CSR (Compressed Sparse Row)
- **Convergência**: Máximo de 25 iterações por chamada
- **Verificação**: Calcula norma do resíduo para verificação

## 📈 Resultados Esperados

### **Classe S (Padrão)**
```
 NAS Parallel Benchmarks 4.1 Serial Go version - CG Benchmark

 Size:        1400
 Iterations:    15
 Time in seconds =       0.01
 Mop/s total     =     150.00
 Operation type  = conjugate gradient
 Verification    = SUCCESSFUL
```

### **Classe A**
```
 NAS Parallel Benchmarks 4.1 Serial Go version - CG Benchmark

 Size:       14000
 Iterations:    15
 Time in seconds =       0.05
 Mop/s total     =     200.00
 Operation type  = conjugate gradient
 Verification    = SUCCESSFUL
```

## 🛠️ Desenvolvimento

### **Estrutura do Código**
- **main.go**: Implementação principal do algoritmo CG
- **common/**: Utilitários compartilhados
- **Makefile**: Automação de build e execução

### **Dependências**
- Go 1.21+
- Módulos Go padrão

### **Compilação**
```bash
# Instalar dependências
make deps

# Compilar
make

# Executar testes
make test

# Formatar código
make fmt

# Lint
make lint
```

## 📚 Documentação

### **Arquivos de Documentação**
- **README.md**: Este arquivo
- **main.go**: Comentários inline no código
- **common/**: Documentação das funções utilitárias

### **Referências**
- **NPB Original**: http://www.nas.nasa.gov/Software/NPB/
- **NPB-CPP**: https://github.com/GMAP/NPB-CPP
- **NPB-Rust**: Implementação Rust de referência

## 🎯 Características Técnicas

### **Algoritmo**
- **Método**: Gradiente Conjugado
- **Matriz**: Esparsa (CSR format)
- **Convergência**: 25 iterações máximo
- **Verificação**: Norma do resíduo

### **Implementação**
- **Linguagem**: Go
- **Paradigma**: Serial
- **Estrutura**: Modular
- **Performance**: Otimizada para Go

### **Verificação**
- **Métrica**: Zeta (soma das inversas das normas)
- **Referência**: Valores de verificação conhecidos
- **Tolerância**: Precisão dupla

## 🚀 Próximos Passos

### **Melhorias Futuras**
- **Paralelização**: Versão paralela usando gorrotinas
- **Otimizações**: Melhorias de performance
- **Documentação**: Documentação mais detalhada

### **Extensões**
- **Classes Adicionais**: Suporte a mais classes
- **Métricas**: Métricas de performance detalhadas
- **Visualização**: Gráficos de convergência

---

**Desenvolvido por**: Igor Yuji Ishihara Sakuma  
**Baseado em**: NPB-CPP e NPB-Rust  
**Versão**: 1.0  
**Data**: 2024
