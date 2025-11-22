package main

import "os"

func main() {
	// Set problem size based on command line arguments
	var nx, ny, nz, nit, lm int
	var class string

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "S":
			nx, ny, nz = 32, 32, 32
			nit = 4
			lm = 5
			class = "S"
		case "W":
			nx, ny, nz = 64, 64, 64
			nit = 4
			lm = 5
			class = "W"
		case "A":
			nx, ny, nz = 256, 256, 256
			nit = 4
			lm = 5
			class = "A"
		case "B":
			nx, ny, nz = 256, 256, 256
			nit = 20
			lm = 5
			class = "B"
		case "C":
			nx, ny, nz = 512, 512, 512
			nit = 20
			lm = 5
			class = "C"
		case "D":
			nx, ny, nz = 1024, 1024, 1024
			nit = 50
			lm = 5
			class = "D"
		case "E":
			nx, ny, nz = 2048, 2048, 2048
			nit = 50
			lm = 5
			class = "E"
		default:
			nx, ny, nz = 32, 32, 32
			nit = 4
			lm = 5
			class = "S"
		}
	} else {
		// Default to class S
		nx, ny, nz = 32, 32, 32
		nit = 4
		lm = 5
		class = "S"
	}

	// Create benchmark instance
	mg := NewMGBenchmark()
	mg.nx = nx
	mg.ny = ny
	mg.nz = nz
	mg.nit = nit
	mg.lm = lm
	mg.class = class

	// Run benchmark
	mg.run()
}
