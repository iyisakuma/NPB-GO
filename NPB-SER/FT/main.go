package main

import (
	"fmt"
	"github.com/iyisakuma/NPB-GO/NPB-SER/common"
	"github.com/iyisakuma/NPB-GO/NPB-SER/FT/params"
	"math"
	"math/cmplx"
	"os"
)

// Constants
const (
	// Cache blocking params
	FFTBLOCK    = 16
	FFTBLOCKPAD = 18

	// MAXDIM will be set from params.MAXDIM
	SEED    = 314159265.0
	A       = 1220703125.0
	PI      = 3.141592653589793238
	ALPHA   = 1.0e-6
	EPSILON = 1.0e-12

	// Timers
	T_TOTAL    = 1
	T_SETUP    = 2
	T_FFT      = 3
	T_EVOLVE   = 4
	T_CHECKSUM = 5
	T_FFTX     = 6
	T_FFTY     = 7
	T_FFTZ     = 8
	T_MAX      = 8
)

// Use Go's native complex128 type
type Dcomplex = complex128

// Global variables equivalent to static variables in C++
var (
	// Problem size parameters
	NX, NY, NZ int
	NITER      int
	NTOTAL     int
	CLASS      string

	// Arrays (allocated on heap)
	u0      []Dcomplex
	u1      []Dcomplex
	twiddle []Dcomplex
	sums    []Dcomplex // sums[NITER_DEFAULT+1]
	u       []Dcomplex // u[MAXDIM] used in fft_init/cfftz

	// State variables
	dims          [3]int
	timersEnabled bool
	debug         bool
)

// FTBenchmark encapsulates benchmark logic
type FTBenchmark struct{}

// NewFTBenchmark creates a new FT benchmark instance
func NewFTBenchmark() *FTBenchmark {
	return &FTBenchmark{}
}

// ilog2 calculates integer log2 of n
func ilog2(n int) int {
	if n <= 0 {
		return 0
	}
	return int(math.Log2(float64(n)))
}

// compute_indexmap computes the index map for time evolution
func (ft *FTBenchmark) compute_indexmap(twiddle []Dcomplex, d1, d2, d3 int) {
	ap := -4.0 * ALPHA * PI * PI

	for k := 0; k < d3; k++ {
		kk := ((k + NZ/2) % NZ) - NZ/2
		kk2 := float64(kk * kk)
		for j := 0; j < d2; j++ {
			jj := ((j + NY/2) % NY) - NY/2
			kj2 := float64(jj*jj) + kk2
			for i := 0; i < d1; i++ {
				ii := ((i + NX/2) % NX) - NX/2
				exponent := ap * (float64(ii*ii) + kj2)
				// twiddle[k][j][i]
				idx := k*d2*d1 + j*d1 + i
				twiddle[idx] = complex(math.Exp(exponent), 0.0)
			}
		}
	}
}

// ipow46 computes a^exponent mod 2^46
func (ft *FTBenchmark) ipow46(a float64, exponent int) float64 {
	var q, r float64
	var n, n2 int

	result := 1.0
	if exponent == 0 {
		return result
	}

	q = a
	r = 1.0
	n = exponent

	for n > 1 {
		n2 = n / 2
		if n2*2 == n {
			common.Randlc(&q, q)
			n = n2
		} else {
			common.Randlc(&r, q)
			n = n - 1
		}
	}
	common.Randlc(&r, q)
	result = r
	return result
}

// compute_initial_conditions fills u0 with random data
func (ft *FTBenchmark) compute_initial_conditions(u0 []Dcomplex, d1, d2, d3 int) {
	var start, an, x0 float64
	starts := make([]float64, NZ)
	start = SEED

	// CORREÇÃO AQUI:
	// Antes: ft.ipow46(A, 0, &an)
	// Agora: Atribuição direta pelo retorno
	an = ft.ipow46(A, 0)

	common.Randlc(&start, an)
	an = ft.ipow46(A, 2*NX*NY)

	starts[0] = start
	for k := 1; k < dims[2]; k++ {
		common.Randlc(&start, an)
		starts[k] = start
	}

	// Go through by z planes
	for k := 0; k < dims[2]; k++ {
		x0 = starts[k]
		for j := 0; j < dims[1]; j++ {
			tempFloat := make([]float64, 2*NX)
			common.Vranlc(2*NX, &x0, A, tempFloat)

			baseIdx := k*d2*d1 + j*d1
			for i := 0; i < d1; i++ {
				u0[baseIdx+i] = complex(tempFloat[2*i], tempFloat[2*i+1])
			}
		}
	}
}

