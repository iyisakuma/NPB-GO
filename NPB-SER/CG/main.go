package main

import (
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

// makea generates the sparse matrix A
func (cg *CGBenchmark) makea(naa, nzz int, a []float64, colidx []int, rowstr []int,
	firstrow, lastrow, firstcol, lastcol int) {

	// Initialize random number generator
	tran := 314159265.0
	amult := 1220703125.0
	common.Randlc(&tran, amult)

	// Generate matrix elements
	rowstr[0] = 0
	for i := 0; i < naa; i++ {
		rowstr[i+1] = rowstr[i] + NONZER
	}

	// Fill matrix values
	k := 0
	for i := 0; i < naa; i++ {
		for j := 0; j < NONZER; j++ {
			colidx[k] = int(float64(naa) * common.Randlc(&tran, amult))
			if colidx[k] == i {
				colidx[k] = (colidx[k] + 1) % naa
			}
			a[k] = common.Randlc(&tran, amult)
			k++
		}
	}
}

// conj_grad performs conjugate gradient algorithm
func (cg *CGBenchmark) conj_grad(colidx []int, rowstr []int, x []float64, z []float64, a []float64,
	p []float64, q []float64, r []float64, rnorm *float64) {

	cgitmax := 25
	var d, rho, rho0, alpha, beta float64

	// Initialize the CG algorithm
	for i := 0; i < NA; i++ {
		q[i] = 0.0
		z[i] = 0.0
		r[i] = x[i]
		p[i] = r[i]
	}

	// rho = r.r
	rho = 0.0
	for i := 0; i < NA; i++ {
		rho += r[i] * r[i]
	}

	// The conjugate gradient iteration loop
	for cgit := 1; cgit <= cgitmax; cgit++ {
		// q = A.p (matrix-vector multiply)
		for i := 0; i < NA; i++ {
			q[i] = 0.0
			for j := rowstr[i]; j < rowstr[i+1]; j++ {
				if colidx[j] >= 0 && colidx[j] < NA {
					q[i] += a[j] * p[colidx[j]]
				}
			}
		}

		// d = p.q
		d = 0.0
		for i := 0; i < NA; i++ {
			d += p[i] * q[i]
		}

		// alpha = rho / d
		alpha = rho / d

		// Save temporary of rho
		rho0 = rho

		// z = z + alpha*p and r = r - alpha*q
		for i := 0; i < NA; i++ {
			z[i] += alpha * p[i]
			r[i] -= alpha * q[i]
		}

		// rho = r.r
		rho = 0.0
		for i := 0; i < NA; i++ {
			rho += r[i] * r[i]
		}

		// beta = rho / rho0
		beta = rho / rho0

		// p = r + beta*p
		for i := 0; i < NA; i++ {
			p[i] = r[i] + beta*p[i]
		}
	}

	// Compute residual norm explicitly: ||r|| = ||x - A.z||
	// First, form A.z
	for i := 0; i < NA; i++ {
		q[i] = 0.0
		for j := rowstr[i]; j < rowstr[i+1]; j++ {
			if colidx[j] >= 0 && colidx[j] < NA {
				q[i] += a[j] * z[colidx[j]]
			}
		}
	}

	// Compute ||r|| = ||x - A.z||
	*rnorm = 0.0
	for i := 0; i < NA; i++ {
		*rnorm += (x[i] - q[i]) * (x[i] - q[i])
	}
	*rnorm = math.Sqrt(*rnorm)
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
	for j := 0; j < cg.lastcol-cg.firstcol+1; j++ {
		q[j] = 0.0
		z[j] = 0.0
		r[j] = 0.0
		p[j] = 0.0
	}

	zeta = 0.0

	// Main CG loop
	startTime := time.Now()

	for it := 1; it <= NITER; it++ {
		// Perform conjugate gradient
		var rnorm float64
		cg.conj_grad(colidx, rowstr, x, z, a, p, q, r, &rnorm)

		// Update zeta
		zeta += 1.0 / (1.0 + rnorm)
	}

	endTime := time.Now()
	elapsed := endTime.Sub(startTime).Seconds()

	// Calculate Mop/s
	mops := float64(2*NITER*NA) / elapsed / 1e6

	// Verify result
	verified = math.Abs(zeta-zetaVerifyValue) < 1e-10

	// Print results
	common.PrintResults("CG", classNPB, NA, 0, 0, NITER, elapsed, mops, "conjugate gradient", verified, "4.1", "Unknown", "Go", "")
}

func main() {
	// Set problem size based on command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "S":
			NA = 1400
			NZ = 7 * NA
			NITER = 15
			SHIFT = 10.0
			NONZER = 7
			zetaVerifyValue = 8.5971775078648
			classNPB = "S"
		case "W":
			NA = 7000
			NZ = 8 * NA
			NITER = 15
			SHIFT = 12.0
			NONZER = 8
			zetaVerifyValue = 10.362595087124
			classNPB = "W"
		case "A":
			NA = 14000
			NZ = 11 * NA
			NITER = 15
			SHIFT = 20.0
			NONZER = 11
			zetaVerifyValue = 17.130235054029
			classNPB = "A"
		case "B":
			NA = 75000
			NZ = 13 * NA
			NITER = 75
			SHIFT = 60.0
			NONZER = 13
			zetaVerifyValue = 22.712745482631
			classNPB = "B"
		case "C":
			NA = 150000
			NZ = 15 * NA
			NITER = 75
			SHIFT = 110.0
			NONZER = 15
			zetaVerifyValue = 28.973605592845
			classNPB = "C"
		case "D":
			NA = 1500000
			NZ = 21 * NA
			NITER = 100
			SHIFT = 500.0
			NONZER = 21
			zetaVerifyValue = 52.514532105794
			classNPB = "D"
		case "E":
			NA = 9000000
			NZ = 26 * NA
			NITER = 100
			SHIFT = 1500.0
			NONZER = 26
			zetaVerifyValue = 77.522164599383
			classNPB = "E"
		default:
			NA = 1400
			NZ = 7 * NA
			NITER = 15
			SHIFT = 10.0
			NONZER = 7
			zetaVerifyValue = 8.5971775078648
			classNPB = "S"
		}
	} else {
		// Default to class S
		NA = 1400
		NZ = 7 * NA
		NITER = 15
		SHIFT = 10.0
		NONZER = 7
		zetaVerifyValue = 8.5971775078648
		classNPB = "S"
	}

	// Allocate arrays
	a = make([]float64, NZ)
	colidx = make([]int, NZ)
	rowstr = make([]int, NA+1)
	x = make([]float64, NA+1)
	z = make([]float64, NA+1)
	p = make([]float64, NA+1)
	q = make([]float64, NA+1)
	r = make([]float64, NA+1)

	// Create benchmark instance
	cg := NewCGBenchmark()
	cg.naa = NA
	cg.nzz = NZ

	// Run benchmark
	cg.run()
}
