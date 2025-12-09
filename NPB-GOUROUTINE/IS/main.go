package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/iyisakuma/NPB-GO/NPB-GOUROUTINE/IS/params"
	"github.com/iyisakuma/NPB-GO/NPB-GOUROUTINE/IS/types"
	"github.com/iyisakuma/NPB-GO/NPB-GOUROUTINE/common"
)

const (
	TOTAL_KEYS        = 1 << params.TOTAL_KEYS_LOG_2
	MAX_KEY           = 1 << params.MAX_KEY_LOG_2
	NUM_BUCKETS       = 1 << params.NUM_BUCKETS_LOG_2
	NUM_KEYS          = TOTAL_KEYS
	SIZE_OF_BUFFERS   = NUM_KEYS
	T_BENCHMARKING    = 0
	T_INITIALIZATION  = 1
	T_SORTING         = 2
	T_TOTAL_EXECUTION = 3
	MAX_ITERATIONS    = 10
	USE_BUCKETS       = true
	TEST_ARRAY_SIZE   = 5
)

// ISBenchmark represents the IS (Integer Sort) benchmark
// This struct encapsulates all the state that was global in the C++ version
type ISBenchmark struct {
	// Main arrays (equivalent to global arrays in C++)
	keyArray          []types.INT_TYPE
	keyBuff1          []types.INT_TYPE
	keyBuff2          []types.INT_TYPE
	partialVerifyVals []types.INT_TYPE

	// For USE_BUCKETS mode
	bucketSize [][]types.INT_TYPE // [numProcs][NUM_BUCKETS]
	bucketPtrs []types.INT_TYPE   // [NUM_BUCKETS]

	// For !USE_BUCKETS mode
	keyBuff1Aptr [][]types.INT_TYPE // [numProcs][MAX_KEY]

	// Global state (equivalent to global variables in C++)
	keyBuffPtrGlobal   []types.INT_TYPE // Points to keyBuff1 (like pointer in C++)
	passedVerification int

	numProcs          int
	verificationMutex sync.Mutex
}

// NewISBenchmark creates a new IS benchmark instance
func NewISBenchmark() *ISBenchmark {
	numProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(numProcs)

	bench := &ISBenchmark{
		keyArray:          make([]types.INT_TYPE, SIZE_OF_BUFFERS),
		keyBuff1:          make([]types.INT_TYPE, MAX_KEY),
		keyBuff2:          make([]types.INT_TYPE, SIZE_OF_BUFFERS),
		partialVerifyVals: make([]types.INT_TYPE, TEST_ARRAY_SIZE),
		keyBuffPtrGlobal:  make([]types.INT_TYPE, MAX_KEY),
		numProcs:          numProcs,
	}

	// Initialize bucketPtrs for USE_BUCKETS mode
	if USE_BUCKETS {
		bench.bucketPtrs = make([]types.INT_TYPE, NUM_BUCKETS)
	}

	return bench
}

func main() {
	if params.EmptyTag {
		fmt.Println("To make a NAS benchmark type ")
		fmt.Println("\t go build -o is -tags=<CLASS>")
		fmt.Println("where: <class> is \"S\", \"W\", \"A\", \"B\", \"C\" or \"D\"")
		return
	}

	bench := NewISBenchmark()
	bench.run()
}

