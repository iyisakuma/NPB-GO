package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iyisakuma/NPB-GO/NPB-GOUROUTINE/EP/params"
	"github.com/iyisakuma/NPB-GO/NPB-GOUROUTINE/common"
)

const (
	MK      = 16
	MM      = params.M - MK
	NN      = 1 << MM
	NK      = 1 << MK
	NQ      = 10
	EPSILON = 1.0e-8
	A       = 1220703125.0
	S       = 271828183.0
	NK_PLUS = (2*NK + 1)
)

type WorkerResults struct {
	QPartial  []float64
	SXPartial float64
	SYPartial float64
}

func epWorker(
	startK int,
	endK int,
	an float64,
	resultsChan chan<- WorkerResults,
	wg *sync.WaitGroup,
	timersEnabled bool,
	goID int,
) {
	defer wg.Done()

	var t1, t2, t3, t4, x1, x2 float64
	var sx, sy float64
	var kk, i, ik, l int
	kOffset := -1

	qq := make([]float64, NQ)
	x := make([]float64, NK_PLUS)

	for i = 0; i < NQ; i++ {
		qq[i] = 0.0
	}
	for k := startK; k < endK; k++ {
		kk = kOffset + k

		t1 = S
		t2 = an

		/* find starting seed t1 for this kk */
		for i = 1; i <= 100; i++ {
			ik = kk / 2
			if (2 * ik) != kk {
				t3 = common.Randlc(&t1, t2)
			}
			if ik == 0 {
				break
			}
			t3 = common.Randlc(&t2, t2)
			kk = ik
		}

		/* compute uniform pseudorandom numbers */
		if timersEnabled && goID == 0 {
			common.TimerStart(2)
		}
		common.Vranlc(2*NK, &t1, A, x)
		if timersEnabled && goID == 0 {
			common.TimerStop(2)
		}

		/*
		 * compute gaussian deviates by acceptance-rejection method and
		 * tally counts in concentric square annuli. this loop is not
		 * vectorizable.
		 */
		if timersEnabled && goID == 0 {
			common.TimerStart(1)
		}
		for i = 0; i < NK; i++ {
			x1 = 2.0*x[2*i] - 1.0
			x2 = 2.0*x[2*i+1] - 1.0
			t1 = x1*x1 + x2*x2
			if t1 <= 1.0 {
				t2 = math.Sqrt(-2.0 * math.Log(t1) / t1)
				t3 = (x1 * t2)
				t4 = (x2 * t2)
				l = int(math.Max(math.Abs(t3), math.Abs(t4)))
				qq[l] += 1.0
				sx += t3
				sy += t4
			}
		}
		if timersEnabled && goID == 0 {
			common.TimerStop(1)
		}
	}

	resultsChan <- WorkerResults{
		QPartial:  qq,
		SXPartial: sx,
		SYPartial: sy,
	}
}

var x = make([]float64, NK_PLUS)
var q = make([]float64, NQ)

