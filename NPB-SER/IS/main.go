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
	USE_BUCKET        = true
	TEST_ARRAY_SIZE   = 5
)

var (
	keyArray           []types.INT_TYPE
	keyBuff1           []types.INT_TYPE
	keyBuff2           []types.INT_TYPE
	partialVerifyVals  []types.INT_TYPE
	keyBuff1Aptr       [][]types.INT_TYPE // apenas quando !USE_BUCKETS
	passedVerification int
	keyBuffPtrGlobal   []types.INT_TYPE
	// buckets
	bucketSize [][]types.INT_TYPE // [numProcs][NUM_BUCKETS]
	bucketPtrs []types.INT_TYPE   // [NUM_BUCKETS]
)

func main() {
	timerOn := false
	fp, err := os.Open("timer.flag")
	if err == nil {
		fp.Close()
		timerOn = true
	}
	// alocação dinâmica para evitar arrays massivos no binário
	keyArray = make([]types.INT_TYPE, int(SIZE_OF_BUFFERS))
	keyBuff1 = make([]types.INT_TYPE, int(MAX_KEY))
	keyBuff2 = make([]types.INT_TYPE, int(SIZE_OF_BUFFERS))
	partialVerifyVals = make([]types.INT_TYPE, TEST_ARRAY_SIZE)
	keyBuffPtrGlobal = make([]types.INT_TYPE, int(MAX_KEY)) // tamanho típico para ptrs
	if USE_BUCKET {
		bucketPtrs = make([]types.INT_TYPE, int(NUM_BUCKETS))
	}

	common.TimerClear(T_BENCHMARKING)

	if timerOn {
		common.TimerClear(T_INITIALIZATION)
		common.TimerClear(T_SORTING)
		common.TimerClear(T_TOTAL_EXECUTION)
		common.TimerStart(T_TOTAL_EXECUTION)
	}

	fmt.Printf("\n\n NAS Parallel Benchmarks 4.1 Serial C++ version - IS Benchmark\n\n")
	fmt.Printf(" Size:  %d  (class %v)\n", TOTAL_KEYS, params.CLASS)
	fmt.Printf(" Iterations:   %d\n", MAX_ITERATIONS)
	fmt.Printf("\n")

	if timerOn {
		common.TimerStart(T_INITIALIZATION)
	}

	createSeq(314159265.00, /* Random number gen seed */
		1220703125.00 /* Random number gen mult */)
	allocKeyBuff()

	if timerOn {
		common.TimerStop(T_INITIALIZATION)
	}

	rank(1)

	passedVerification = 0
	if params.CLASS != "S" {
		fmt.Println("\n iteration")
	}

	common.TimerStart(T_BENCHMARKING)

	/* This is the main iteration */
	for iteration := types.INT_TYPE(1); iteration <= MAX_ITERATIONS; iteration++ {
		if params.CLASS != "S" {
			fmt.Printf("        %d\n", iteration)
		}
		rank(iteration)
	}

	/* End of timing, obtain maximum time of all processors */
	common.TimerStop(T_BENCHMARKING)
	timecounter := common.TimerRead(T_BENCHMARKING)

	if timerOn {
		common.TimerStart(T_SORTING)
	}
	fullVeify()
	if timerOn {
		common.TimerStop(T_SORTING)
	}
	if timerOn {
		common.TimerStop(T_TOTAL_EXECUTION)
	}
	if passedVerification != 5*MAX_ITERATIONS+1 {
		passedVerification = 0
	}

	common.PrintResults(
		"IS",
		params.CLASS,
		int(TOTAL_KEYS/64),
		64,
		0,
		MAX_ITERATIONS,
		timecounter,
		float64((MAX_ITERATIONS*TOTAL_KEYS))/timecounter/1000000.0,
		"keys ranked",
		passedVerification > 0,
		"",
		"",
		"",
		"",
	)
}

func allocKeyBuff() {
	numProcs := 1

	if USE_BUCKET {
		// bucket_size = make([][]types.INT_TYPE, numProcs)
		bucketSize = make([][]types.INT_TYPE, numProcs)

		for i := 0; i < numProcs; i++ {
			bucketSize[i] = make([]types.INT_TYPE, NUM_BUCKETS)
		}
		for i := range NUM_KEYS {
			keyBuff2[i] = 0
		}
	} else {
		// key_buff1_aptr = make([][]types.INT_TYPE, numProcs)
		keyBuff1Aptr = make([][]types.INT_TYPE, numProcs)

		keyBuff1Aptr[0] = keyBuff1[:]
		for i := 1; i < numProcs; i++ {
			keyBuff1Aptr[i] = make([]types.INT_TYPE, MAX_KEY)
		}
	}
}

