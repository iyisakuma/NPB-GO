package main

import (
	"fmt"
	"os"

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
	USE_BUCKET        = true
	TEST_ARRAY_SIZE   = 5
)

// ISBenchmark represents the IS (Integer Sort) benchmark
type ISBenchmark struct {
	keyArray           []types.INT_TYPE
	keyBuff1           []types.INT_TYPE
	keyBuff2           []types.INT_TYPE
	partialVerifyVals  []types.INT_TYPE
	keyBuff1Aptr       [][]types.INT_TYPE // only when !USE_BUCKETS
	passedVerification int
	keyBuffPtrGlobal   []types.INT_TYPE
	bucketSize         [][]types.INT_TYPE // [numProcs][NUM_BUCKETS]
	bucketPtrs         []types.INT_TYPE   // [NUM_BUCKETS]
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

	if USE_BUCKET {
		bench.bucketPtrs = make([]types.INT_TYPE, NUM_BUCKETS)
	}

	return bench
}

func main() {
	bench := NewISBenchmark()
	bench.run()
}

func (b *ISBenchmark) run() {
	timerOn := b.checkTimerFlag()

	common.TimerClear(T_BENCHMARKING)
	if timerOn {
		b.setupTimers()
		common.TimerStart(T_TOTAL_EXECUTION)
	}

	b.printHeader()

	if timerOn {
		common.TimerStart(T_INITIALIZATION)
	}

	b.createSequence(314159265.00, 1220703125.00)
	b.allocKeyBuff()

	if timerOn {
		common.TimerStop(T_INITIALIZATION)
	}

	// Initialization call (untimed)
	b.rank(1)

	b.passedVerification = 0
	if params.CLASS != "S" {
		fmt.Println("\n iteration")
	}

	common.TimerStart(T_BENCHMARKING)

	// Main iterations
	for iteration := types.INT_TYPE(1); iteration <= MAX_ITERATIONS; iteration++ {
		if params.CLASS != "S" {
			fmt.Printf("        %d\n", iteration)
		}
		b.rank(iteration)
	}

	common.TimerStop(T_BENCHMARKING)
	timecounter := common.TimerRead(T_BENCHMARKING)

	if timerOn {
		common.TimerStart(T_SORTING)
	}
	b.fullVerify()
	if timerOn {
		common.TimerStop(T_SORTING)
		common.TimerStop(T_TOTAL_EXECUTION)
	}

	if b.passedVerification != 5*MAX_ITERATIONS+1 {
		b.passedVerification = 0
	}

	b.printResults(timecounter)
}

func (b *ISBenchmark) checkTimerFlag() bool {
	_, err := os.Open("timer.flag")
	return err == nil
}

func (b *ISBenchmark) setupTimers() {
	common.TimerClear(T_INITIALIZATION)
	common.TimerClear(T_SORTING)
	common.TimerClear(T_TOTAL_EXECUTION)
}

func (b *ISBenchmark) printHeader() {
	fmt.Printf("\n\n NAS Parallel Benchmarks 4.1 Serial Go version - IS Benchmark\n\n")
	fmt.Printf(" Size:  %d  (class %v)\n", TOTAL_KEYS, params.CLASS)
	fmt.Printf(" Iterations:   %d\n", MAX_ITERATIONS)
	fmt.Printf("\n")
}

func (b *ISBenchmark) printResults(timecounter float64) {
	common.PrintResults(
		"IS",
		params.CLASS,
		int(TOTAL_KEYS/64),
		64,
		0,
		MAX_ITERATIONS,
		timecounter,
		float64(MAX_ITERATIONS*TOTAL_KEYS)/timecounter/1000000.0,
		"keys ranked",
		b.passedVerification > 0,
		"",
		"",
		"",
		"",
	)
}

