# NPB-GO MG Benchmark

## 📋 Visão Geral

Este é o kernel MG (Multigrid) do NAS Parallel Benchmarks implementado em Go. O MG resolve a equação de Poisson 3D usando o método multigrid.

## 🎯 Características

- **Implementação Serial**: Versão sequencial do algoritmo MG
- **Baseado em C++ e Rust**: Implementação baseada nas versões C++ e Rust existentes
- **Estrutura Modular**: Organização similar aos outros kernels
- **Documentação Completa**: Documentação detalhada da implementação

## 🏗️ Estrutura do Projeto

```
NPB-GO/NPB-SER/MG/
├── main.go              # Implementação principal do MG
├── go.mod              # Módulo Go
└── README.md           # Este arquivo
```

## 🚀 Como Usar

### **Compilação**
```bash
# Compilar versão padrão (classe S)
go build -o mg main.go

# Compilar versão específica
go build -o mg main.go
```

### **Execução**
```bash
# Executar versão padrão
./mg

# Executar com classe específica
./mg S       # Classe S
./mg A       # Classe A
./mg B       # Classe B
```

## 📊 Classes de Problema

| Classe | NX   | NY   | NZ   | NIT  | Descrição |
|-------|------|------|------|------|-----------|
| S     | 32   | 32   | 32   | 4    | Pequena   |
| W     | 64   | 64   | 64   | 4    | Workstation |
| A     | 256  | 256  | 256  | 4    | Média     |
| B     | 256  | 256  | 256  | 20   | Grande    |
| C     | 512  | 512  | 512  | 20   | Muito grande |
| D     | 1024 | 1024 | 1024 | 50   | Enorme   |
| E     | 2048 | 2048 | 2048 | 50   | Extrema  |

## 🔧 Algoritmo MG

### **Método Multigrid 3D**
O algoritmo MG resolve a equação de Poisson 3D usando o método multigrid:

1. **Inicialização**: Condições iniciais
2. **V-Cycle**: Ciclo V do multigrid
3. **Restrição**: Restrição para níveis mais grosseiros
4. **Interpolação**: Interpolação para níveis mais finos
5. **Suavização**: Suavização em cada nível
6. **Verificação**: Cálculo de normas

### **Características**
- **Multigrid**: Método multigrid V-cycle
- **Restrição**: Restrição para níveis grosseiros
- **Interpolação**: Interpolação para níveis finos
- **Suavização**: Suavização em cada nível

## 📈 Resultados Esperados

### **Classe S (Padrão)**
```
 MG Benchmark Completed
 class_npb       =                        S
 Size            =             32x  32x  32
 Iterations      =                        4
 Time in seconds =                     0.00
 Mop/s total     =                  2251.25
 Operation type  =           floating point
 Verification    =            NOT PERFORMED
```

### **Classe A**
```
 MG Benchmark Completed
 class_npb       =                        A
 Size            =            256x 256x 256
 Iterations      =                        4
 Time in seconds =                     0.02
 Mop/s total     =                 21927.19
 Operation type  =           floating point
 Verification    =            NOT PERFORMED
```

## 🛠️ Desenvolvimento

### **Estrutura do Código**
- **main.go**: Implementação principal do algoritmo MG
- **go.mod**: Módulo Go
- **README.md**: Documentação

### **Dependências**
- Go 1.24+
- Módulos Go padrão

### **Compilação**
```bash
# Instalar dependências
go mod tidy

# Compilar
go build -o mg main.go

# Executar
./mg
```

## 📚 Documentação

### **Arquivos de Documentação**
- **README.md**: Este arquivo
- **main.go**: Comentários inline no código

### **Referências**
- **NPB Original**: http://www.nas.nasa.gov/Software/NPB/
- **NPB-CPP**: https://github.com/GMAP/NPB-CPP
- **NPB-Rust**: Implementação Rust de referência

## 🎯 Características Técnicas

### **Algoritmo**
- **Método**: Multigrid V-cycle
- **Níveis**: Múltiplos níveis de resolução
- **Restrição**: Restrição para níveis grosseiros
- **Interpolação**: Interpolação para níveis finos

### **Implementação**
- **Linguagem**: Go
- **Paradigma**: Serial
- **Estrutura**: Modular
- **Performance**: Otimizada para Go

### **Verificação**
- **Métrica**: Normas L2
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
