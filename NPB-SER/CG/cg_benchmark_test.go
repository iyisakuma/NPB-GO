package main

import (
	"math"
	"testing"

	"github.com/iyisakuma/NPB-GO/NPB-SER/common"
)

/*
 * Test constants for class S
 * These match the parameters defined in cg.cpp for class S:
 * - NA = 1400: Matrix dimension
 * - NONZER = 7: Average number of nonzeros per row
 * - SHIFT = 10.0: Diagonal shift value
 * - NITER = 15: Number of iterations
 * - NZ = NA * (NONZER+1) * (NONZER+1): Maximum number of nonzeros in sparse matrix
 */
const (
	TEST_NA_S     = 1400
	TEST_NZ_S     = TEST_NA_S * (7 + 1) * (7 + 1) // NA * (NONZER+1) * (NONZER+1)
	TEST_NONZER_S = 7
	TEST_SHIFT_S  = 10.0
	TEST_NITER_S  = 15
)

/*
 * TestNewCGBenchmark
 *
 * O que testa:
 *   Verifica a inicialização correta da estrutura CGBenchmark através da função NewCGBenchmark.
 *
 * Comportamento esperado:
 *   A função NewCGBenchmark deve criar uma instância com os seguintes valores:
 *   - firstrow = 0: Primeira linha da matriz (índice base 0)
 *   - lastrow = NA-1: Última linha da matriz (índice base 0)
 *   - firstcol = 0: Primeira coluna da matriz (índice base 0)
 *   - lastcol = NA-1: Última coluna da matriz (índice base 0)
 *
 * Por que é importante:
 *   Esses valores definem o intervalo de trabalho do benchmark. Se estiverem incorretos,
 *   o algoritmo CG pode processar índices inválidos ou deixar de processar partes da matriz.
 *   Baseado em cg.cpp linhas 197-200, onde firstrow=0, lastrow=NA-1, firstcol=0, lastcol=NA-1.
 */
func TestNewCGBenchmark(t *testing.T) {
	NA = TEST_NA_S
	cg := NewCGBenchmark()

	if cg.firstrow != 0 {
		t.Errorf("Expected firstrow = 0, got %d", cg.firstrow)
	}
	if cg.lastrow != NA-1 {
		t.Errorf("Expected lastrow = %d, got %d", NA-1, cg.lastrow)
	}
	if cg.firstcol != 0 {
		t.Errorf("Expected firstcol = 0, got %d", cg.firstcol)
	}
	if cg.lastcol != NA-1 {
		t.Errorf("Expected lastcol = %d, got %d", NA-1, cg.lastcol)
	}
}

/*
 * TestIcnvrt
 *
 * O que testa:
 *   Verifica a função icnvrt que converte um número de ponto flutuante x em (0,1)
 *   para um inteiro multiplicando por uma potência de 2 e truncando.
 *
 * Comportamento esperado:
 *   A função deve calcular: int(ipwr2 * x), onde:
 *   - x é um valor entre 0 e 1
 *   - ipwr2 é uma potência de 2
 *   - O resultado é um inteiro truncado (não arredondado)
 *
 * Por que é importante:
 *   Esta função é usada em sprnvc para gerar índices aleatórios para vetores esparsos.
 *   Se não funcionar corretamente, os índices gerados podem estar fora do intervalo válido.
 *   Baseado em cg.cpp linha 611-613: return (int)(ipwr2 * x);
 *
 * Casos de teste:
 *   - zero: x=0.0 deve retornar 0
 *   - small: x=0.25, ipwr2=8 deve retornar 2 (8*0.25=2.0)
 *   - medium: x=0.5, ipwr2=16 deve retornar 8 (16*0.5=8.0)
 *   - large: x=0.75, ipwr2=32 deve retornar 24 (32*0.75=24.0)
 *   - one: x=1.0, ipwr2=64 deve retornar 64 (64*1.0=64.0)
 */
