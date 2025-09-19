package common

import (
	"fmt"
	"math"
)

func PrintResults(name, classNPB string, n1, n2, n3, niter int, t, mops float64, optype string, passedVerification bool, npbversion, compiletime, compilerversion, rand string) {
	fmt.Printf("\n\n %s Benchmark Completed\n", name)
	fmt.Printf(" class_npb       =                        %s\n", classNPB)

	if len(name) >= 2 && name[:2] == "IS" {
		if n3 == 0 {
			nn := int64(n1)
			if n2 != 0 {
				nn *= int64(n2)
			}
			fmt.Printf(" Size            =             %12d\n", nn)
		} else {
			fmt.Printf(" Size            =             %4dx%4dx%4d\n", n1, n2, n3)
		}
	} else {
		if n2 == 0 && n3 == 0 {
			if len(name) >= 2 && name[:2] == "EP" {
				size := fmt.Sprintf("%15.0f", math.Pow(2.0, float64(n1)))
				// remove ponto final se presente
				if size[len(size)-1] == '.' {
					size = size[:len(size)-1]
				}
				fmt.Printf(" Size            =          %15s\n", size)
			} else {
				fmt.Printf(" Size            =             %12d\n", n1)
			}
		} else {
			fmt.Printf(" Size            =           %4dx%4dx%4d\n", n1, n2, n3)
		}
	}

	fmt.Printf(" Iterations      =             %12d\n", niter)
	fmt.Printf(" Time in seconds =             %12.2f\n", t)
	fmt.Printf(" Mop/s total     =             %12.2f\n", mops)
	fmt.Printf(" Operation type  = %24s\n", optype)

	if passedVerification {
		fmt.Println(" Verification    =               SUCCESSFUL")
	} else {
		fmt.Println(" Verification    =            NOT PERFORMED")
	}

	fmt.Printf(" Version         =             %12s\n", npbversion)
	fmt.Printf(" Compiler ver    =             %12s\n", compilerversion)
	fmt.Printf(" Compile date    =             %12s\n", compiletime)

	fmt.Println("\n Compile options:")
	fmt.Printf("    RAND         = %s\n", rand)
	fmt.Println("\n\n----------------------------------------------------------------------")
	fmt.Println("    NPB-GO is developed by: ")
	fmt.Println("        Igor Yuji Ishihara Sakuma")
	fmt.Println()
	fmt.Println("----------------------------------------------------------------------")
	fmt.Println()
}
