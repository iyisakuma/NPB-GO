package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/iyisakuma/NPB-GO/NPB-SER/common"
)

// Constants
const (
	MAX_ITERATIONS = 100
	MAX_NA         = 1500000
	MAX_NZ         = 27 * MAX_NA
	MAX_NONZER     = 26
)

// Global variables
var (
	// Problem size parameters
	NA     int
	NZ     int
	NITER  int
	SHIFT  float64
	NONZER int

	// Arrays
	a      []float64
	colidx []int
	rowstr []int
	x      []float64
	z      []float64
	p      []float64
	q      []float64
	r      []float64

	// Verification
	zeta            float64
	zetaVerifyValue float64
	classNPB        string
	verified        bool
)

// CGBenchmark represents the CG benchmark
type CGBenchmark struct {
	naa      int
	nzz      int
	firstrow int
	lastrow  int
	firstcol int
	lastcol  int
}

// NewCGBenchmark creates a new CG benchmark instance
func NewCGBenchmark() *CGBenchmark {
	return &CGBenchmark{
		firstrow: 0,
		lastrow:  NA - 1,
		firstcol: 0,
		lastcol:  NA - 1,
	}
}

// icnvrt scales a double precision number x in (0,1) by a power of 2 and chops it
func icnvrt(x float64, ipwr2 int) int {
	return int(float64(ipwr2) * x)
}

// sprnvc generates a sparse n-vector (v, iv) having nzv nonzeros
func sprnvc(n, nz, nn1 int, v []float64, iv []int, tran *float64) {
	nzv := 0
	amult := 1220703125.0

	for nzv < nz {
		vecelt := common.Randlc(tran, amult)
		vecloc := common.Randlc(tran, amult)
		i := icnvrt(vecloc, nn1) + 1
		if i > n {
			continue
		}

		// Check if this integer was already generated
		wasGen := false
		for ii := 0; ii < nzv; ii++ {
			if iv[ii] == i {
				wasGen = true
				break
			}
		}
		if wasGen {
			continue
		}
		v[nzv] = vecelt
		iv[nzv] = i
		nzv++
	}
}

// vecset sets ith element of sparse vector (v, iv) with nzv nonzeros to val
func vecset(n int, v []float64, iv []int, nzv *int, i int, val float64) {
	set := false
	for k := 0; k < *nzv; k++ {
		if iv[k] == i {
			v[k] = val
			set = true
		}
	}
	if !set {
		v[*nzv] = val
		iv[*nzv] = i
		*nzv++
	}
}

// sparse generates a sparse matrix from a list of [col, row, element] triples
func sparse(a []float64, colidx []int, rowstr []int, n int, nz int, nozer int,
	arow []int, acol [][]int, aelt [][]float64, firstrow, lastrow int, nzloc []int, rcond, shift float64) {

	nrows := lastrow - firstrow + 1

	// Count the number of triples in each row
	for j := 0; j < nrows+1; j++ {
		rowstr[j] = 0
	}
	for i := 0; i < n; i++ {
		for nza := 0; nza < arow[i]; nza++ {
			j := acol[i][nza] + 1
			rowstr[j] += arow[i]
		}
	}
	rowstr[0] = 0
	for j := 1; j < nrows+1; j++ {
		rowstr[j] += rowstr[j-1]
	}
	nza := rowstr[nrows] - 1

	if nza > nz {
		fmt.Printf("Space for matrix elements exceeded in sparse\n")
		fmt.Printf("nza, nzmax = %d, %d\n", nza, nz)
		os.Exit(1)
	}

	// Preload data pages
	for j := 0; j < nrows; j++ {
		for k := rowstr[j]; k < rowstr[j+1]; k++ {
			a[k] = 0.0
			colidx[k] = -1
		}
		nzloc[j] = 0
	}

	// Generate actual values by summing duplicates
	size := 1.0
	ratio := math.Pow(rcond, 1.0/float64(n))
	for i := 0; i < n; i++ {
		for nza := 0; nza < arow[i]; nza++ {
			j := acol[i][nza]
			scale := size * aelt[i][nza]
			for nzrow := 0; nzrow < arow[i]; nzrow++ {
				jcol := acol[i][nzrow]
				va := aelt[i][nzrow] * scale

				// Add the identity * rcond to the generated matrix
				if jcol == j && j == i {
					va = va + rcond - shift
				}

				goto40 := false
				k := 0
				for k = rowstr[j]; k < rowstr[j+1]; k++ {
					if colidx[k] > jcol {
						// Insert colidx here orderly
						for kk := rowstr[j+1] - 2; kk >= k; kk-- {
							if colidx[kk] > -1 {
								a[kk+1] = a[kk]
								colidx[kk+1] = colidx[kk]
							}
						}
						colidx[k] = jcol
						a[k] = 0.0
						goto40 = true
						break
					} else if colidx[k] == -1 {
						colidx[k] = jcol
						goto40 = true
						break
					} else if colidx[k] == jcol {
						// Mark the duplicated entry
						nzloc[j]++
						goto40 = true
						break
					}
				}
				if !goto40 {
					fmt.Printf("internal error in sparse: i=%d\n", i)
					os.Exit(1)
				}
				a[k] += va
			}
		}
		size *= ratio
	}

	// Remove empty entries and generate final results
	for j := 1; j < nrows; j++ {
		nzloc[j] += nzloc[j-1]
	}

	for j := 0; j < nrows; j++ {
		j1 := 0
		if j > 0 {
			j1 = rowstr[j] - nzloc[j-1]
		}
		j2 := rowstr[j+1] - nzloc[j]
		nza := rowstr[j]
		for k := j1; k < j2; k++ {
			a[k] = a[nza]
			colidx[k] = colidx[nza]
			nza++
		}
	}
	for j := 1; j < nrows+1; j++ {
		rowstr[j] -= nzloc[j-1]
	}
}