func (b *ISBenchmark) run() {
	var timerOn bool
	var timecounter float64

	// Initialize timers
	timerOn = false
	if _, err := os.Stat("timer.flag"); err == nil {
		timerOn = true
	}

	common.TimerClear(T_BENCHMARKING)
	if timerOn {
		common.TimerClear(T_INITIALIZATION)
		common.TimerClear(T_SORTING)
		common.TimerClear(T_TOTAL_EXECUTION)
	}

	if timerOn {
		common.TimerStart(T_TOTAL_EXECUTION)
	}

	// Printout initial NPB info
	fmt.Printf("\n\n NAS Parallel Benchmarks 4.1 Serial Go version - IS Benchmark\n\n")
	fmt.Printf(" Size:  %d  (class %s)\n", TOTAL_KEYS, params.CLASS)
	fmt.Printf(" Iterations:   %d\n", MAX_ITERATIONS)
	fmt.Printf("\n")

	if timerOn {
		common.TimerStart(T_INITIALIZATION)
	}

	// Generate random number sequence and subsequent keys
	b.createSeq(314159265.00, 1220703125.00)

	b.allocKeyBuff()

	if timerOn {
		common.TimerStop(T_INITIALIZATION)
	}

	// Do one iteration for free (i.e., untimed) to guarantee initialization
	b.rank(1)

	// Start verification counter
	b.passedVerification = 0

	if params.CLASS != "S" {
		fmt.Println("\n   iteration")
	}

	// Start timer
	common.TimerStart(T_BENCHMARKING)

	// This is the main iteration
	for iteration := types.INT_TYPE(1); iteration <= MAX_ITERATIONS; iteration++ {
		if params.CLASS != "S" {
			fmt.Printf("        %d\n", iteration)
		}
		b.rank(iteration)
	}

	// End of timing
	common.TimerStop(T_BENCHMARKING)
	timecounter = common.TimerRead(T_BENCHMARKING)

	// This tests that keys are in sequence: sorting of last ranked key seq
	if timerOn {
		common.TimerStart(T_SORTING)
	}
	b.fullVerify()
	if timerOn {
		common.TimerStop(T_SORTING)
		common.TimerStop(T_TOTAL_EXECUTION)
	}

	// The final printout
	if b.passedVerification != 5*MAX_ITERATIONS+1 {
		b.passedVerification = 0
	}

	// Print results (simplified version)
	fmt.Printf("\n")
	fmt.Printf(" IS Benchmark Completed\n")
	fmt.Printf(" class_npb       =                        %s\n", params.CLASS)
	fmt.Printf(" Size            =                        %d\n", TOTAL_KEYS)
	fmt.Printf(" Iterations      =                        %d\n", MAX_ITERATIONS)
	fmt.Printf(" Time in seconds =                     %.2f\n", timecounter)
	if timecounter > 0 {
		mops := float64(MAX_ITERATIONS*TOTAL_KEYS) / timecounter / 1000000.0
		fmt.Printf(" Mop/s total     =                    %.2f\n", mops)
	}
	fmt.Printf(" Operation type  =              keys ranked\n")
	if b.passedVerification > 0 {
		fmt.Printf(" Verification    =               SUCCESSFUL\n")
	} else {
		fmt.Printf(" Verification    =             UNSUCCESSFUL\n")
	}
	fmt.Printf(" Version         =                         \n")
	fmt.Printf(" Compiler ver    =                         \n")
	fmt.Printf(" Compile date    =                         \n")
	fmt.Printf("\n")
	fmt.Printf(" Compile options:\n")
	fmt.Printf("    RAND         = \n")
	fmt.Printf("\n")
	fmt.Printf("----------------------------------------------------------------------\n")
	fmt.Printf("    NPB-GO is developed by: \n")
	fmt.Printf("        Igor Yuji Ishihara Sakuma\n")
	fmt.Printf("\n")
	fmt.Printf("----------------------------------------------------------------------\n")
	fmt.Printf("\n")

	// Print additional timers
	if timerOn {
		tTotal := common.TimerRead(T_TOTAL_EXECUTION)
		fmt.Printf("\nAdditional timers -\n")
		fmt.Printf(" Total execution: %8.3f\n", tTotal)
		if tTotal == 0.0 {
			tTotal = 1.0
		}
		timecounter = common.TimerRead(T_INITIALIZATION)
		tPercent := timecounter / tTotal * 100.0
		fmt.Printf(" Initialization : %8.3f (%5.2f%%)\n", timecounter, tPercent)
		timecounter = common.TimerRead(T_BENCHMARKING)
		tPercent = timecounter / tTotal * 100.0
		fmt.Printf(" Benchmarking   : %8.3f (%5.2f%%)\n", timecounter, tPercent)
		timecounter = common.TimerRead(T_SORTING)
		tPercent = timecounter / tTotal * 100.0
		fmt.Printf(" Sorting        : %8.3f (%5.2f%%)\n", timecounter, tPercent)
	}
}

