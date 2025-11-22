# NPB-GO FT Benchmark

## 📋 Visão Geral

Este é o kernel FT (Fourier Transform) do NAS Parallel Benchmarks implementado em Go. O FT resolve a equação de onda 3D usando transformadas de Fourier.

## 🎯 Características

- **Implementação Serial**: Versão sequencial do algoritmo FT
- **Baseado em C++ e Rust**: Implementação baseada nas versões C++ e Rust existentes
- **Estrutura Modular**: Organização similar aos outros kernels
- **Documentação Completa**: Documentação detalhada da implementação

## 🏗️ Estrutura do Projeto

```
NPB-GO/NPB-SER/FT/
├── main.go              # Implementação principal do FT
├── go.mod              # Módulo Go
└── README.md           # Este arquivo
```

## 🚀 Como Usar

### **Compilação**
```bash
# Compilar versão padrão (classe S)
go build -o ft main.go

# Compilar versão específica
go build -o ft main.go
```

### **Execução**
```bash
# Executar versão padrão
./ft

# Executar com classe específica
./ft S       # Classe S
./ft A       # Classe A
./ft B       # Classe B
```

## 📊 Classes de Problema

| Classe | NX   | NY   | NZ   | NITER | Descrição |
|-------|------|------|------|-------|-----------|
| S     | 64   | 64   | 64   | 6     | Pequena   |
| W     | 128  | 128  | 32   | 6     | Workstation |
| A     | 256  | 256  | 128  | 6     | Média     |
| B     | 512  | 256  | 256  | 20    | Grande    |
| C     | 512  | 512  | 512  | 20    | Muito grande |
| D     | 2048 | 1024 | 1024 | 25    | Enorme   |
| E     | 4096 | 2048 | 2048 | 25    | Extrema  |

## 🔧 Algoritmo FT

### **Transformada de Fourier 3D**
O algoritmo FT resolve a equação de onda 3D usando transformadas de Fourier:

1. **Inicialização**: Condições iniciais
2. **FFT Forward**: Transformada de Fourier direta
3. **Evolução**: Evolução temporal no domínio da frequência
4. **FFT Backward**: Transformada de Fourier inversa
5. **Verificação**: Cálculo de checksums

### **Características**
- **FFT 3D**: Transformada de Fourier tridimensional
- **Evolução**: Evolução temporal no domínio da frequência
- **Verificação**: Checksums para verificação

## 📈 Resultados Esperados

### **Classe S (Padrão)**
```
 FT Benchmark Completed
 class_npb       =                        S
 Size            =             64x  64x  64
 Iterations      =                        6
 Time in seconds =                     0.03
 Mop/s total     =                   242.08
 Operation type  =           floating point
 Verification    =            NOT PERFORMED
```

### **Classe A**
```
 FT Benchmark Completed
 class_npb       =                        A
 Size            =            256x 256x 128
 Iterations      =                        6
 Time in seconds =                     1.02
 Mop/s total     =                   246.76
 Operation type  =           floating point
 Verification    =            NOT PERFORMED
```

## 🛠️ Desenvolvimento

### **Estrutura do Código**
- **main.go**: Implementação principal do algoritmo FT
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
go build -o ft main.go

# Executar
./ft
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
- **Método**: Transformada de Fourier 3D
- **Domínio**: Frequência e tempo
- **Evolução**: Evolução temporal
- **Verificação**: Checksums

### **Implementação**
- **Linguagem**: Go
- **Paradigma**: Serial
- **Estrutura**: Modular
- **Performance**: Otimizada para Go

### **Verificação**
- **Métrica**: Checksums
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
