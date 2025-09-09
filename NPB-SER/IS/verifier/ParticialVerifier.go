package verifier

import "github.com/iyisakuma/NPB-GO/NPB-SER/IS/types"

type PartialVerifier interface {
	Do(index int, iteration types.INT_TYPE, keyRank types.INT_TYPE, testRankArray []types.INT_TYPE, passedVerification *int) (failed bool)
}