func EpParallel() {

	if params.EmptyTag {
		fmt.Println("To make a NAS benchmark type ")
		fmt.Println("\t go build -o ep -tags=<CLASS>")
		fmt.Println("where: <class> is \"S\", \"W\", \"A\", \"B\", \"C\", \"D\" or \"E\"")
		return
	}

	var Mops, t1 float64
	var sx, sy, tm, an, tt, gc float64
	var sxErr, syErr float64
	var np int
	var i, nit int
	var verified, timersEnabled bool
	var dum = []float64{1.0, 1.0, 1.0}
	var size string

	timersEnabled = checkTimeFlag()
	size = fmt.Sprintf("%15.0f", math.Pow(2.0, params.M+1))

	size = strings.TrimRight(size, ".")

	fmt.Println("\n\n NAS Parallel Benchmarks 4.1 Serial Go version - EP Benchmark\n")
	fmt.Printf(" Number of random numbers generated: %15s\n", size)
	verified = false

	/*
	 * --------------------------------------------------------------------
	 * compute the number of "batches" of random number pairs generated
	 * per processor. Adjust if the number of processors does not evenly
	 * divide the total number
	 * --------------------------------------------------------------------
	 */
	np = NN

	/*
	 * call the random number generator functions and initialize
	 * the x-array to reduce the effects of paging on the timings.
	 * also, call all mathematical functions that are used. make
	 * sure these initializations cannot be eliminated as dead code.
	 */
	common.Vranlc(0, &dum[0], dum[1], dum)
	dum[0] = common.Randlc(&dum[1], dum[2])

	for i = 0; i < NK_PLUS; i++ {
		x[i] = -1.0e99
	}

	Mops = math.Log(math.Sqrt(math.Abs(math.Max(1.0, 1.0))))

	common.TimerClear(0)
	common.TimerClear(1)
	common.TimerClear(2)
	common.TimerStart(0)

	t1 = A
	common.Vranlc(0, &t1, A, x)

	for i = 0; i < MK+1; i++ {
		common.Randlc(&t1, t1)
	}

	an = t1
	tt = S
	gc = 0.0
	sx = 0.0
	sy = 0.0

	for i = 0; i <= NQ-1; i++ {
		q[i] = 0.0
	}

	numCPUs := runtime.NumCPU()
	if nw := os.Getenv("GO_NUM_THREADS"); nw != "" {
		if n, err := strconv.Atoi(nw); err == nil && n > 0 {
			numCPUs = n
		}
	}
	runtime.GOMAXPROCS(numCPUs)
	var wg sync.WaitGroup
	partialResultsChan := make(chan WorkerResults, numCPUs)
	chunks := np / numCPUs

	if np%numCPUs != 0 {
		chunks++
	}

	wg.Add(numCPUs)
	for i := 0; i < numCPUs; i++ {
		startK := i*chunks + 1
		endK := startK + chunks
		if endK > np+1 {
			endK = np + 1
		}

		if startK >= endK {
			continue
		}

		go epWorker(startK, endK, an, partialResultsChan, &wg, timersEnabled, i)
	}

	wg.Wait()
	close(partialResultsChan)

	for workerRes := range partialResultsChan {
		// Agrega q (contagens de anéis)
		for i := 0; i < NQ; i++ {
			q[i] += workerRes.QPartial[i]
		}
		sx += workerRes.SXPartial
		sy += workerRes.SYPartial

	}

	for i = 0; i <= NQ-1; i++ {
		gc = gc + q[i]
	}
	common.TimerStop(0)
	tm = common.TimerRead(0)

	nit = 0

	sxErr = math.Abs((sx - params.SX_VERIFY_VALUE) / params.SX_VERIFY_VALUE)
	syErr = math.Abs((sy - params.SY_VERIFY_VALUE) / params.SY_VERIFY_VALUE)
	verified = (sxErr <= EPSILON) && (syErr <= EPSILON)

	Mops = math.Pow(2.0, params.M+1) / tm / 1000000.0

	fmt.Println("\n EP Benchmark Results:\n")
	fmt.Printf(" CPU Time =%10.4f\n", tm)
	fmt.Printf(" N = 2^%5d\n", params.M)
	fmt.Printf(" No. Gaussian Pairs = %15.0f\n", gc)
	fmt.Printf(" Sums = %25.15e %25.15e\n", sx, sy)
	fmt.Printf(" No. Goroutines: %5d\n", numCPUs)
	fmt.Printf(" Chunks Size: %5d\n", chunks)
	fmt.Println(" Counts:")

	for i = 0; i < NQ-1; i++ {
		fmt.Printf("%3d%15.0f\n", i, q[i])
	}

	common.PrintResults(
		"EP",
		params.CLASS,
		params.M+1,
		0,
		0,
		nit,
		tm,
		Mops,
		"Random numbers generated",
		verified,
		"4.1",
		time.Now().Format("03 Jun 2006"),
		"go1.24.2 linux/amd64",
		"randdp",
	)
	if timersEnabled {
		if tm <= 0.0 {
			tm = 1.0
		}
		tt = common.TimerRead(0)
		fmt.Printf("\nTotal time:     %9.3f (%6.2f)\n", tt, tt*100.0/tm)

		tt = common.TimerRead(1)
		fmt.Printf("Gaussian pairs: %9.3f (%6.2f)\n", tt, tt*100.0/tm)

		tt = common.TimerRead(2)
		fmt.Printf("Random numbers: %9.3f (%6.2f)\n", tt, tt*100.0/tm)
	}
}

func checkTimeFlag() bool {
	_, err := os.Stat("timer.flag")
	return os.IsNotExist(err)
}

func main() {
	EpParallel()
}
