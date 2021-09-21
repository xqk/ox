package otime

import "time"

//
// Time
// @Description: 时间
//
type Time struct {
	time.Time
}

//
// Now
// @Description: 返回当前时间
// @return *Time
//
func Now() *Time {
	return &Time{
		Time: time.Now(),
	}
}

//
// Unix
// @Description: 返回由时间戳转换的时间
// @param sec
// @param nsec
// @return *Time
//
func Unix(sec, nsec int64) *Time {
	return &Time{
		Time: time.Unix(sec, nsec),
	}
}

//
// Today
// @Description: 今天的开始时间
// @return *Time
//
func Today() *Time {
	return Now().BeginOfDay()
}

//
// BeginOfYear
// @Description: 此时间年的开始时间
// @receiver t
// @return *Time
//
func (t *Time) BeginOfYear() *Time {
	y, _, _ := t.Date()
	return &Time{time.Date(y, time.January, 1, 0, 0, 0, 0, t.Location())}
}

//
// EndOfYear
// @Description: 此时间年的结束时间
// @receiver t
// @return *Time
//
func (t *Time) EndOfYear() *Time {
	return &Time{t.BeginOfYear().AddDate(1, 0, 0).Add(-time.Nanosecond)}
}

//
// BeginOfMonth
// @Description: 此时间年的开始月份
// @receiver t
// @return *Time
//
func (t *Time) BeginOfMonth() *Time {
	y, m, _ := t.Date()
	return &Time{time.Date(y, m, 1, 0, 0, 0, 0, t.Location())}
}

//
// EndOfMonth
// @Description: 此时间年的结束月份
// @receiver t
// @return *Time
//
func (t *Time) EndOfMonth() *Time {
	return &Time{t.BeginOfMonth().AddDate(0, 1, 0).Add(-time.Nanosecond)}
}

//
// BeginOfWeek
// @Description: 此时间周的开始时间，注意：一周的开始日期是星期日
// @receiver t
// @return *Time
//
func (t *Time) BeginOfWeek() *Time {
	y, m, d := t.AddDate(0, 0, 0-int(t.BeginOfDay().Weekday())).Date()
	return &Time{time.Date(y, m, d, 0, 0, 0, 0, t.Location())}
}

//
// EndOfWeek
// @Description: 此时间周的结束时间，注意：一周的结束日期是星期六
// @receiver t
// @return *Time
//
func (t *Time) EndOfWeek() *Time {
	y, m, d := t.BeginOfWeek().AddDate(0, 0, 7).Add(-time.Nanosecond).Date()
	return &Time{time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())}
}

//
// BeginOfDay
// @Description: 此时间天的开始时间
// @receiver t
// @return *Time
//
func (t *Time) BeginOfDay() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, 0, 0, 0, 0, t.Location())}
}

//
// EndOfDay
// @Description: 此时间天的结束时间
// @receiver t
// @return *Time
//
func (t *Time) EndOfDay() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())}
}

//
// BeginOfHour
// @Description: 此时间小时的开始时间
// @receiver t
// @return *Time
//
func (t *Time) BeginOfHour() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())}
}

//
// EndOfHour
// @Description: 此时间小时的结束时间
// @receiver t
// @return *Time
//
func (t *Time) EndOfHour() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, t.Hour(), 59, 59, int(time.Second-time.Nanosecond), t.Location())}
}

//
// BeginOfMinute
// @Description: 此时间分钟的开始时间
// @receiver t
// @return *Time
//
func (t *Time) BeginOfMinute() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, t.Hour(), t.Minute(), 0, 0, t.Location())}
}

//
// EndOfMinute
// @Description: 此时间分钟的结束时间
// @receiver t
// @return *Time
//
func (t *Time) EndOfMinute() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, t.Hour(), t.Minute(), 59, int(time.Second-time.Nanosecond), t.Location())}
}

var TS TimeFormat = "2006-01-02 15:04:05"

type TimeFormat string

func (ts TimeFormat) Format(t time.Time) string {
	return t.Format(string(ts))
}

const (
	DateFormat         = "2006-01-02"
	UnixTimeUnitOffset = uint64(time.Millisecond / time.Nanosecond)
)

// FormatTimeMillis formats Unix timestamp (ms) to time string.

//
// FormatTimeMillis
// @Description: 格式化时间戳（毫秒）为字符串时间格式
// @param tsMillis
// @return string
//
func FormatTimeMillis(tsMillis uint64) string {
	return time.Unix(0, int64(tsMillis*UnixTimeUnitOffset)).Format(string(TS))
}

//
// FormatDate
// @Description: 格式化时间戳（毫秒）为字符串日期格式
// @param tsMillis
// @return string
//
func FormatDate(tsMillis uint64) string {
	return time.Unix(0, int64(tsMillis*UnixTimeUnitOffset)).Format(DateFormat)
}

//
// CurrentTimeMillis
// @Description: 当前时间的时间戳（毫秒）
// @return uint64
//
func CurrentTimeMillis() uint64 {
	// Read from cache first.
	tickerNow := CurrentTimeMillsWithTicker()
	if tickerNow > uint64(0) {
		return tickerNow
	}
	return uint64(time.Now().UnixNano()) / UnixTimeUnitOffset
}

//
// CurrentTimeNano
// @Description: 当前时间的nanoseconds
// @return uint64
//
func CurrentTimeNano() uint64 {
	return uint64(time.Now().UnixNano())
}
