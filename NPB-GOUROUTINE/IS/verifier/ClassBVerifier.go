package verifier

import (
	"github.com/iyisakuma/NPB-GO/NPB-GOUROUTINE/IS/types"
)

type ClassBVerifier struct{}

func (m *ClassBVerifier) Do(index int, iteration types.INT_TYPE, keyRank types.INT_TYPE, testRankArray []types.INT_TYPE, passedVerification *int) (failed bool) {
	if index == 1 || index == 2 || index == 4 {
		if keyRank != testRankArray[index]+iteration {
			failed = true
		} else {
			*passedVerification++
		}
	} else {
		if keyRank != testRankArray[index]-iteration {
			failed = true
		} else {
			*passedVerification++
		}
	}
	return
}