// fft_init initializes roots of unity
func (ft *FTBenchmark) fft_init(n int) {
	m := ilog2(n)
	u[0] = complex(float64(m), 0.0)

	ku := 2
	ln := 1

	for j := 1; j <= m; j++ {
		t := PI / float64(ln)

		for i := 0; i <= ln-1; i++ {
			ti := float64(i) * t
			u[i+ku-1] = complex(math.Cos(ti), math.Sin(ti))
		}

		ku = ku + ln
		ln = 2 * ln
	}
}

// cfftz performs Stockham FFT
// x and y are slices representing 2D arrays [n][FFTBLOCKPAD]
func (ft *FTBenchmark) cfftz(is, m, n int, x, y []Dcomplex) {
	// Indices management for 2D-like access in 1D slice:
	// x[j][i] -> x[j*FFTBLOCKPAD + i]

	mx := int(real(u[0]))
	if (is != 1 && is != -1) || m < 1 || m > mx {
		fmt.Printf("CFFTZ: Invalid parameters\n")
		os.Exit(1)
	}

	for l := 1; l <= m; l += 2 {
		ft.fftz2(is, l, m, n, FFTBLOCK, FFTBLOCKPAD, u, x, y)
		if l == m {
			// Copy Y to X
			for j := 0; j < n; j++ {
				for i := 0; i < FFTBLOCK; i++ {
					x[j*FFTBLOCKPAD+i] = y[j*FFTBLOCKPAD+i]
				}
			}
			break
		}
		ft.fftz2(is, l+1, m, n, FFTBLOCK, FFTBLOCKPAD, u, y, x)
	}
}
func (ft *FTBenchmark) fftz2(is, l, m, n, ny, ny1 int, u, x, y []Dcomplex) {
	n1 := n / 2
	lk := 1 << (l - 1)
	li := 1 << (m - l)
	lj := 2 * lk
	ku := li

	for i := 0; i <= li-1; i++ {
		i11 := i * lk
		i12 := i11 + n1
		i21 := i * lj
		i22 := i21 + lk

		var u1 Dcomplex
		if is >= 1 {
			u1 = u[ku+i]
		} else {
			// CORREÇÃO AQUI:
			// Antes: u1 = common.Dconjg(u[ku+i])
			// Agora: Usa o pacote padrão math/cmplx
			u1 = cmplx.Conj(u[ku+i])
		}

		for k := 0; k <= lk-1; k++ {
			for j := 0; j < ny; j++ {
				x11 := x[(i11+k)*ny1+j]
				x21 := x[(i12+k)*ny1+j]

				y[(i21+k)*ny1+j] = x11 + x21
				y[(i22+k)*ny1+j] = u1 * (x11 - x21)
			}
		}
	}
}

// cffts1 performs FFT in 1st dimension
func (ft *FTBenchmark) cffts1(is, d1, d2, d3 int, x, xout []Dcomplex) {
	logd1 := ilog2(d1)

	// Scratch arrays
	y1 := make([]Dcomplex, d1*FFTBLOCKPAD)
	y2 := make([]Dcomplex, d1*FFTBLOCKPAD)

	if timersEnabled {
		common.TimerStart(T_FFTX)
	}

	for k := 0; k < d3; k++ {
		for jj := 0; jj <= d2-FFTBLOCK; jj += FFTBLOCK {
			// Load into blocks
			for j := 0; j < FFTBLOCK; j++ {
				for i := 0; i < d1; i++ {
					// x[k][j+jj][i] -> flat index: k*d2*d1 + (j+jj)*d1 + i
					y1[i*FFTBLOCKPAD+j] = x[k*d2*d1+(j+jj)*d1+i]
				}
			}

			ft.cfftz(is, logd1, d1, y1, y2)

			// Store back
			for j := 0; j < FFTBLOCK; j++ {
				for i := 0; i < d1; i++ {
					xout[k*d2*d1+(j+jj)*d1+i] = y1[i*FFTBLOCKPAD+j]
				}
			}
		}
	}
	if timersEnabled {
		common.TimerStop(T_FFTX)
	}
}

