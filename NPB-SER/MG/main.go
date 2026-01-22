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
	mg := NewMGBenchmark()
	mg.nit = params.NIT
	mg.class = params.CLASS

	lm := 0
	for n := params.NX; n > 1; n >>= 1 {
		lm++
	}

	mg.lt = lm
	mg.lt_default = lm

	maxlevel := lm + 1
	mg.nx = make([]int, maxlevel+1)
	mg.ny = make([]int, maxlevel+1)
	mg.nz = make([]int, maxlevel+1)
	mg.m1 = make([]int, maxlevel+1)
	mg.m2 = make([]int, maxlevel+1)
	mg.m3 = make([]int, maxlevel+1)
	mg.ir = make([]int, maxlevel+1)

	mg.nx[lm] = params.NX
	mg.ny[lm] = params.NY
	mg.nz[lm] = params.NZ
	mg.run()
}
