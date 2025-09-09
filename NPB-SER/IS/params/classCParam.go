//go:build C

package params

import (
	"github.com/iyisakuma/NPB-GO/NPB-SER/IS/types"
	"github.com/iyisakuma/NPB-GO/NPB-SER/IS/verifier"
)

const (
	CLASS             = "C"
	TOTAL_KEYS_LOG_2  = 27
	MAX_KEY_LOG_2     = 23
	NUM_BUCKETS_LOG_2 = 10
)

var TEST_INDEX_ARRAY = [5]types.INT_TYPE{44172927, 72999161, 74326391, 129606274, 21736814}
var TEXT_RANK_ARRAY = [5]types.INT_TYPE{61147, 882988, 266290, 133997595, 133525895}
var Verifier verifier.PartialVerifier = &verifier.ClassCVerifier{}