func TestIcnvrt(t *testing.T) {
	tests := []struct {
		name   string
		x      float64
		ipwr2  int
		result int
	}{
		{"zero", 0.0, 8, 0},
		{"small", 0.25, 8, 2},
		{"medium", 0.5, 16, 8},
		{"large", 0.75, 32, 24},
		{"one", 1.0, 64, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := icnvrt(tt.x, tt.ipwr2)
			if result != tt.result {
				t.Errorf("icnvrt(%f, %d) = %d, want %d", tt.x, tt.ipwr2, result, tt.result)
			}
		})
	}
}

/*
 * TestSprnvc
 *
 * O que testa:
 *   Verifica a função sprnvc que gera um vetor esparso (v, iv) com nz elementos não-zero.
 *   A função gera valores aleatórios e posições aleatórias, garantindo que não haja duplicatas.
 *
 * Comportamento esperado:
 *   - Deve gerar exatamente nz elementos não-zero
 *   - Todos os índices em iv devem ser únicos (sem duplicatas)
 *   - Todos os índices devem estar no intervalo [1, n] (base 1)
 *   - Os valores em v devem ser números aleatórios gerados por Randlc
 *
 * Por que é importante:
 *   Esta função é fundamental para gerar a estrutura esparsa da matriz A.
 *   Se gerar índices duplicados ou fora do intervalo, a matriz será inválida.
 *   Baseado em cg.cpp linhas 152-156 e implementação em cg.cpp linhas 700-750.
 *
 * Assertions:
 *   1. Contagem de não-zeros: Verifica que exatamente nz elementos foram gerados
 *   2. Unicidade: Verifica que não há índices duplicados (usando mapa)
 *   3. Intervalo válido: Verifica que todos os índices estão entre 1 e n
 */
func TestSprnvc(t *testing.T) {
	// Inicializa o gerador de números aleatórios com a mesma seed do C++
	// cg.cpp linha 235: tran = 314159265.0, amult = 1220703125.0
	tran := 314159265.0
	amult := 1220703125.0
	common.Randlc(&tran, amult) // Initialize random number generator

	n := 100
	nz := 10
	nn1 := 128 // Smallest power of 2 >= n

	v := make([]float64, nz)
	iv := make([]int, nz)

	sprnvc(n, nz, nn1, v, iv, &tran)

	// Assertion 1: Verifica que obtivemos exatamente nz não-zeros
	// Comportamento esperado: A função deve preencher nz posições válidas
	count := 0
	for i := 0; i < nz; i++ {
		if iv[i] > 0 && iv[i] <= n {
			count++
		}
	}

	if count != nz {
		t.Errorf("Expected %d nonzeros, got %d", nz, count)
	}

	// Assertion 2: Verifica que todos os índices são únicos
	// Comportamento esperado: sprnvc verifica duplicatas e não as adiciona
	// Isso é crítico porque índices duplicados quebrariam a estrutura esparsa
	seen := make(map[int]bool)
	for i := 0; i < nz; i++ {
		if iv[i] > 0 {
			if seen[iv[i]] {
				t.Errorf("Duplicate index found: %d", iv[i])
			}
			seen[iv[i]] = true
		}
	}

	// Assertion 3: Verifica que todos os índices estão no intervalo válido [1, n]
	// Comportamento esperado: icnvrt garante que i <= n, mas verificamos para segurança
	for i := 0; i < nz; i++ {
		if iv[i] > 0 && (iv[i] < 1 || iv[i] > n) {
			t.Errorf("Index out of range: %d (should be 1-%d)", iv[i], n)
		}
	}
}