// makea generates the sparse matrix A - complete implementation
func (cg *CGBenchmark) makea(naa, nzz int, a []float64, colidx []int, rowstr []int,
	firstrow, lastrow, firstcol, lastcol int) {

	// Initialize random number generator
	tran := 314159265.0
	amult := 1220703125.0
	common.Randlc(&tran, amult)

	// Allocate workspace arrays
	arow := make([]int, naa)
	acol := make([][]int, naa)
	aelt := make([][]float64, naa)
	for i := 0; i < naa; i++ {
		acol[i] = make([]int, NONZER+1)
		aelt[i] = make([]float64, NONZER+1)
	}
	nzloc := make([]int, lastrow-firstrow+1)

	// nn1 is the smallest power of two not less than n
	nn1 := 1
	for nn1 < naa {
		nn1 *= 2
	}

	// Generate nonzero positions and save for the use in sparse
	for iouter := 0; iouter < naa; iouter++ {
		nzv := NONZER
		ivc := make([]int, NONZER+1)
		vc := make([]float64, NONZER+1)
		sprnvc(naa, nzv, nn1, vc, ivc, &tran)
		vecset(naa, vc, ivc, &nzv, iouter+1, 0.5)
		arow[iouter] = nzv
		for ivelt := 0; ivelt < nzv; ivelt++ {
			acol[iouter][ivelt] = ivc[ivelt] - 1
			aelt[iouter][ivelt] = vc[ivelt]
		}
	}

	// Make the sparse matrix from list of elements with duplicates
	sparse(a, colidx, rowstr, naa, nzz, NONZER, arow, acol, aelt, firstrow, lastrow, nzloc, 0.1, SHIFT)
}

// conj_grad performs conjugate gradient algorithm
func (cg *CGBenchmark) conj_grad(colidx []int, rowstr []int, x []float64, z []float64, a []float64,
	p []float64, q []float64, r []float64, rnorm *float64) {

	cgitmax := 25
	var d, rho, rho0, alpha, beta float64

	// Initialize the CG algorithm
	for i := 0; i < NA+1; i++ {
		q[i] = 0.0
		z[i] = 0.0
		r[i] = x[i]
		p[i] = r[i]
	}

	// rho = r.r
	rho = 0.0
	for i := 0; i < cg.lastcol-cg.firstcol+1; i++ {
		rho += r[i] * r[i]
	}

	// The conjugate gradient iteration loop
	for cgit := 1; cgit <= cgitmax; cgit++ {
		// q = A.p (matrix-vector multiply) - following C++ implementation
		for i := 0; i < NA; i++ {
			sum := 0.0
			for j := rowstr[i]; j < rowstr[i+1]; j++ {
				sum += a[j] * p[colidx[j]]
			}
			q[i] = sum
		}

		// d = p.q
		d = 0.0
		for i := 0; i < cg.lastcol-cg.firstcol+1; i++ {
			d += p[i] * q[i]
		}

		// alpha = rho / d
		if d == 0.0 {
			alpha = 0.0
		} else {
			alpha = rho / d
		}

		// Save temporary of rho
		rho0 = rho

		// z = z + alpha*p and r = r - alpha*q
		rho = 0.0
		for i := 0; i < cg.lastcol-cg.firstcol+1; i++ {
			z[i] = z[i] + alpha*p[i]
			r[i] = r[i] - alpha*q[i]
		}

		// rho = r.r
		rho = 0.0
		for i := 0; i < cg.lastcol-cg.firstcol+1; i++ {
			rho += r[i] * r[i]
		}

		// beta = rho / rho0
		if rho0 == 0.0 {
			beta = 0.0
		} else {
			beta = rho / rho0
		}

		// p = r + beta*p
		for i := 0; i < cg.lastcol-cg.firstcol+1; i++ {
			p[i] = r[i] + beta*p[i]
		}
	}

	// Compute residual norm explicitly: ||r|| = ||x - A.z||
	// First, form A.z
	sum := 0.0
	for i := 0; i < NA; i++ {
		d := 0.0
		for j := rowstr[i]; j < rowstr[i+1]; j++ {
			d += a[j] * z[colidx[j]]
		}
		r[i] = d
	}

	// Compute ||r|| = ||x - A.z||
	for i := 0; i < cg.lastcol-cg.firstcol+1; i++ {
		d := x[i] - r[i]
		sum += d * d
	}
	*rnorm = math.Sqrt(sum)
}

