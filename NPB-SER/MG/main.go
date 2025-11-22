package main

import (
	"fmt"

	"github.com/iyisakuma/NPB-GO/NPB-SER/MG/params"
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
	// Initialize nx, ny, nz arrays (will be set properly in setup)
	mg.nx = make([]int, MAXLEVEL+1)
	mg.ny = make([]int, MAXLEVEL+1)
	mg.nz = make([]int, MAXLEVEL+1)
	mg.nit = params.NIT
	mg.class = params.CLASS
	// Store initial values at top level (lt = LT_DEFAULT = 5)
	mg.nx[mg.lt] = params.NX
	mg.ny[mg.lt] = params.NY
	mg.nz[mg.lt] = params.NZ
	mg.debug_vec[0] = 0 // Ativa os prints de rep_nrm
	// Run benchmark
	mg.run()
}