func (b *ISBenchmark) allocKeyBuff() {
	numProcs := b.numProcs

	if USE_BUCKETS {
		b.bucketSize = make([][]types.INT_TYPE, numProcs)
		for i := 0; i < numProcs; i++ {
			b.bucketSize[i] = make([]types.INT_TYPE, NUM_BUCKETS)
		}

		// Initialize keyBuff2 (parallel)
		var wg sync.WaitGroup
		chunk := (NUM_KEYS + numProcs - 1) / numProcs
		wg.Add(numProcs)
		for myid := 0; myid < numProcs; myid++ {
			go func(threadID int) {
				defer wg.Done()
				start := threadID * chunk
				end := start + chunk
				if end > NUM_KEYS {
					end = NUM_KEYS
				}
				for i := start; i < end; i++ {
					b.keyBuff2[i] = 0
				}
			}(myid)
		}
		wg.Wait()
	} else {
		b.keyBuff1Aptr = make([][]types.INT_TYPE, numProcs)
		b.keyBuff1Aptr[0] = b.keyBuff1
		for i := 1; i < numProcs; i++ {
			b.keyBuff1Aptr[i] = make([]types.INT_TYPE, MAX_KEY)
		}
	}
}

// findMySeed returns parallel random number seq seed
func findMySeed(kn int, np int, nn int64, s float64, a float64) float64 {
	if kn == 0 {
		return s
	}

	mq := (nn/4 + int64(np) - 1) / int64(np)
	nq := mq * 4 * int64(kn) // number of rans to be skipped

	t1 := s
	t2 := a
	kk := nq

	for kk > 1 {
		ik := kk / 2
		if 2*ik == kk {
			// CORREÇÃO: Ignorar o retorno, apenas atualizar t2 via ponteiro
			common.Randlc(&t2, t2)
			kk = ik
		} else {
			// CORREÇÃO: Ignorar o retorno, apenas atualizar t1 via ponteiro
			common.Randlc(&t1, t2)
			kk = kk - 1
		}
	}
	// CORREÇÃO: Ignorar o retorno
	common.Randlc(&t1, t2)

	return t1
}

// createSeq generates random number sequence and subsequent keys
func (b *ISBenchmark) createSeq(seed float64, a float64) {
	var wg sync.WaitGroup
	wg.Add(b.numProcs)

	// Parallel region - each goroutine processes a chunk of keys
	for myid := 0; myid < b.numProcs; myid++ {
		go func(threadID int) {
			defer wg.Done()
			var x, s float64
			var k types.INT_TYPE

			mq := (NUM_KEYS + b.numProcs - 1) / b.numProcs
			k1 := mq * threadID
			k2 := k1 + mq
			if k2 > NUM_KEYS {
				k2 = NUM_KEYS
			}

			s = findMySeed(threadID, b.numProcs, int64(4*NUM_KEYS), seed, a)
			k = MAX_KEY / 4

			for i := k1; i < k2; i++ {
				x = common.Randlc(&s, a)
				x += common.Randlc(&s, a)
				x += common.Randlc(&s, a)
				x += common.Randlc(&s, a)
				b.keyArray[i] = types.INT_TYPE(float64(k) * x)
			}
		}(myid)
	}
	wg.Wait()
}

