package otime

import "time"

//
// GetTimestampInMilli
// @Description:
// @return int64
//
func GetTimestampInMilli() int64 {
	return int64(time.Now().UnixNano() / 1e6)
}

//
// Elapse
// @Description:消费的时长
// @param f
// @return int64
//
func Elapse(f func()) int64 {
	now := time.Now().UnixNano()
	f()
	return time.Now().UnixNano() - now
}

//
// IsLeapYear
// @Description: 是否是闰年
// @param year
// @return bool
//
func IsLeapYear(year int) bool {
	if year%100 == 0 {
		return year%400 == 0
	}

	return year%4 == 0
}