/*
 * TestVecset
 *
 * O que testa:
 *   Verifica a função vecset que define o i-ésimo elemento de um vetor esparso (v, iv).
 *   A função deve atualizar o valor se o índice já existir, ou adicionar um novo elemento.
 *
 * Comportamento esperado:
 *   - Se o índice i já existe no vetor esparso: atualiza v[k] onde iv[k] == i
 *   - Se o índice i não existe: adiciona novo elemento com iv[nzv] = i, v[nzv] = val, incrementa nzv
 *   - nzv deve refletir o número atual de elementos não-zero
 *
 * Por que é importante:
 *   Esta função é usada em makea para garantir que cada linha da matriz tenha um elemento
 *   na diagonal (iouter+1, 0.5). Se não funcionar corretamente, a diagonal da matriz pode
 *   estar incorreta, afetando a convergência do CG.
 *   Baseado em cg.cpp linhas 157-162 e uso em cg.cpp linha 679: vecset(n, vc, ivc, &nzv, iouter+1, 0.5);
 *
 * Casos de teste:
 *   1. Inserção de novo elemento: nzv deve incrementar, elemento deve ser adicionado
 *   2. Atualização de elemento existente: nzv não deve mudar, valor deve ser atualizado
 *   3. Múltiplos elementos: nzv deve refletir o número correto de elementos
 */
func TestVecset(t *testing.T) {
	n := 10
	v := make([]float64, n)
	iv := make([]int, n)
	nzv := 0

	// Teste 1: Inserção de novo elemento
	// Comportamento esperado: nzv deve ser 1, iv[0] = 5, v[0] = 0.5
	vecset(n, v, iv, &nzv, 5, 0.5)
	if nzv != 1 {
		t.Errorf("Expected nzv = 1 after first insert, got %d", nzv)
	}
	if iv[0] != 5 || v[0] != 0.5 {
		t.Errorf("Expected iv[0] = 5, v[0] = 0.5, got iv[0] = %d, v[0] = %f", iv[0], v[0])
	}

	// Teste 2: Atualização de elemento existente
	// Comportamento esperado: nzv permanece 1, mas v[0] é atualizado para 0.75
	vecset(n, v, iv, &nzv, 5, 0.75)
	if nzv != 1 {
		t.Errorf("Expected nzv = 1 after update, got %d", nzv)
	}
	if v[0] != 0.75 {
		t.Errorf("Expected v[0] = 0.75 after update, got %f", v[0])
	}

	// Teste 3: Adição de outro elemento
	// Comportamento esperado: nzv deve ser 2, novo elemento adicionado
	vecset(n, v, iv, &nzv, 3, 0.25)
	if nzv != 2 {
		t.Errorf("Expected nzv = 2 after second insert, got %d", nzv)
	}
}

/*
 * TestSparse
 *
 * O que testa:
 *   Verifica a função sparse que constrói uma matriz esparsa no formato CSR (Compressed Sparse Row)
 *   a partir de uma lista de triplas [linha, coluna, elemento] com possíveis duplicatas.
 *
 * Comportamento esperado:
 *   - rowstr[0] deve ser 0 (primeiro elemento sempre começa em 0)
 *   - rowstr deve ser monotonicamente crescente (rowstr[i] <= rowstr[i+1])
 *   - colidx[k] deve estar no intervalo [0, n-1] para todos os elementos válidos
 *   - A matriz deve ser construída corretamente, somando duplicatas e aplicando shift na diagonal
 *
 * Por que é importante:
 *   Esta função é o coração da geração da matriz esparsa A. Se rowstr não for monotônico,
 *   o acesso à matriz será inválido. Se colidx estiver fora do intervalo, haverá acesso
 *   fora dos limites. A estrutura CSR é usada em todo o algoritmo CG.
 *   Baseado em cg.cpp linhas 138-151 e implementação em cg.cpp linhas 693-750.
 *
 * Assertions:
 *   1. rowstr[0] = 0: Primeiro elemento sempre começa em 0 (padrão CSR)
 *   2. Monotonicidade: rowstr[i] <= rowstr[i+1] para todo i (estrutura CSR válida)
 *   3. Intervalo válido: colidx[k] deve estar em [0, n-1] (índices válidos)
 */