// run performs the CG benchmark
func (cg *CGBenchmark) run() {
	// Initialize arrays
	naa := cg.naa
	nzz := cg.nzz

	// Initialize random number generator
	tran := 314159265.0
	amult := 1220703125.0
	common.Randlc(&tran, amult)

	// Generate matrix
	cg.makea(naa, nzz, a, colidx, rowstr, cg.firstrow, cg.lastrow, cg.firstcol, cg.lastcol)

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

	zeta = 0.0

	// Do one iteration untimed to init all code and data page tables
	for it := 1; it <= 1; it++ {
		// Perform conjugate gradient
		var rnorm float64
		cg.conj_grad(colidx, rowstr, x, z, a, p, q, r, &rnorm)

		// Calculate norm_temp1 = x.z and norm_temp2 = z.z
		norm_temp1 := 0.0
		norm_temp2 := 0.0
		for j := 0; j < cg.lastcol-cg.firstcol+1; j++ {
			norm_temp1 += x[j] * z[j]
			norm_temp2 += z[j] * z[j]
		}
		norm_temp2 = 1.0 / math.Sqrt(norm_temp2)

		// Normalize z to obtain x
		for j := 0; j < cg.lastcol-cg.firstcol+1; j++ {
			x[j] = norm_temp2 * z[j]
		}
	}

	// Set starting vector to (1, 1, ..., 1) again
	for i := 0; i < NA+1; i++ {
		x[i] = 1.0
	}
	zeta = 0.0

	// Main CG loop
	startTime := time.Now()

	for it := 1; it <= NITER; it++ {
		// Perform conjugate gradient
		var rnorm float64
		cg.conj_grad(colidx, rowstr, x, z, a, p, q, r, &rnorm)

		// Calculate norm_temp1 = x.z and norm_temp2 = z.z
		norm_temp1 := 0.0
		norm_temp2 := 0.0
		for j := 0; j < cg.lastcol-cg.firstcol+1; j++ {
			norm_temp1 += x[j] * z[j]
			norm_temp2 += z[j] * z[j]
		}
		norm_temp2 = 1.0 / math.Sqrt(norm_temp2)
		zeta = SHIFT + 1.0/norm_temp1

		if it == 1 {
			fmt.Printf("\n   iteration           ||r||                 zeta\n")
		}
		fmt.Printf("    %5d       %20.14e%20.13e\n", it, rnorm, zeta)

		// Normalize z to obtain x
		for j := 0; j < cg.lastcol-cg.firstcol+1; j++ {
			x[j] = norm_temp2 * z[j]
		}
	}

	endTime := time.Now()
	elapsed := endTime.Sub(startTime).Seconds()

	// Calculate Mop/s using the same formula as C++
	mops := float64(2*NITER*NA) * (3.0 + float64(NONZER*(NONZER+1)) + 25.0*(5.0+float64(NONZER*(NONZER+1))) + 3.0) / elapsed / 1e6

	// Verify result
	verified = math.Abs(zeta-zetaVerifyValue) < 1e-10
	err := math.Abs(zeta-zetaVerifyValue) / zetaVerifyValue

	// Print detailed verification results
	fmt.Printf("\n Benchmark completed\n")
	if verified {
		fmt.Printf(" VERIFICATION SUCCESSFUL\n")
		fmt.Printf(" Zeta is    %20.13e\n", zeta)
		fmt.Printf(" Error is   %20.13e\n", err)
	} else {
		fmt.Printf(" VERIFICATION FAILED\n")
		fmt.Printf(" Zeta                %20.13e\n", zeta)
		fmt.Printf(" The correct zeta is %20.13e\n", zetaVerifyValue)
	}

	// Print results
	common.PrintResults("CG", classNPB, NA, 0, 0, NITER, elapsed, mops, "conjugate gradient", verified, "4.1", "Unknown", "Go", "")
}
