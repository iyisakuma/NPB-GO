package verifier

import (
	"github.com/iyisakuma/NPB-GO/NPB-SER/IS/types"
)

type ClassAVerifier struct{}

func (m *ClassAVerifier) Do(index int, iteration types.INT_TYPE, keyRank types.INT_TYPE, testRankArray []types.INT_TYPE, passedVerification *int) (failed bool) {
	if index <= 2 {
		if keyRank != testRankArray[index]+(iteration-1) {
			failed = true
		} else {
			*passedVerification++
		}
	} else {
		if keyRank != testRankArray[index]-(iteration-1) {
			failed = true
		} else {
			*passedVerification++
		}
	}
	return
}