// fullVerify verifies that all keys are correctly sorted
func (b *ISBenchmark) fullVerify() {
	if USE_BUCKETS {
		// Buckets are already sorted. Sorting keys within each bucket
		// Parallelize bucket processing (dynamic schedule equivalent)
		var wg sync.WaitGroup
		wg.Add(NUM_BUCKETS)

		for j := 0; j < NUM_BUCKETS; j++ {
			go func(bucketID int) {
				defer wg.Done()
				// Make all variables local to this goroutine
				var k, k1 types.INT_TYPE
				if bucketID > 0 {
					k1 = b.bucketPtrs[bucketID-1]
				} else {
					k1 = 0
				}
				// Read bucketPtrs once and store locally to avoid race conditions
				k2 := b.bucketPtrs[bucketID]

				for i := k1; i < k2; i++ {
					// Need to use mutex for keyBuffPtrGlobal decrement
					b.verificationMutex.Lock()
					k = b.keyBuffPtrGlobal[b.keyBuff2[i]] - 1
					b.keyBuffPtrGlobal[b.keyBuff2[i]] = k
					// Keep mutex locked during write to keyArray to prevent races
					b.keyArray[k] = b.keyBuff2[i]
					b.verificationMutex.Unlock()
				}
			}(j)
		}
		wg.Wait()
	} else {
		// Copy keyArray to keyBuff2
		for i := 0; i < NUM_KEYS; i++ {
			b.keyBuff2[i] = b.keyArray[i]
		}
		// This is actual sorting. Each thread is responsible for a subset of key values
		j := b.numProcs
		j = (MAX_KEY + j - 1) / j
		var wg sync.WaitGroup
		wg.Add(b.numProcs)

		for myid := 0; myid < b.numProcs; myid++ {
			go func(threadID int) {
				defer wg.Done()
				// Make k1, k2, k local to this goroutine
				var k, k1, k2 types.INT_TYPE
				k1 = types.INT_TYPE(j * threadID)
				k2 = k1 + types.INT_TYPE(j)
				if k2 > MAX_KEY {
					k2 = MAX_KEY
				}
				for i := 0; i < NUM_KEYS; i++ {
					if b.keyBuff2[i] >= k1 && b.keyBuff2[i] < k2 {
						b.verificationMutex.Lock()
						k = b.keyBuffPtrGlobal[b.keyBuff2[i]] - 1
						b.keyBuffPtrGlobal[b.keyBuff2[i]] = k
						// Keep mutex locked during write to keyArray
						b.keyArray[k] = b.keyBuff2[i]
						b.verificationMutex.Unlock()
					}
				}
			}(myid)
		}
		wg.Wait()
	}

	// Confirm keys correctly sorted: count incorrectly sorted keys, if any
	// Parallelize the verification loop
	jChan := make(chan int, b.numProcs)
	chunk := (NUM_KEYS - 1) / b.numProcs
	if chunk == 0 {
		chunk = 1
	}

	var wg sync.WaitGroup
	wg.Add(b.numProcs)
	for myid := 0; myid < b.numProcs; myid++ {
		go func(threadID int) {
			defer wg.Done()
			start := threadID*chunk + 1
			end := start + chunk
			if threadID == b.numProcs-1 {
				end = NUM_KEYS
			}
			localJ := 0
			for i := start; i < end; i++ {
				if b.keyArray[i-1] > b.keyArray[i] {
					localJ++
				}
			}
			jChan <- localJ
		}(myid)
	}
	wg.Wait()
	close(jChan)

	j := 0
	for count := range jChan {
		j += count
	}

	if j != 0 {
		fmt.Printf("Full_verify: number of keys out of sort: %d\n", j)
	} else {
		b.verificationMutex.Lock()
		b.passedVerification++
		b.verificationMutex.Unlock()
	}
}

