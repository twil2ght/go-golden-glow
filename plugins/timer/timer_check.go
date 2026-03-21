package timer

import (
	"goldenglow/plugins/checker"
	"strconv"
)

var (
	WithinRangeEvent = "timer_within_range"
)

// WithinRange 检查目标值是否在 (min, max) 开区间内
// 参数格式：tar & min & max
// 规则：左右均为开区间，即 tar > min && tar < max
func WithinRange(c *checker.Context) bool {
	if c == nil || len(c.Payload) != 3 {
		return false
	}

	var tar, min, max int
	var err error

	tar, err = strconv.Atoi(c.Payload[0])
	if err != nil {
		return false
	}
	min, err = strconv.Atoi(c.Payload[1])
	if err != nil {
		return false
	}
	max, err = strconv.Atoi(c.Payload[2])
	if err != nil {
		return false
	}

	if min >= max {
		return false
	}

	return tar > min && tar < max
}