func createSeq(seed, constantMultiplier float64) {
	var k1, k2 types.INT_TYPE
	var x float64
	myId := 0
	numProcs := 1

	mq := (NUM_KEYS + types.INT_TYPE(numProcs) - 1) / types.INT_TYPE(numProcs)
	k1 = mq * types.INT_TYPE(myId)
	k2 = k1 + mq
	if k2 > NUM_KEYS {
		k2 = NUM_KEYS
	}
	mySeed := findMySeed(myId,
		numProcs, 4*NUM_KEYS, seed, constantMultiplier)

	k := MAX_KEY / 4
	for i := k1; i < k2; i++ {
		x = common.Randlc(&mySeed, constantMultiplier)
		x += common.Randlc(&mySeed, constantMultiplier)
		x += common.Randlc(&mySeed, constantMultiplier)
		x += common.Randlc(&mySeed, constantMultiplier)
		keyArray[i] = types.INT_TYPE((float64(k) * x))
	}
}

func findMySeed(processorRank int, numberProcessor int, numRanNumber uint, seed float64, constantMultiplier float64) (mySeed float64) {
	var t2 float64
	var mq, nq, kk, ik uint
	mq = (uint((int(numRanNumber)/4 + numberProcessor - 1)) / uint(numberProcessor))
	nq = mq * 4 * uint(processorRank)

	mySeed = seed
	t2 = constantMultiplier
	kk = nq
	for kk > 1 {
		ik = kk / 2
		if 2*ik == kk {
			common.Randlc(&t2, t2)
			kk = ik
		} else {
			common.Randlc(&mySeed, t2)
			kk = kk - 1
		}
	}
	return
}

func rank(iteration types.INT_TYPE) {

	var keyBuffPtr, keyBuffPtr2 []types.INT_TYPE

	if USE_BUCKET {
		shift := params.MAX_KEY_LOG_2 - params.NUM_BUCKETS_LOG_2
		numBucketKeys := types.INT_TYPE(1) << shift

		// ajustes para verificação
		keyArray[iteration] = iteration
		keyArray[iteration+MAX_ITERATIONS] = MAX_KEY - iteration

		// copiar valores de verificação parcial
		for i := range TEST_ARRAY_SIZE {
			index := params.TEST_INDEX_ARRAY[i]
			partialVerifyVals[i] = keyArray[index]
		}

		keyBuffPtr2 = keyBuff2[:]
		keyBuffPtr = keyBuff1[:]

		myid, numProcs := 0, 1

		workBuff := bucketSize[myid]

		// zera contagens
		for i := 0; i < NUM_BUCKETS; i++ {
			workBuff[i] = 0
		}
		// conta chaves por bucket
		for i := 0; i < NUM_KEYS; i++ {
			idx := keyArray[i] >> shift
			workBuff[idx]++
		}

		// ponteiros acumulados
		if bucketPtrs == nil || types.INT_TYPE(len(bucketPtrs)) != NUM_BUCKETS {
			bucketPtrs = make([]types.INT_TYPE, NUM_BUCKETS)
		}
		bucketPtrs[0] = 0
		for k := 0; k < myid; k++ {
			bucketPtrs[0] += bucketSize[k][0]
		}

		for i := 1; i < NUM_BUCKETS; i++ {
			bucketPtrs[i] = bucketPtrs[i-1]
			for k := 0; k < myid; k++ {
				bucketPtrs[i] += bucketSize[k][i]
			}
			for k := myid; k < numProcs; k++ {
				bucketPtrs[i] += bucketSize[k][i-1]
			}
		}

		// distribui nas regiões de bucket
		for i := 0; i < NUM_KEYS; i++ {
			k := keyArray[i]
			idx := k >> shift
			pos := bucketPtrs[idx]
			keyBuff2[pos] = k
			bucketPtrs[idx] = pos + 1
		}

		// ajusta ponteiros para tamanhos finais
		if myid < numProcs-1 {
			for i := 0; i < NUM_BUCKETS; i++ {
				for k := types.INT_TYPE(myid + 1); k < types.INT_TYPE(numProcs); k++ {
					bucketPtrs[i] += bucketSize[k][i]
				}
			}
		}

		// ordena dentro de cada bucket
		for i := 0; i < NUM_BUCKETS; i++ {
			k1 := types.INT_TYPE(i) * numBucketKeys
			k2 := k1 + numBucketKeys
			for k := k1; k < k2 && k < MAX_KEY; k++ {
				keyBuffPtr[k] = 0
			}
			m := types.INT_TYPE(0)
			if i > 0 {
				m = bucketPtrs[i-1]
			}
			for k := m; k < bucketPtrs[i]; k++ {
				key := keyBuffPtr2[k]
				keyBuffPtr[key]++
			}
			if k1 < types.INT_TYPE(len(keyBuffPtr)) {
				keyBuffPtr[k1] += m
			}
			for k := k1 + 1; k < k2 && k < MAX_KEY; k++ {
				keyBuffPtr[k] += keyBuffPtr[k-1]
			}
		}

	} else {
		// Sem buckets
		keyArray[iteration] = iteration
		keyArray[iteration+MAX_ITERATIONS] = MAX_KEY - iteration

		for i := 0; i < TEST_ARRAY_SIZE; i++ {
			partialVerifyVals[i] = keyArray[params.TEST_INDEX_ARRAY[i]]
		}

		keyBuffPtr2 = keyArray[:]
		keyBuffPtr = keyBuff1[:]

		myid, numProcs := 0, 1
		workBuff := keyBuff1Aptr[myid]

		for i := 0; i < MAX_KEY; i++ {
			workBuff[i] = 0
		}
		for i := 0; i < NUM_KEYS; i++ {
			workBuff[keyBuffPtr2[i]]++
		}
		for i := 0; i < MAX_KEY-1; i++ {
			workBuff[i+1] += workBuff[i]
		}
		for k := 1; k < numProcs; k++ {
			for i := 0; i < MAX_KEY; i++ {
				keyBuffPtr[i] += keyBuff1Aptr[int(k)][i]
			}
		}
	}

	// verificação parcial
	for i := 0; i < TEST_ARRAY_SIZE; i++ {
		k := partialVerifyVals[i]
		if 0 < k && k <= NUM_KEYS-1 {
			keyRank := keyBuffPtr[k-1]
			failed := params.Verifier.Do(i, iteration, keyRank, params.TEXT_RANK_ARRAY[:], &passedVerification)
			// switch CLASS {
			// case 'W':
			// 	if i < 2 {
			// 		if keyRank != testRankArray[i]+(iteration-2) {
			// 			failed = true
			// 		} else {
			// 			passedVerification++
			// 		}
			// 	} else {
			// 		if keyRank != testRankArray[i]-iteration {
			// 			failed = true
			// 		} else {
			// 			passedVerification++
			// 		}
			// 	}

			if failed {
				fmt.Printf("Failed partial verification: iteration %d, test key %d\n", iteration, i)
			}
		}
	}

	// guarda ponteiro para full_verify na última iteração
	if iteration == MAX_ITERATIONS {
		keyBuffPtrGlobal = keyBuffPtr
	}
}

