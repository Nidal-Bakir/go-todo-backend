package utils

import (
	"fmt"
	"math"
)

func SafeIntToInt32(v int) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("value %d out of range for int32", v)
	}
	return int32(v), nil
}