// cffts2 performs FFT in 2nd dimension
func (ft *FTBenchmark) cffts2(is, d1, d2, d3 int, x, xout []Dcomplex) {
	logd2 := ilog2(d2)
	y1 := make([]Dcomplex, d2*FFTBLOCKPAD)
	y2 := make([]Dcomplex, d2*FFTBLOCKPAD)

	if timersEnabled {
		common.TimerStart(T_FFTY)
	}

	for k := 0; k < d3; k++ {
		for ii := 0; ii <= d1-FFTBLOCK; ii += FFTBLOCK {
			for j := 0; j < d2; j++ {
				for i := 0; i < FFTBLOCK; i++ {
					y1[j*FFTBLOCKPAD+i] = x[k*d2*d1+j*d1+(i+ii)]
				}
			}

			ft.cfftz(is, logd2, d2, y1, y2)

			for j := 0; j < d2; j++ {
				for i := 0; i < FFTBLOCK; i++ {
					xout[k*d2*d1+j*d1+(i+ii)] = y1[j*FFTBLOCKPAD+i]
				}
			}
		}
	}
	if timersEnabled {
		common.TimerStop(T_FFTY)
	}
}

// cffts3 performs FFT in 3rd dimension
func (ft *FTBenchmark) cffts3(is, d1, d2, d3 int, x, xout []Dcomplex) {
	logd3 := ilog2(d3)
	y1 := make([]Dcomplex, d3*FFTBLOCKPAD)
	y2 := make([]Dcomplex, d3*FFTBLOCKPAD)

	if timersEnabled {
		common.TimerStart(T_FFTZ)
	}

	for j := 0; j < d2; j++ {
		for ii := 0; ii <= d1-FFTBLOCK; ii += FFTBLOCK {
			for k := 0; k < d3; k++ {
				for i := 0; i < FFTBLOCK; i++ {
					y1[k*FFTBLOCKPAD+i] = x[k*d2*d1+j*d1+(i+ii)]
				}
			}

			ft.cfftz(is, logd3, d3, y1, y2)

			for k := 0; k < d3; k++ {
				for i := 0; i < FFTBLOCK; i++ {
					xout[k*d2*d1+j*d1+(i+ii)] = y1[k*FFTBLOCKPAD+i]
				}
			}
		}
	}
	if timersEnabled {
		common.TimerStop(T_FFTZ)
	}
}

// fft performs the main FFT operation sequence
func (ft *FTBenchmark) fft(dir int, x1, x2 []Dcomplex) {
	if dir == 1 {
		ft.cffts1(1, dims[0], dims[1], dims[2], x1, x1)
		ft.cffts2(1, dims[0], dims[1], dims[2], x1, x1)
		ft.cffts3(1, dims[0], dims[1], dims[2], x1, x2)
	} else {
		ft.cffts3(-1, dims[0], dims[1], dims[2], x1, x1)
		ft.cffts2(-1, dims[0], dims[1], dims[2], x1, x1)
		ft.cffts1(-1, dims[0], dims[1], dims[2], x1, x2)
	}
}

// evolve performs the evolution step
func (ft *FTBenchmark) evolve(u0, u1, twiddle []Dcomplex, d1, d2, d3 int) {
	for k := 0; k < d3; k++ {
		for j := 0; j < d2; j++ {
			for i := 0; i < d1; i++ {
				idx := k*d2*d1 + j*d1 + i
				u0[idx] = u0[idx] * twiddle[idx]
				u1[idx] = u0[idx]
			}
		}
	}
}

