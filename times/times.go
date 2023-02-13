package times

import "time"

// 当前日期时间字符串
func NowDateTimeString() string {
	return time.Now().Format("20060102_150405")
}

func ParseTimeWithLocation(layout, timeStr, localName string) (time.Time, error) {
	if local, err := time.LoadLocation(localName); err != nil {
		return time.Time{}, err
	} else {
		return time.ParseInLocation(layout, timeStr, local)
	}
}