func fullVeify() {
	numProcs := types.INT_TYPE(1)
	myId := types.INT_TYPE(0)
	if USE_BUCKET {
		for j := 0; j < NUM_BUCKETS; j++ {
			var k1 types.INT_TYPE
			if j > 0 {
				k1 = bucketPtrs[j-1]
			} else {
				k1 = 0
			}
			for i := k1; i < bucketPtrs[j]; i++ {
				for i = k1; i < bucketPtrs[j]; i++ {
					k := keyBuffPtrGlobal[keyBuff2[i]] - 1
					keyBuffPtrGlobal[keyBuff2[i]] = k
					keyArray[k] = keyBuff2[i]
				}
			}
		}
	} else {
		// Copia keyArray para keyBuff2
		copy(keyBuff2[:], keyArray[:])
		var j types.INT_TYPE
		// Cada thread é responsável por um subconjunto de valores de chave
		j = numProcs
		j = (MAX_KEY + j - 1) / j
		k1 := j * myId
		k2 := k1 + j
		if k2 > MAX_KEY {
			k2 = MAX_KEY
		}

		for i := 0; i < NUM_KEYS; i++ {
			if keyBuff2[i] >= k1 && keyBuff2[i] < k2 {
				k := keyBuffPtrGlobal[keyBuff2[i]] - 1
				keyBuffPtrGlobal[keyBuff2[i]] = k
				keyArray[k] = keyBuff2[i]
			}
		}
	}
	// Confirmar se as chaves estão corretamente ordenadas
	j := 0
	for i := 1; i < NUM_KEYS; i++ {
		if keyArray[i-1] > keyArray[i] {
			j++
		}
	}

	if j != 0 {
		fmt.Printf("Full_verify: número de chaves fora de ordem: %d\n", j)
	} else {
		passedVerification++
	}
}
