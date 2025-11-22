package main

import (
	"fmt"

	"github.com/iyisakuma/NPB-GO/NPB-SER/CG/params"
)

func main() {
	if params.EmptyTag {
		fmt.Println("To make a NAS benchmark type ")
		fmt.Println("\t go build -o cg -tags=<CLASS>")
		fmt.Println("where: <class> is \"S\", \"W\", \"A\", \"B\", \"C\", \"D\" or \"E\"")
		return
	}

	// Set global variables from params
	NA = params.NA
	NITER = params.NITER
	SHIFT = params.SHIFT
	NONZER = params.NONZER
	NZ = NA * (NONZER + 1) * (NONZER + 1)
	zetaVerifyValue = params.ZETA_VERIFY_VALUE
	classNPB = params.CLASS

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