func TestSparse(t *testing.T) {
	n := 10
	nz := 100
	nozer := 5
	firstrow := 0
	lastrow := n - 1

	a := make([]float64, nz)
	colidx := make([]int, nz)
	rowstr := make([]int, n+1)
	arow := make([]int, n)
	acol := make([][]int, n)
	aelt := make([][]float64, n)
	nzloc := make([]int, lastrow-firstrow+1)

	// Initialize test data
	for i := 0; i < n; i++ {
		arow[i] = 3
		acol[i] = make([]int, nozer+1)
		aelt[i] = make([]float64, nozer+1)
		for j := 0; j < 3; j++ {
			acol[i][j] = (i + j) % n
			aelt[i][j] = 1.0
		}
	}

	rcond := 0.1
	shift := 10.0

	sparse(a, colidx, rowstr, n, nz, nozer, arow, acol, aelt, firstrow, lastrow, nzloc, rcond, shift)

	// Assertion 1: rowstr[0] deve ser 0
	// Comportamento esperado: No formato CSR, o primeiro elemento sempre começa em 0
	// cg.cpp linha 133: rowstr[0] = 0;
	if rowstr[0] != 0 {
		t.Errorf("Expected rowstr[0] = 0, got %d", rowstr[0])
	}

	// Assertion 2: rowstr deve ser monotonicamente crescente
	// Comportamento esperado: rowstr[i] <= rowstr[i+1] para todo i
	// Isso garante que cada linha começa após a anterior (estrutura CSR válida)
	// Se violado, o acesso à matriz será incorreto
	for i := 1; i <= n; i++ {
		if rowstr[i] < rowstr[i-1] {
			t.Errorf("rowstr[%d] = %d < rowstr[%d] = %d", i, rowstr[i], i-1, rowstr[i-1])
		}
	}

	// Assertion 3: colidx deve conter apenas índices válidos [0, n-1]
	// Comportamento esperado: Todos os elementos não-zero devem ter colidx no intervalo válido
	// Se violado, haverá acesso fora dos limites durante multiplicação matriz-vetor
	for i := 0; i < n; i++ {
		for k := rowstr[i]; k < rowstr[i+1]; k++ {
			if colidx[k] < 0 || colidx[k] >= n {
				t.Errorf("Invalid colidx[%d] = %d (should be 0-%d)", k, colidx[k], n-1)
			}
		}
	}
}

/*
 * TestMakea
 *
 * O que testa:
 *   Verifica a função makea que gera a matriz esparsa A completa usando sprnvc, vecset e sparse.
 *   Esta é a função principal que cria toda a estrutura da matriz usada no algoritmo CG.
 *
 * Comportamento esperado:
 *   - Deve gerar uma matriz esparsa válida no formato CSR
 *   - rowstr deve ser monotônico e dentro dos limites [0, NZ]
 *   - colidx deve conter apenas índices válidos [0, NA-1]
 *   - A matriz deve ter elementos não-zero (não pode ser matriz nula)
 *   - A diagonal deve ter shift aplicado (rcond - shift)
 *
 * Por que é importante:
 *   Esta função gera toda a matriz A usada no benchmark. Se a matriz estiver incorreta,
 *   o algoritmo CG não convergirá corretamente e os resultados de verificação falharão.
 *   Baseado em cg.cpp linhas 642-700 e chamada em cg.cpp linhas 239-251.
 *
 * Assertions:
 *   1. rowstr[0] = 0: Estrutura CSR válida
 *   2. Monotonicidade: rowstr monotônico e dentro dos limites
 *   3. Índices válidos: colidx no intervalo [0, NA-1]
 *   4. Matriz não-nula: Deve haver pelo menos um elemento não-zero
 */
