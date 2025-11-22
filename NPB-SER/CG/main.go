package main

import "os"

func main() {
	// Set problem size based on command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "S":
			NA = 1400
			NITER = 15
			SHIFT = 10.0
			NONZER = 7
			NZ = NA * (NONZER + 1) * (NONZER + 1)
			zetaVerifyValue = 8.5971775078648
			classNPB = "S"
		case "W":
			NA = 7000
			NITER = 15
			SHIFT = 12.0
			NONZER = 8
			NZ = NA * (NONZER + 1) * (NONZER + 1)
			zetaVerifyValue = 10.362595087124
			classNPB = "W"
		case "A":
			NA = 14000
			NITER = 15
			SHIFT = 20.0
			NONZER = 11
			NZ = NA * (NONZER + 1) * (NONZER + 1)
			zetaVerifyValue = 17.130235054029
			classNPB = "A"
		case "B":
			NA = 75000
			NITER = 75
			SHIFT = 60.0
			NONZER = 13
			NZ = NA * (NONZER + 1) * (NONZER + 1)
			zetaVerifyValue = 22.712745482631
			classNPB = "B"
		case "C":
			NA = 150000
			NITER = 75
			SHIFT = 110.0
			NONZER = 15
			NZ = NA * (NONZER + 1) * (NONZER + 1)
			zetaVerifyValue = 28.973605592845
			classNPB = "C"
		case "D":
			NA = 1500000
			NITER = 100
			SHIFT = 500.0
			NONZER = 21
			NZ = NA * (NONZER + 1) * (NONZER + 1)
			zetaVerifyValue = 52.514532105794
			classNPB = "D"
		case "E":
			NA = 9000000
			NITER = 100
			SHIFT = 1500.0
			NONZER = 26
			NZ = NA * (NONZER + 1) * (NONZER + 1)
			zetaVerifyValue = 77.522164599383
			classNPB = "E"
		default:
			NA = 1400
			NITER = 15
			SHIFT = 10.0
			NONZER = 7
			NZ = NA * (NONZER + 1) * (NONZER + 1)
			zetaVerifyValue = 8.5971775078648
			classNPB = "S"
		}
	} else {
		// Default to class S
		NA = 1400
		NITER = 15
		SHIFT = 10.0
		NONZER = 7
		NZ = NA * (NONZER + 1) * (NONZER + 1)
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
