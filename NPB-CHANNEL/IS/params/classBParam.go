//go:build B

package params

import (
	"github.com/iyisakuma/NPB-GO/NPB-CHANNEL/IS/types"
	"github.com/iyisakuma/NPB-GO/NPB-CHANNEL/IS/verifier"
)

const (
	CLASS             = "B"
	TOTAL_KEYS_LOG_2  = 25
	MAX_KEY_LOG_2     = 21
	NUM_BUCKETS_LOG_2 = 10
)

var TEST_INDEX_ARRAY = [5]types.INT_TYPE{41869, 812306, 5102857, 18232239, 26860214}
var TEXT_RANK_ARRAY = [5]types.INT_TYPE{33422937, 10244, 59149, 33135281, 99}

var Verifier verifier.PartialVerifier = &verifier.ClassBVerifier{}
