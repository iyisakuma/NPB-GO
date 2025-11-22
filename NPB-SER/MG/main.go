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
	mg.nx = params.NX
	mg.ny = params.NY
	mg.nz = params.NZ
	mg.nit = params.NIT
	mg.lm = params.LM
	mg.class = params.CLASS

	// Run benchmark
	mg.run()
}
