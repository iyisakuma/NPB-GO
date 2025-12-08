//go:build D

package params

import (
	"github.com/iyisakuma/NPB-GO/NPB-SER/IS/types"
	"github.com/iyisakuma/NPB-GO/NPB-SER/IS/verifier"
)

const (
	CLASS             = "D"
	TOTAL_KEYS_LOG_2  = 31
	MAX_KEY_LOG_2     = 27
	NUM_BUCKETS_LOG_2 = 10
	EmptyTag          = false
)

var TEST_INDEX_ARRAY = [5]types.INT_TYPE{1317351170, 995930646, 1157283250, 1503301535, 1453734525}
var TEXT_RANK_ARRAY = [5]types.INT_TYPE{1, 36538729, 1978098519, 2145192618, 2147425337}

var Verifier verifier.PartialVerifier = &verifier.ClassDVerifier{}