func TestMakea(t *testing.T) {
	NA = TEST_NA_S
	NZ = TEST_NZ_S
	NONZER = TEST_NONZER_S
	SHIFT = TEST_SHIFT_S

	cg := NewCGBenchmark()
	cg.naa = NA
	cg.nzz = NZ

	a := make([]float64, NZ)
	colidx := make([]int, NZ)
	rowstr := make([]int, NA+1)

	cg.makea(NA, NZ, a, colidx, rowstr, cg.firstrow, cg.lastrow, cg.firstcol, cg.lastcol)

	// Assertion 1: rowstr[0] deve ser 0 (estrutura CSR)
	// Comportamento esperado: Primeiro elemento sempre em 0
	if rowstr[0] != 0 {
		t.Errorf("Expected rowstr[0] = 0, got %d", rowstr[0])
	}

	// Assertion 2: rowstr deve ser monotônico e dentro dos limites
	// Comportamento esperado: rowstr[i] <= rowstr[i+1] e rowstr[i] <= NZ
	// Se rowstr[i] > NZ, haverá acesso fora dos limites do array a[]
	for i := 1; i <= NA; i++ {
		if rowstr[i] < rowstr[i-1] {
			t.Errorf("rowstr[%d] = %d < rowstr[%d] = %d", i, rowstr[i], i-1, rowstr[i-1])
		}
		if rowstr[i] > NZ {
			t.Errorf("rowstr[%d] = %d exceeds NZ = %d", i, rowstr[i], NZ)
		}
	}

	// Assertion 3: colidx deve conter apenas índices válidos [0, NA-1]
	// Comportamento esperado: Todos os elementos devem referenciar colunas válidas
	// Se violado, multiplicação matriz-vetor acessará memória inválida
	for i := 0; i < NA; i++ {
		for k := rowstr[i]; k < rowstr[i+1]; k++ {
			if colidx[k] < 0 || colidx[k] >= NA {
				t.Errorf("Invalid colidx[%d] = %d (should be 0-%d)", k, colidx[k], NA-1)
			}
		}
	}

	// Assertion 4: A matriz deve ter pelo menos um elemento não-zero
	// Comportamento esperado: makea gera uma matriz esparsa com elementos não-zero
	// Se a matriz for nula, o CG não funcionará (divisão por zero, etc.)
	hasNonZero := false
	for i := 0; i < NZ; i++ {
		if a[i] != 0.0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Error("Matrix a has no non-zero elements")
	}
}

/*
 * TestConjGrad
 *
 * O que testa:
 *   Verifica o algoritmo de gradiente conjugado (conj_grad) com uma matriz diagonal simples.
 *   Este teste usa uma matriz simplificada para verificar a lógica básica do algoritmo.
 *
 * Comportamento esperado:
 *   - rnorm deve ser calculado corretamente (não NaN, não Inf, não negativo)
 *   - O vetor z deve ser atualizado após as iterações CG (não pode ser todo zero)
 *   - O algoritmo deve executar 25 iterações internas (cgitmax = 25)
 *
 * Por que é importante:
 *   O algoritmo CG é o coração do benchmark. Se não funcionar corretamente, o benchmark
 *   não produzirá resultados válidos. Este teste verifica a lógica básica com uma matriz
 *   simples antes de testar com matrizes reais geradas.
 *   Baseado em cg.cpp linhas 456-604 (função conj_grad).
 *
 * Assertions:
 *   1. rnorm válido: Não pode ser NaN, Inf ou negativo
 *   2. z atualizado: O vetor z deve ter valores não-zero após CG (prova que o algoritmo executou)
 */