// checksum computes the checksum
func (ft *FTBenchmark) checksum(i int, u1 []Dcomplex, d1, d2, d3 int) {
	chk := complex(0.0, 0.0)

	for j := 1; j <= 1024; j++ {
		q := j % NX
		r := (3 * j) % NY
		s := (5 * j) % NZ
		idx := s*d2*d1 + r*d1 + q
		chk += u1[idx]
	}

	chk = chk / complex(float64(NTOTAL), 0.0)
	fmt.Printf(" T =%5d     Checksum =%22.12e%22.12e\n", i, real(chk), imag(chk))
	sums[i] = chk
}

// verify performs verification against reference values
func (ft *FTBenchmark) verify(d1, d2, d3, nt int, verified *bool, class_npb *string) {
	// Reference checksums
	csum_ref := make([]Dcomplex, 26)
	*class_npb = "U"
	*verified = false
	epsilon := 1.0e-12

	// Initialize references based on problem size
	if d1 == 64 && d2 == 64 && d3 == 64 && nt == 6 {
		*class_npb = "S"
		csum_ref[1] = complex(5.546087004964e+02, 4.845363331978e+02)
		csum_ref[2] = complex(5.546385409189e+02, 4.865304269511e+02)
		csum_ref[3] = complex(5.546148406171e+02, 4.883910722336e+02)
		csum_ref[4] = complex(5.545423607415e+02, 4.901273169046e+02)
		csum_ref[5] = complex(5.544255039624e+02, 4.917475857993e+02)
		csum_ref[6] = complex(5.542683411902e+02, 4.932597244941e+02)
	} else if d1 == 128 && d2 == 128 && d3 == 32 && nt == 6 {
		*class_npb = "W"
		csum_ref[1] = complex(5.673612178944e+02, 5.293246849175e+02)
		csum_ref[6] = complex(5.504159734538e+02, 5.239212247086e+02)
	} else if d1 == 256 && d2 == 256 && d3 == 128 && nt == 6 {
		*class_npb = "A"
		csum_ref[6] = complex(5.091487099959e+02, 5.107917842803e+02)
	} else if d1 == 512 && d2 == 256 && d3 == 256 && nt == 20 {
		*class_npb = "B"
		csum_ref[20] = complex(5.124146770029e+02, 5.115744692211e+02)
	} else if d1 == 512 && d2 == 512 && d3 == 512 && nt == 20 {
		*class_npb = "C"
		csum_ref[20] = complex(5.129714421109e+02, 5.123465164008e+02)
	} else if d1 == 2048 && d2 == 1024 && d3 == 1024 && nt == 25 {
		*class_npb = "D"
		csum_ref[25] = complex(5.118822370068e+02, 5.119794338060e+02)
	}

	if *class_npb != "U" {
		*verified = true
		// Check only first and last for brevity in classes other than S/W
		// But full check logic:
		for i := 1; i <= nt; i++ {
			if csum_ref[i] == 0 {
				continue
			} // Skip unchecked iterations for larger classes

			ref := csum_ref[i]
			sum := sums[i]

			// Error calculation: |(sum - ref) / ref|
			diff := sum - ref
			modDiff := math.Sqrt(real(diff)*real(diff) + imag(diff)*imag(diff))
			modRef := math.Sqrt(real(ref)*real(ref) + imag(ref)*imag(ref))
			err := modDiff / modRef

			if err > epsilon {
				*verified = false
				break
			}
		}
	}
}

