package times

import "time"

// 当前日期时间字符串
func NowDateTimeString() string {
	return time.Now().Format("20060102_150405")
}
