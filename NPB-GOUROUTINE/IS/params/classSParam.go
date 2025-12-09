//go:build S

package params

import (
	"github.com/iyisakuma/NPB-GO/NPB-GOUROUTINE/IS/types"
	"github.com/iyisakuma/NPB-GO/NPB-GOUROUTINE/IS/verifier"
)

const (
	CLASS             = "S"
	TOTAL_KEYS_LOG_2  = 16
	MAX_KEY_LOG_2     = 11
	NUM_BUCKETS_LOG_2 = 9
	EmptyTag          = false
)

var TEST_INDEX_ARRAY = [5]types.INT_TYPE{48427, 17148, 23627, 62548, 4431}

var TEXT_RANK_ARRAY = [5]types.INT_TYPE{0, 18, 346, 64917, 65463}

var Verifier verifier.PartialVerifier = &verifier.ClassSVerifier{}
