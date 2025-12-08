package main

import (
	"fmt"
	"os"

	"github.com/iyisakuma/NPB-GO/NPB-SER/IS/params"
	"github.com/iyisakuma/NPB-GO/NPB-SER/IS/types"
	"github.com/iyisakuma/NPB-GO/NPB-SER/common"
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
}

// NewISBenchmark creates a new IS benchmark instance
func NewISBenchmark() *ISBenchmark {
	bench := &ISBenchmark{
		keyArray:          make([]types.INT_TYPE, SIZE_OF_BUFFERS),
		keyBuff1:          make([]types.INT_TYPE, MAX_KEY),
		keyBuff2:          make([]types.INT_TYPE, SIZE_OF_BUFFERS),
		partialVerifyVals: make([]types.INT_TYPE, TEST_ARRAY_SIZE),
		keyBuffPtrGlobal:  make([]types.INT_TYPE, MAX_KEY),
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
	fmt.Printf(" Size            =                    %d\n", TOTAL_KEYS)
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
	numProcs := 1

	if USE_BUCKETS {
		b.bucketSize = make([][]types.INT_TYPE, numProcs)
		for i := 0; i < numProcs; i++ {
			b.bucketSize[i] = make([]types.INT_TYPE, NUM_BUCKETS)
		}

		// Initialize keyBuff2
		for i := 0; i < NUM_KEYS; i++ {
			b.keyBuff2[i] = 0
		}
	} else {
		b.keyBuff1Aptr = make([][]types.INT_TYPE, numProcs)
		b.keyBuff1Aptr[0] = b.keyBuff1
		for i := 1; i < numProcs; i++ {
			b.keyBuff1Aptr[i] = make([]types.INT_TYPE, MAX_KEY)
		}
	}
}

// findMySeed returns parallel random number seq seed
// Equivalent to find_my_seed in C++
func findMySeed(kn int, np int, nn int64, s float64, a float64) float64 {
	/*
	 * Create a random number sequence of total length nn residing
	 * on np number of processors.  Each processor will therefore have a
	 * subsequence of length nn/np.  This routine returns that random
	 * number which is the first random number for the subsequence belonging
	 * to processor rank kn, and which is used as seed for proc kn ran # gen.
	 */
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
			t2 = common.Randlc(&t2, t2)
			kk = ik
		} else {
			t1 = common.Randlc(&t1, t2)
			kk = kk - 1
		}
	}
	t1 = common.Randlc(&t1, t2)

	return t1
}

// createSeq generates random number sequence and subsequent keys
// Equivalent to create_seq in C++
func (b *ISBenchmark) createSeq(seed float64, a float64) {
	var x, s float64
	var k types.INT_TYPE

	myid := 0
	numProcs := 1

	mq := (NUM_KEYS + numProcs - 1) / numProcs
	k1 := mq * myid
	k2 := k1 + mq
	if k2 > NUM_KEYS {
		k2 = NUM_KEYS
	}

	s = findMySeed(myid, numProcs, int64(4*NUM_KEYS), seed, a)

	k = MAX_KEY / 4

	for i := k1; i < k2; i++ {
		x = common.Randlc(&s, a)
		x += common.Randlc(&s, a)
		x += common.Randlc(&s, a)
		x += common.Randlc(&s, a)
		b.keyArray[i] = types.INT_TYPE(float64(k) * x)
	}
}

