package main

import (
	"fmt"
	"github.com/iyisakuma/NPB-GO/NPB-GOUROUTINE/MG/params"
)

func main() {
	if params.EmptyTag {
		fmt.Println("To make a NAS benchmark type ")
		fmt.Println("\t go build -o mg -tags=<CLASS>")
		fmt.Println("where: <class> is \"S\", \"W\", \"A\", \"B\", \"C\", \"D\" or \"E\"")
		return
	}

	// Create benchmark instance
	mg := NewMGBenchmark()
	mg.nit = params.NIT
	mg.class = params.CLASS
	mg.debug_vec[0] = 0 // Ativa os prints de rep_nrm

	// Calculate LM and LT_DEFAULT based on problem size
	// LM is log2 of NX (assuming NX = NY = NZ for MG benchmark)
	lm := 0
	for n := params.NX; n > 1; n >>= 1 {
		lm++
	}

	// Set lt and lt_default
	mg.lt = lm
	mg.lt_default = lm

	// Initialize arrays with correct size
	maxlevel := lm + 1
	mg.nx = make([]int, maxlevel+1)
	mg.ny = make([]int, maxlevel+1)
	mg.nz = make([]int, maxlevel+1)
	mg.m1 = make([]int, maxlevel+1)
	mg.m2 = make([]int, maxlevel+1)
	mg.m3 = make([]int, maxlevel+1)
	mg.ir = make([]int, maxlevel+1)

	// Store initial values at top level (will be set properly in setup())
	mg.nx[lm] = params.NX
	mg.ny[lm] = params.NY
	mg.nz[lm] = params.NZ

	// Run benchmark
	mg.run()
}