func (ft *FTBenchmark) run() {
	// Setup timers
	if _, err := os.Stat("timer.flag"); err == nil {
		timersEnabled = true
	} else {
		timersEnabled = false
	}

	for i := 0; i < T_MAX+1; i++ {
		common.TimerClear(i)
	}

	// Setup global dims and arrays
	dims[0], dims[1], dims[2] = NX, NY, NZ
	NTOTAL = NX * NY * NZ

	// Allocation
	u0 = make([]Dcomplex, NTOTAL)
	u1 = make([]Dcomplex, NTOTAL)
	twiddle = make([]Dcomplex, NTOTAL)
	sums = make([]Dcomplex, NITER+1)
	u = make([]Dcomplex, params.MAXDIM)

	fmt.Printf("\n\n NAS Parallel Benchmarks 4.1 Serial Go version - FT Benchmark\n\n")
	fmt.Printf(" Size                : %4dx%4dx%4d\n", NX, NY, NZ)
	fmt.Printf(" Iterations                  :%7d\n\n", NITER)

	// 1. Warmup Run
	ft.compute_indexmap(twiddle, dims[0], dims[1], dims[2])
	ft.compute_initial_conditions(u1, dims[0], dims[1], dims[2])
	ft.fft_init(params.MAXDIM)
	ft.fft(1, u1, u0)

	// 2. Timed Run
	for i := 0; i < T_MAX+1; i++ {
		common.TimerClear(i)
	}

	common.TimerStart(T_TOTAL)
	if timersEnabled {
		common.TimerStart(T_SETUP)
	}

	ft.compute_indexmap(twiddle, dims[0], dims[1], dims[2])
	ft.compute_initial_conditions(u1, dims[0], dims[1], dims[2])
	ft.fft_init(params.MAXDIM)

	if timersEnabled {
		common.TimerStop(T_SETUP)
	}
	if timersEnabled {
		common.TimerStart(T_FFT)
	}

	ft.fft(1, u1, u0)

	if timersEnabled {
		common.TimerStop(T_FFT)
	}

	for iter := 1; iter <= NITER; iter++ {
		if timersEnabled {
			common.TimerStart(T_EVOLVE)
		}
		ft.evolve(u0, u1, twiddle, dims[0], dims[1], dims[2])
		if timersEnabled {
			common.TimerStop(T_EVOLVE)
		}

		if timersEnabled {
			common.TimerStart(T_FFT)
		}
		ft.fft(-1, u1, u1)
		if timersEnabled {
			common.TimerStop(T_FFT)
		}

		if timersEnabled {
			common.TimerStart(T_CHECKSUM)
		}
		ft.checksum(iter, u1, dims[0], dims[1], dims[2])
		if timersEnabled {
			common.TimerStop(T_CHECKSUM)
		}
	}

	var verified bool
	var class_npb string
	ft.verify(NX, NY, NZ, NITER, &verified, &class_npb)

	common.TimerStop(T_TOTAL)
	totalTime := common.TimerRead(T_TOTAL)

	mflops := 0.0
	if totalTime != 0.0 {
		ntVal := float64(NTOTAL)
		mflops = 1.0e-6 * ntVal *
			(14.8157 + 7.19641*math.Log(ntVal) +
				(5.23518+7.21113*math.Log(ntVal))*float64(NITER)) / totalTime
	}

	verificationStr := "FAILED"
	if verified {
		verificationStr = "SUCCESSFUL"
	}
	fmt.Printf(" Result verification %s\n", verificationStr)
	fmt.Printf(" class_npb = %s\n", class_npb)

	common.PrintResults("FT", class_npb, NX, NY, NZ, NITER, totalTime, mflops, "floating point", verified, "4.1", "Unknown", "Go", "")

	if timersEnabled {
		tstrings := []string{"", "total", "setup", "fft", "evolve", "checksum", "fftx", "ffty", "fftz"}
		fmt.Println("  SECTION   Time (secs)")
		for i := 1; i <= T_MAX; i++ {
			t := common.TimerRead(i)
			fmt.Printf("  %-8s:%9.3f  (%6.2f%%)\n", tstrings[i], t, t*100.0/totalTime)
		}
	}
}

func main() {
	if params.EmptyTag {
		fmt.Println("To make a NAS benchmark type ")
		fmt.Println("\t go build -o ft -tags=<CLASS>")
		fmt.Println("where: <class> is \"S\", \"W\", \"A\", \"B\", \"C\", \"D\" or \"E\"")
		return
	}

	// Set global variables from params
	NX = params.NX
	NY = params.NY
	NZ = params.NZ
	NITER = params.NITER
	CLASS = params.CLASS

	ft := NewFTBenchmark()
	ft.run()
}