// fullVerify verifies that all keys are correctly sorted
// Equivalent to full_verify in C++
func (b *ISBenchmark) fullVerify() {
	var k, k1, k2 types.INT_TYPE

	myid := 0
	numProcs := 1

	// Now, finally, sort the keys:
	// Copy keys into work array; keys in key_array will be reassigned.

	if USE_BUCKETS {
		// Buckets are already sorted. Sorting keys within each bucket
		for j := 0; j < NUM_BUCKETS; j++ {
			if j > 0 {
				k1 = b.bucketPtrs[j-1]
			} else {
				k1 = 0
			}
			for i := k1; i < b.bucketPtrs[j]; i++ {
				// Equivalent to: k = --key_buff_ptr_global[key_buff2[i]];
				// In C++, key_buff_ptr_global points to key_buff1, so we use keyBuffPtrGlobal
				// which is a copy of keyBuff1 from the last iteration
				k = b.keyBuffPtrGlobal[b.keyBuff2[i]] - 1
				b.keyBuffPtrGlobal[b.keyBuff2[i]] = k
				b.keyArray[k] = b.keyBuff2[i]
			}
		}
	} else {
		// Copy keyArray to keyBuff2
		for i := 0; i < NUM_KEYS; i++ {
			b.keyBuff2[i] = b.keyArray[i]
		}
		// This is actual sorting. Each thread is responsible for a subset of key values
		j := numProcs
		j = (MAX_KEY + j - 1) / j
		k1 = types.INT_TYPE(j * myid)
		k2 = k1 + types.INT_TYPE(j)
		if k2 > MAX_KEY {
			k2 = MAX_KEY
		}
		for i := 0; i < NUM_KEYS; i++ {
			if b.keyBuff2[i] >= k1 && b.keyBuff2[i] < k2 {
				// Equivalent to: k = --key_buff_ptr_global[key_buff2[i]];
				k = b.keyBuffPtrGlobal[b.keyBuff2[i]] - 1
				b.keyBuffPtrGlobal[b.keyBuff2[i]] = k
				b.keyArray[k] = b.keyBuff2[i]
			}
		}
	}

	// Confirm keys correctly sorted: count incorrectly sorted keys, if any
	j := 0
	for i := 1; i < NUM_KEYS; i++ {
		if b.keyArray[i-1] > b.keyArray[i] {
			j++
		}
	}
	if j != 0 {
		fmt.Printf("Full_verify: number of keys out of sort: %d\n", j)
	} else {
		b.passedVerification++
	}
}