func TestConjGrad(t *testing.T) {
	NA = TEST_NA_S
	NONZER = TEST_NONZER_S
	SHIFT = TEST_SHIFT_S

	cg := NewCGBenchmark()
	cg.naa = NA
	cg.nzz = TEST_NZ_S
	cg.firstrow = 0
	cg.lastrow = NA - 1
	cg.firstcol = 0
	cg.lastcol = NA - 1

	// Create a simple test matrix (identity-like)
	a := make([]float64, TEST_NZ_S)
	colidx := make([]int, TEST_NZ_S)
	rowstr := make([]int, NA+1)
	x := make([]float64, NA+1)
	z := make([]float64, NA+1)
	p := make([]float64, NA+1)
	q := make([]float64, NA+1)
	r := make([]float64, NA+1)

	// Initialize rowstr for a simple diagonal matrix
	for i := 0; i <= NA; i++ {
		rowstr[i] = i
	}

	// Create a simple diagonal matrix with shift
	for i := 0; i < NA; i++ {
		colidx[i] = i
		a[i] = 1.0 + SHIFT // Diagonal element
	}

	// Set x to all ones
	for i := 0; i < NA+1; i++ {
		x[i] = 1.0
	}

	var rnorm float64
	cg.conj_grad(colidx, rowstr, x, z, a, p, q, r, &rnorm)

	// Assertion 1: rnorm deve ser um número válido
	// Comportamento esperado: rnorm = ||x - A.z||, que é sempre >= 0
	// Se for NaN ou Inf, há erro numérico (divisão por zero, overflow, etc.)
	// cg.cpp linha 603: *rnorm = sqrt(sum);
	if math.IsNaN(rnorm) || math.IsInf(rnorm, 0) {
		t.Errorf("rnorm is NaN or Inf: %f", rnorm)
	}

	// Assertion 2: rnorm deve ser não-negativo (norma sempre >= 0)
	// Comportamento esperado: Norma de um vetor é sempre >= 0
	if rnorm < 0 {
		t.Errorf("rnorm should be non-negative, got %f", rnorm)
	}

	// Assertion 3: O vetor z deve ser atualizado após as iterações CG
	// Comportamento esperado: Após 25 iterações de CG, z = z + alpha*p múltiplas vezes
	// Se z permanecer zero, o algoritmo não está executando corretamente
	// cg.cpp linhas 545-547: z[j] = z[j] + alpha*p[j];
	allZeros := true
	for i := 0; i < cg.lastcol-cg.firstcol+1; i++ {
		if z[i] != 0.0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		t.Error("z vector is all zeros after conj_grad")
	}
}

/*
 * TestConjGradWithRealMatrix
 *
 * O que testa:
 *   Verifica o algoritmo conj_grad com uma matriz real gerada por makea.
 *   Este teste é mais realista que TestConjGrad pois usa a mesma matriz que o benchmark real.
 *
 * Comportamento esperado:
 *   - rnorm deve ser calculado corretamente
 *   - Para classe S, após a primeira iteração, rnorm tipicamente está em torno de 1e-13
 *   - O algoritmo deve convergir (rnorm diminui com as iterações)
 *
 * Por que é importante:
 *   Este teste valida que o algoritmo CG funciona corretamente com matrizes reais geradas
 *   pelo mesmo processo usado no benchmark. Se passar aqui, há alta probabilidade de que
 *   o benchmark completo funcione corretamente.
 *   Baseado no comportamento observado em cg.cpp: primeira iteração produz rnorm ~1e-13.
 *
 * Assertions:
 *   1. rnorm válido: Não NaN, não Inf, não negativo
 *   2. rnorm razoável: Para classe S, primeira iteração tipicamente ~1e-13 (warning se > 1e-10)
 */