func (b *ISBenchmark) allocKeyBuff() {
	numProcs := 1

	if USE_BUCKET {
		b.bucketSize = make([][]types.INT_TYPE, numProcs)
		for i := 0; i < numProcs; i++ {
			b.bucketSize[i] = make([]types.INT_TYPE, NUM_BUCKETS)
		}
		// Initialize keyBuff2
		for i := range b.keyBuff2 {
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

// createSequence generates a sequence of pseudo-random integer keys
func (b *ISBenchmark) createSequence(seed, multiplier float64) {
	myId := 0
	numProcs := 1

	mq := (NUM_KEYS + types.INT_TYPE(numProcs) - 1) / types.INT_TYPE(numProcs)
	k1 := mq * types.INT_TYPE(myId)
	k2 := k1 + mq
	if k2 > NUM_KEYS {
		k2 = NUM_KEYS
	}

	mySeed := b.findMySeed(myId, numProcs, 4*NUM_KEYS, seed, multiplier)
	k := MAX_KEY / 4

	for i := k1; i < k2; i++ {
		x := common.Randlc(&mySeed, multiplier)
		x += common.Randlc(&mySeed, multiplier)
		x += common.Randlc(&mySeed, multiplier)
		x += common.Randlc(&mySeed, multiplier)
		b.keyArray[i] = types.INT_TYPE(float64(k) * x)
	}
}

func (b *ISBenchmark) findMySeed(processorRank, numberProcessor int, numRanNumber uint, seed, constantMultiplier float64) float64 {
	mq := (uint(int(numRanNumber)/4 + numberProcessor - 1)) / uint(numberProcessor)
	nq := mq * 4 * uint(processorRank)

	mySeed := seed
	t2 := constantMultiplier
	kk := nq

	for kk > 1 {
		ik := kk / 2
		if 2*ik == kk {
			common.Randlc(&t2, t2)
			kk = ik
		} else {
			common.Randlc(&mySeed, t2)
			kk--
		}
	}
	return mySeed
}

// rank performs the main ranking/sorting operation for each iteration
func (b *ISBenchmark) rank(iteration types.INT_TYPE) {
	// Set test values for verification
	b.keyArray[iteration] = iteration
	b.keyArray[iteration+MAX_ITERATIONS] = MAX_KEY - iteration

	// Copy verification values
	for i := range TEST_ARRAY_SIZE {
		b.partialVerifyVals[i] = b.keyArray[params.TEST_INDEX_ARRAY[i]]
	}

	var keyBuffPtr []types.INT_TYPE
	if USE_BUCKET {
		keyBuffPtr, _ = b.rankWithBuckets()
	} else {
		keyBuffPtr, _ = b.rankWithoutBuckets()
	}

	// Partial verification
	b.partialVerify(iteration, keyBuffPtr)

	// Store pointer for full_verify on last iteration
	if iteration == MAX_ITERATIONS {
		b.keyBuffPtrGlobal = keyBuffPtr
	}
}

func (b *ISBenchmark) rankWithBuckets() ([]types.INT_TYPE, []types.INT_TYPE) {
	shift := params.MAX_KEY_LOG_2 - params.NUM_BUCKETS_LOG_2
	numBucketKeys := types.INT_TYPE(1) << shift

	keyBuffPtr2 := b.keyBuff2
	keyBuffPtr := b.keyBuff1

	myid, numProcs := 0, 1
	workBuff := b.bucketSize[myid]

	// Clear counts
	for i := range workBuff {
		workBuff[i] = 0
	}

	// Count keys per bucket
	for _, key := range b.keyArray {
		idx := key >> shift
		workBuff[idx]++
	}

	// Calculate accumulated bucket pointers
	b.calculateBucketPointers(myid, numProcs)

	// Distribute keys to buckets
	for _, key := range b.keyArray {
		idx := key >> shift
		pos := b.bucketPtrs[idx]
		if pos < types.INT_TYPE(len(b.keyBuff2)) {
			b.keyBuff2[pos] = key
		}
		b.bucketPtrs[idx]++
	}

	// Adjust pointers to final sizes
	if myid < numProcs-1 {
		for i := range b.bucketPtrs {
			for k := myid + 1; k < numProcs; k++ {
				b.bucketPtrs[i] += b.bucketSize[k][i]
			}
		}
	}

	// Sort within each bucket
	b.sortWithinBuckets(numBucketKeys, keyBuffPtr, keyBuffPtr2)

	return keyBuffPtr, keyBuffPtr2
}

func (b *ISBenchmark) calculateBucketPointers(myid, numProcs int) {
	if b.bucketPtrs == nil || types.INT_TYPE(len(b.bucketPtrs)) != NUM_BUCKETS {
		b.bucketPtrs = make([]types.INT_TYPE, NUM_BUCKETS)
	}

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
}

func (b *ISBenchmark) sortWithinBuckets(numBucketKeys types.INT_TYPE, keyBuffPtr, keyBuffPtr2 []types.INT_TYPE) {
	for i := 0; i < NUM_BUCKETS; i++ {
		k1 := types.INT_TYPE(i) * numBucketKeys
		k2 := k1 + numBucketKeys
		if k2 > MAX_KEY {
			k2 = MAX_KEY
		}

		// Clear work array section for this bucket
		for k := k1; k < k2; k++ {
			keyBuffPtr[k] = 0
		}

		// Count keys in this bucket
		m := types.INT_TYPE(0)
		if i > 0 {
			m = b.bucketPtrs[i-1]
		}
		for k := m; k < b.bucketPtrs[i]; k++ {
			key := keyBuffPtr2[k]
			keyBuffPtr[key]++
		}

		// Calculate cumulative counts
		if k1 < types.INT_TYPE(len(keyBuffPtr)) {
			keyBuffPtr[k1] += m
		}
		for k := k1 + 1; k < k2; k++ {
			keyBuffPtr[k] += keyBuffPtr[k-1]
		}
	}
}

func (b *ISBenchmark) rankWithoutBuckets() ([]types.INT_TYPE, []types.INT_TYPE) {
	keyBuffPtr2 := b.keyArray
	keyBuffPtr := b.keyBuff1

	myid, numProcs := 0, 1
	workBuff := b.keyBuff1Aptr[myid]

	// Clear work array
	for i := range workBuff {
		workBuff[i] = 0
	}

	// Count keys
	for i := 0; i < NUM_KEYS; i++ {
		workBuff[keyBuffPtr2[i]]++
	}

	// Calculate cumulative counts
	for i := 0; i < MAX_KEY-1; i++ {
		workBuff[i+1] += workBuff[i]
	}

	// Accumulate global key population
	for k := 1; k < numProcs; k++ {
		for i := 0; i < MAX_KEY; i++ {
			keyBuffPtr[i] += b.keyBuff1Aptr[k][i]
		}
	}

	return keyBuffPtr, keyBuffPtr2
}

func (b *ISBenchmark) partialVerify(iteration types.INT_TYPE, keyBuffPtr []types.INT_TYPE) {
	for i := 0; i < TEST_ARRAY_SIZE; i++ {
		k := b.partialVerifyVals[i]
		if 0 < k && k <= NUM_KEYS-1 {
			keyRank := keyBuffPtr[k-1]
			failed := params.Verifier.Do(i, iteration, keyRank, params.TEXT_RANK_ARRAY[:], &b.passedVerification)

			if failed {
				fmt.Printf("Failed partial verification: iteration %d, test key %d\n", iteration, i)
			}
		}
	}
}

// fullVerify verifies that all keys are correctly sorted
func (b *ISBenchmark) fullVerify() {
	if USE_BUCKET {
		b.fullVerifyWithBuckets()
	} else {
		b.fullVerifyWithoutBuckets()
	}

	// Check if keys are correctly sorted
	incorrectCount := 0
	for i := 1; i < NUM_KEYS; i++ {
		if b.keyArray[i-1] > b.keyArray[i] {
			incorrectCount++
		}
	}

	if incorrectCount != 0 {
		fmt.Printf("Full_verify: number of keys out of sort: %d\n", incorrectCount)
	} else {
		b.passedVerification++
	}
}

func (b *ISBenchmark) fullVerifyWithBuckets() {
	for j := 0; j < NUM_BUCKETS; j++ {
		k1 := types.INT_TYPE(0)
		if j > 0 {
			k1 = b.bucketPtrs[j-1]
		}

		for i := k1; i < b.bucketPtrs[j]; i++ {
			k := b.keyBuffPtrGlobal[b.keyBuff2[i]] - 1
			b.keyBuffPtrGlobal[b.keyBuff2[i]] = k
			if k < types.INT_TYPE(len(b.keyArray)) {
				b.keyArray[k] = b.keyBuff2[i]
			}
		}
	}
}

func (b *ISBenchmark) fullVerifyWithoutBuckets() {
	// Copy keyArray to keyBuff2
	copy(b.keyBuff2, b.keyArray)

	numProcs := types.INT_TYPE(1)
	myId := types.INT_TYPE(0)

	// Each thread is responsible for a subset of key values
	j := numProcs
	j = (MAX_KEY + j - 1) / j
	k1 := j * myId
	k2 := k1 + j
	if k2 > MAX_KEY {
		k2 = MAX_KEY
	}

	for i := 0; i < NUM_KEYS; i++ {
		if b.keyBuff2[i] >= k1 && b.keyBuff2[i] < k2 {
			k := b.keyBuffPtrGlobal[b.keyBuff2[i]] - 1
			b.keyBuffPtrGlobal[b.keyBuff2[i]] = k
			b.keyArray[k] = b.keyBuff2[i]
		}
	}
}