// rank performs the main ranking/sorting operation for each iteration
// Equivalent to rank in C++
func (b *ISBenchmark) rank(iteration types.INT_TYPE) {
	var shift int
	var numBucketKeys types.INT_TYPE
	var keyBuffPtr, keyBuffPtr2 []types.INT_TYPE
	var workBuff []types.INT_TYPE
	var m, k1, k2 types.INT_TYPE

	myid := 0
	numProcs := 1

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
		workBuff = b.bucketSize[myid]

		// Initialize
		for i := 0; i < NUM_BUCKETS; i++ {
			workBuff[i] = 0
		}

		// Determine the number of keys in each bucket
		for i := 0; i < NUM_KEYS; i++ {
			workBuff[b.keyArray[i]>>shift]++
		}

		// Accumulative bucket sizes are the bucket pointers.
		// These are global sizes accumulated upon to each bucket
		b.bucketPtrs[0] = 0
		for k := 0; k < myid; k++ {
			b.bucketPtrs[0] += b.bucketSize[k][0]
		}

		for i := 1; i < NUM_BUCKETS; i++ {
			b.bucketPtrs[i] = b.bucketPtrs[i-1]
			for k := 0; k < myid; k++ {
				b.bucketPtrs[i] += b.bucketSize[k][i]
			}
			for k := myid; k < numProcs; k++ {
				b.bucketPtrs[i] += b.bucketSize[k][i-1]
			}
		}

		// Sort into appropriate bucket
		for i := 0; i < NUM_KEYS; i++ {
			k := b.keyArray[i]
			b.keyBuff2[b.bucketPtrs[k>>shift]] = k
			b.bucketPtrs[k>>shift]++
		}

		// The bucket pointers now point to the final accumulated sizes
		if myid < numProcs-1 {
			for i := 0; i < NUM_BUCKETS; i++ {
				for k := myid + 1; k < numProcs; k++ {
					b.bucketPtrs[i] += b.bucketSize[k][i]
				}
			}
		}

		// Now, buckets are sorted.  We only need to sort keys inside
		// each bucket, which can be done in parallel.  Because the distribution
		// of the number of keys in the buckets is Gaussian, the use of
		// a dynamic schedule should improve load balance, thus, performance
		for i := 0; i < NUM_BUCKETS; i++ {
			// Clear the work array section associated with each bucket
			k1 = types.INT_TYPE(i) * numBucketKeys
			k2 = k1 + numBucketKeys
			for k := k1; k < k2; k++ {
				keyBuffPtr[k] = 0
			}
			// Ranking of all keys occurs in this section:
			// In this section, the keys themselves are used as their
			// own indexes to determine how many of each there are: their
			// individual population
			if i > 0 {
				m = b.bucketPtrs[i-1]
			} else {
				m = 0
			}
			for k := m; k < b.bucketPtrs[i]; k++ {
				keyBuffPtr[keyBuffPtr2[k]]++ // Now they have individual key population
			}
			// To obtain ranks of each key, successively add the individual key
			// population, not forgetting to add m, the total of lesser keys,
			// to the first key population
			keyBuffPtr[k1] += m
			for k := k1 + 1; k < k2; k++ {
				keyBuffPtr[k] += keyBuffPtr[k-1]
			}
		}
	} else {
		workBuff = b.keyBuff1Aptr[myid]
		// Clear the work array
		for i := 0; i < MAX_KEY; i++ {
			workBuff[i] = 0
		}
		// Ranking of all keys occurs in this section:
		// In this section, the keys themselves are used as their
		// own indexes to determine how many of each there are: their
		// individual population
		for i := 0; i < NUM_KEYS; i++ {
			workBuff[keyBuffPtr2[i]]++ // Now they have individual key population
		}
		// To obtain ranks of each key, successively add the individual key population
		for i := 0; i < MAX_KEY-1; i++ {
			workBuff[i+1] += workBuff[i]
		}
		// Accumulate the global key population
		for k := 1; k < numProcs; k++ {
			for i := 0; i < MAX_KEY; i++ {
				keyBuffPtr[i] += b.keyBuff1Aptr[k][i]
			}
		}
	}

	// This is the partial verify test section
	// Observe that test_rank_array vals are
	// shifted differently for different cases
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
						b.passedVerification++
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.passedVerification++
					}
				}
			case "W":
				if i < 2 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+(iteration-2) {
						failed = true
					} else {
						b.passedVerification++
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.passedVerification++
					}
				}
			case "A":
				if i <= 2 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+(iteration-1) {
						failed = true
					} else {
						b.passedVerification++
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-(iteration-1) {
						failed = true
					} else {
						b.passedVerification++
					}
				}
			case "B":
				if i == 1 || i == 2 || i == 4 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+iteration {
						failed = true
					} else {
						b.passedVerification++
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.passedVerification++
					}
				}
			case "C":
				if i <= 2 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+iteration {
						failed = true
					} else {
						b.passedVerification++
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.passedVerification++
					}
				}
			case "D":
				if i < 2 {
					if keyRank != params.TEXT_RANK_ARRAY[i]+iteration {
						failed = true
					} else {
						b.passedVerification++
					}
				} else {
					if keyRank != params.TEXT_RANK_ARRAY[i]-iteration {
						failed = true
					} else {
						b.passedVerification++
					}
				}
			}
			if failed {
				fmt.Printf("Failed partial verification: iteration %d, test key %d\n", iteration, i)
			}
		}
	}

	// Make copies of rank info for use by full_verify: these variables
	// in rank are local; making them global slows down the code, probably
	// since they cannot be made register by compiler
	// In C++, key_buff_ptr_global = key_buff_ptr just copies the pointer,
	// and both point to key_buff1. In Go, we need to copy the data.
	if iteration == MAX_ITERATIONS {
		copy(b.keyBuffPtrGlobal, keyBuffPtr)
		// Also copy bucketPtrs for USE_BUCKETS mode
		if USE_BUCKETS {
			// bucketPtrs is already a field, so it's preserved
			// But we need to make sure it's not modified after this
		}
	}
}