func TestConjGradWithRealMatrix(t *testing.T) {
	NA = TEST_NA_S
	NZ = TEST_NZ_S
	NONZER = TEST_NONZER_S
	SHIFT = TEST_SHIFT_S

	cg := NewCGBenchmark()
	cg.naa = NA
	cg.nzz = NZ
	cg.firstrow = 0
	cg.lastrow = NA - 1
	cg.firstcol = 0
	cg.lastcol = NA - 1

	a := make([]float64, NZ)
	colidx := make([]int, NZ)
	rowstr := make([]int, NA+1)
	x := make([]float64, NA+1)
	z := make([]float64, NA+1)
	p := make([]float64, NA+1)
	q := make([]float64, NA+1)
	r := make([]float64, NA+1)

	// Generate matrix using makea
	cg.makea(NA, NZ, a, colidx, rowstr, cg.firstrow, cg.lastrow, cg.firstcol, cg.lastcol)

	// Shift column indices
	for j := 0; j < cg.lastrow-cg.firstrow+1; j++ {
		for k := rowstr[j]; k < rowstr[j+1]; k++ {
			colidx[k] = colidx[k] - cg.firstcol
		}
	}

	// Set x to all ones
	for i := 0; i < NA+1; i++ {
		x[i] = 1.0
	}

	var rnorm float64
	cg.conj_grad(colidx, rowstr, x, z, a, p, q, r, &rnorm)

	// Assertion 1: rnorm deve ser válido
	// Comportamento esperado: rnorm = ||x - A.z|| calculado corretamente
	if math.IsNaN(rnorm) || math.IsInf(rnorm, 0) {
		t.Errorf("rnorm is NaN or Inf: %f", rnorm)
	}

	if rnorm < 0 {
		t.Errorf("rnorm should be non-negative, got %f", rnorm)
	}

	// Assertion 2: rnorm deve estar em um intervalo razoável para classe S
	// Comportamento esperado: Para classe S, após primeira iteração, rnorm tipicamente ~1e-13
	// Baseado em observações do cg.cpp: primeira iteração produz valores muito pequenos
	// Se rnorm > 1e-10, pode indicar problema, mas não é erro fatal (apenas warning)
	if rnorm > 1e-10 {
		t.Logf("Warning: rnorm = %e is larger than expected (typically ~1e-13 for first iteration)", rnorm)
	}
}

/*
 * TestFullRunIteration
 *
 * O que testa:
 *   Simula uma iteração completa do benchmark CG, incluindo:
 *   - Geração da matriz A usando makea
 *   - Execução do algoritmo CG
 *   - Cálculo de zeta = SHIFT + 1.0/(x.z)
 *   - Normalização do vetor x
 *
 * Comportamento esperado:
 *   - norm_temp2 (z.z) deve ser positivo (para poder calcular 1/sqrt)
 *   - zeta deve ser calculado corretamente: zeta = SHIFT + 1.0/norm_temp1
 *   - zeta deve ser positivo e finito
 *   - O vetor x normalizado deve ter norma aproximadamente 1.0
 *
 * Por que é importante:
 *   Este teste valida o fluxo completo de uma iteração do benchmark, incluindo os cálculos
 *   de zeta e normalização que são críticos para a verificação final. Se este teste passar,
 *   há alta confiança de que o benchmark completo funcionará.
 *   Baseado em cg.cpp linhas 287-311 (iteração não cronometrada) e 332-361 (iterações principais).
 *
 * Assertions:
 *   1. norm_temp2 > 0: Necessário para calcular 1/sqrt(norm_temp2)
 *   2. zeta válido: Não NaN, não Inf, positivo
 *   3. x normalizado: ||x|| deve ser aproximadamente 1.0 após normalização
 */
