package utils

import "time"

// DateOnly 将 time.Time 截断到日期精度（去除时分秒和时区信息）
func DateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// ParseDate 将 "YYYY-MM-DD" 格式字符串解析为 UTC 零时日期
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.UTC)
}

// FormatDate 将时间格式化为 "YYYY-MM-DD" 字符串
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// DatesBetween 返回 start 到 end（含两端）之间所有日期的切片
func DatesBetween(start, end time.Time) []time.Time {
	start = DateOnly(start)
	end = DateOnly(end)
	var dates []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}
	return dates
}

// IsWeekend 判断给定日期是否为周六或周日
func IsWeekend(t time.Time) bool {
	wd := t.Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

// Today 返回服务器本地当天的 UTC 零时日期（统一时区基准）
func Today() time.Time {
	return DateOnly(time.Now().UTC())
}
