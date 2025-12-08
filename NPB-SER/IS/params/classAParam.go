//go:build A

package params

import (
	"github.com/iyisakuma/NPB-GO/NPB-SER/IS/types"
	"github.com/iyisakuma/NPB-GO/NPB-SER/IS/verifier"
)

const (
	CLASS             = "A"
	TOTAL_KEYS_LOG_2  = 23
	MAX_KEY_LOG_2     = 19
	NUM_BUCKETS_LOG_2 = 10
	EmptyTag          = false
)

var TEST_INDEX_ARRAY = [5]types.INT_TYPE{2112377, 662041, 5336171, 3642833, 4250760}

var TEXT_RANK_ARRAY = [5]types.INT_TYPE{104, 17523, 123928, 8288932, 8388264}

var Verifier verifier.PartialVerifier = &verifier.ClassAVerifier{}