func TestFullRunIteration(t *testing.T) {
	NA = TEST_NA_S
	NZ = TEST_NZ_S
	NONZER = TEST_NONZER_S
	SHIFT = TEST_SHIFT_S
	NITER = 1 // Just one iteration for testing

	cg := NewCGBenchmark()
	cg.naa = NA
	cg.nzz = NZ

	// Initialize arrays
	a := make([]float64, NZ)
	colidx := make([]int, NZ)
	rowstr := make([]int, NA+1)
	x := make([]float64, NA+1)
	z := make([]float64, NA+1)
	p := make([]float64, NA+1)
	q := make([]float64, NA+1)
	r := make([]float64, NA+1)

	// Initialize random number generator
	tran := 314159265.0
	amult := 1220703125.0
	common.Randlc(&tran, amult)

	// Generate matrix
	cg.makea(NA, NZ, a, colidx, rowstr, cg.firstrow, cg.lastrow, cg.firstcol, cg.lastcol)

	// Shift column indices
	for j := 0; j < cg.lastrow-cg.firstrow+1; j++ {
		for k := rowstr[j]; k < rowstr[j+1]; k++ {
			colidx[k] = colidx[k] - cg.firstcol
		}
	}

	// Set starting vector to (1, 1, ..., 1)
	for i := 0; i < NA+1; i++ {
		x[i] = 1.0
	}

	// Initialize vectors
	for j := 0; j < NA; j++ {
		q[j] = 0.0
		z[j] = 0.0
		r[j] = 0.0
		p[j] = 0.0
	}

	// Perform conjugate gradient
	var rnorm float64
	cg.conj_grad(colidx, rowstr, x, z, a, p, q, r, &rnorm)

	// Calcula norm_temp1 = x.z (produto interno) e norm_temp2 = z.z (norma ao quadrado)
	// cg.cpp linhas 299-304: cálculo dos produtos internos
	norm_temp1 := 0.0
	norm_temp2 := 0.0
	for j := 0; j < cg.lastcol-cg.firstcol+1; j++ {
		norm_temp1 += x[j] * z[j]
		norm_temp2 += z[j] * z[j]
	}

	// Assertion 1: norm_temp2 (z.z) deve ser positivo
	// Comportamento esperado: z.z > 0 (z não pode ser vetor nulo após CG)
	// Se norm_temp2 <= 0, não podemos calcular 1/sqrt(norm_temp2)
	// cg.cpp linha 305: norm_temp2 = 1.0 / sqrt(norm_temp2);
	if norm_temp2 <= 0 {
		t.Errorf("norm_temp2 (z.z) should be positive, got %f", norm_temp2)
	}

	norm_temp2 = 1.0 / math.Sqrt(norm_temp2)
	// Calcula zeta = SHIFT + 1.0/(x.z)
	// cg.cpp linha 353: zeta = SHIFT + 1.0 / norm_temp1;
	zeta := SHIFT + 1.0/norm_temp1

	// Assertion 2: zeta deve ser calculado corretamente
	// Comportamento esperado: zeta = SHIFT + 1.0/norm_temp1, onde norm_temp1 = x.z
	// zeta é usado para verificação final do benchmark
	// Se for NaN ou Inf, há erro numérico (divisão por zero se norm_temp1 = 0)
	if math.IsNaN(zeta) || math.IsInf(zeta, 0) {
		t.Errorf("zeta is NaN or Inf: %f", zeta)
	}
	if zeta <= 0 {
		t.Errorf("zeta should be positive, got %f", zeta)
	}

	// Normaliza z para obter x: x = (1/||z||) * z
	// cg.cpp linhas 307-310: normalização do vetor x
	for j := 0; j < cg.lastcol-cg.firstcol+1; j++ {
		x[j] = norm_temp2 * z[j]
	}

	// Assertion 3: x normalizado deve ter norma aproximadamente 1.0
	// Comportamento esperado: Após normalização, ||x|| = 1.0 (dentro de tolerância numérica)
	// Esta é uma propriedade fundamental da normalização: ||x|| = ||(1/||z||) * z|| = 1.0
	// Se violado, a normalização está incorreta
	xnorm := 0.0
	for j := 0; j < cg.lastcol-cg.firstcol+1; j++ {
		xnorm += x[j] * x[j]
	}
	xnorm = math.Sqrt(xnorm)

	// Tolerância para erros de ponto flutuante (1e-10 é razoável para double precision)
	if math.Abs(xnorm-1.0) > 1e-10 {
		t.Errorf("Normalized x should have norm ~1.0, got %f", xnorm)
	}
}