// rank performs the main ranking/sorting operation for each iteration
// Equivalent to rank in C++ - PARALLELIZED
func (b *ISBenchmark) rank(iteration types.INT_TYPE) {
	var shift int
	var numBucketKeys types.INT_TYPE
	var keyBuffPtr, keyBuffPtr2 []types.INT_TYPE

	if USE_BUCKETS {
		shift = params.MAX_KEY_LOG_2 - params.NUM_BUCKETS_LOG_2
		numBucketKeys = types.INT_TYPE(1) << shift
	}

	// Set test values
	b.keyArray[iteration] = iteration
	b.keyArray[iteration+MAX_ITERATIONS] = MAX_KEY - iteration

	// Determine where the partial verify test keys are, load into partial_verify_vals
	for i := 0; i < TEST_ARRAY_SIZE; i++ {
		b.partialVerifyVals[i] = b.keyArray[params.TEST_INDEX_ARRAY[i]]
	}

	// Setup pointers to key buffers
	if USE_BUCKETS {
		keyBuffPtr2 = b.keyBuff2
	} else {
		keyBuffPtr2 = b.keyArray
	}
	keyBuffPtr = b.keyBuff1

	if USE_BUCKETS {
		// Parallel region for bucket processing
		var wg sync.WaitGroup
		wg.Add(b.numProcs)

		for myid := 0; myid < b.numProcs; myid++ {
			go func(threadID int) {
				defer wg.Done()
				workBuff := b.bucketSize[threadID]

				// Initialize
				for i := 0; i < NUM_BUCKETS; i++ {
					workBuff[i] = 0
				}

				// Determine the number of keys in each bucket (parallel loop)
				chunk := (NUM_KEYS + b.numProcs - 1) / b.numProcs
				start := threadID * chunk
				end := start + chunk
				if end > NUM_KEYS {
					end = NUM_KEYS
				}
				for i := start; i < end; i++ {
					workBuff[b.keyArray[i]>>shift]++
				}
			}(myid)
		}
		wg.Wait()

		// Each goroutine calculates its own bucket_ptrs (threadprivate equivalent)
		wg.Add(b.numProcs)
		for myid := 0; myid < b.numProcs; myid++ {
			go func(threadID int) {
				defer wg.Done()
				// Create local bucket_ptrs for this thread (threadprivate equivalent)
				localBucketPtrs := make([]types.INT_TYPE, NUM_BUCKETS)

				// Accumulative bucket sizes are the bucket pointers.
				// These are global sizes accumulated upon to each bucket
				localBucketPtrs[0] = 0
				for k := 0; k < threadID; k++ {
					localBucketPtrs[0] += b.bucketSize[k][0]
				}

				for i := 1; i < NUM_BUCKETS; i++ {
					localBucketPtrs[i] = localBucketPtrs[i-1]
					for k := 0; k < threadID; k++ {
						localBucketPtrs[i] += b.bucketSize[k][i]
					}
					for k := threadID; k < b.numProcs; k++ {
						localBucketPtrs[i] += b.bucketSize[k][i-1]
					}
				}

				// Sort into appropriate bucket - each thread processes its chunk
				chunk := (NUM_KEYS + b.numProcs - 1) / b.numProcs
				start := threadID * chunk
				end := start + chunk
				if end > NUM_KEYS {
					end = NUM_KEYS
				}
				for i := start; i < end; i++ {
					k := b.keyArray[i]
					bucketIdx := k >> shift
					pos := localBucketPtrs[bucketIdx]
					localBucketPtrs[bucketIdx]++
					b.keyBuff2[pos] = k
				}

				// The bucket pointers now point to the final accumulated sizes
				if threadID < b.numProcs-1 {
					for i := 0; i < NUM_BUCKETS; i++ {
						for k := threadID + 1; k < b.numProcs; k++ {
							localBucketPtrs[i] += b.bucketSize[k][i]
						}
					}
				}

				// Store final bucket pointers for this thread's contribution
				// We need to synchronize to update the global bucketPtrs
				// But actually, we only need the final values for the ranking phase
				// So we can store them temporarily or recalculate
			}(myid)
		}
		wg.Wait()

		// Recalculate global bucketPtrs for the ranking phase
		// (This is needed because each thread had its own local copy)
		b.bucketPtrs[0] = 0
		for k := 0; k < b.numProcs; k++ {
			b.bucketPtrs[0] += b.bucketSize[k][0]
		}

		for i := 1; i < NUM_BUCKETS; i++ {
			b.bucketPtrs[i] = b.bucketPtrs[i-1]
			for k := 0; k < b.numProcs; k++ {
				b.bucketPtrs[i] += b.bucketSize[k][i]
			}
		}

		// Now, buckets are sorted. Sort keys inside each bucket (parallel with dynamic schedule)
		wg.Add(NUM_BUCKETS)
		for i := 0; i < NUM_BUCKETS; i++ {
			go func(bucketID int) {
				defer wg.Done()
				var m, k1, k2 types.INT_TYPE
				// Clear the work array section associated with each bucket
				k1 = types.INT_TYPE(bucketID) * numBucketKeys
				k2 = k1 + numBucketKeys
				for k := k1; k < k2; k++ {
					keyBuffPtr[k] = 0
				}
				// Ranking of all keys occurs in this section
				if bucketID > 0 {
					m = b.bucketPtrs[bucketID-1]
				} else {
					m = 0
				}
				for k := m; k < b.bucketPtrs[bucketID]; k++ {
					keyBuffPtr[keyBuffPtr2[k]]++ // Now they have individual key population
				}
				// To obtain ranks of each key, successively add the individual key
				// population, not forgetting to add m, the total of lesser keys
				keyBuffPtr[k1] += m
				for k := k1 + 1; k < k2; k++ {
					keyBuffPtr[k] += keyBuffPtr[k-1]
				}
			}(i)
		}
		wg.Wait()
	} else {
		// !USE_BUCKETS mode - parallelize work per thread
		var wg sync.WaitGroup
		wg.Add(b.numProcs)

		for myid := 0; myid < b.numProcs; myid++ {
			go func(threadID int) {
				defer wg.Done()
				workBuff := b.keyBuff1Aptr[threadID]
				// Clear the work array
				for i := 0; i < MAX_KEY; i++ {
					workBuff[i] = 0
				}
				// Ranking of all keys occurs in this section
				chunk := (NUM_KEYS + b.numProcs - 1) / b.numProcs
				start := threadID * chunk
				end := start + chunk
				if end > NUM_KEYS {
					end = NUM_KEYS
				}
				for i := start; i < end; i++ {
					workBuff[keyBuffPtr2[i]]++ // Now they have individual key population
				}
			}(myid)
		}
		wg.Wait()

		// To obtain ranks of each key, successively add the individual key population
		// (sequential - needs to be done per thread first, then accumulate)
		for myid := 0; myid < b.numProcs; myid++ {
			workBuff := b.keyBuff1Aptr[myid]
			for i := 0; i < MAX_KEY-1; i++ {
				workBuff[i+1] += workBuff[i]
			}
		}

		// Accumulate the global key population (sequential)
		for k := 1; k < b.numProcs; k++ {
			for i := 0; i < MAX_KEY; i++ {
				keyBuffPtr[i] += b.keyBuff1Aptr[k][i]
			}
		}
	}

	// This is the partial verify test section (sequential - as in C++)
	// Observe that test_rank_array vals are shifted differently for different cases
	for i := 0; i < TEST_ARRAY_SIZE; i++ {
		k := b.partialVerifyVals[i] // test vals were put here
		if 0 < k && k <= NUM_KEYS-1 {
			keyRank := keyBuffPtr[k-1]
			failed := false

			switch params.CLASS {
			case "S":
				if i <= 2 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+iteration {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				}
			case "W":
				if i < 2 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+(iteration-2) {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				}
			case "A":
				if i <= 2 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+(iteration-1) {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-(iteration-1) {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				}
			case "B":
				if i == 1 || i == 2 || i == 4 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+iteration {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				}
			case "C":
				if i <= 2 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+iteration {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				}
			case "D":
				if i < 2 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+iteration {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.verificationMutex.Lock()
						b.passedVerification++
						b.verificationMutex.Unlock()
					}
				}
			}
			if failed {
				fmt.Printf("Failed partial verification: iteration %d, test key %d\n", iteration, i)
			}
		}
	}

	// Make copies of rank info for use by full_verify
	if iteration == MAX_ITERATIONS {
		copy(b.keyBuffPtrGlobal, keyBuffPtr)
		// bucketPtrs is already a field, so it's preserved
	}
}
