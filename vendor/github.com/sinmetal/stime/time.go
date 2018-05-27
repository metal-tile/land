package stime

import (
	"time"
)

var permafrost time.Time
var dummyTime []time.Time

// Now is Get Now Time
func Now() time.Time {
	if !permafrost.IsZero() {
		return permafrost
	}
	if len(dummyTime) < 1 {
		return time.Now()
	}
	end := len(dummyTime) - 1
	t := dummyTime[end]
	dummyTime = append(dummyTime[:end-1])
	return t
}

// InTime is targetが現在時刻から、指定時間と比べて、時間内に収まっているのかを判定する
func InTime(now time.Time, target time.Time, duration time.Duration) bool {
	t := target.Add(duration)
	if now.Equal(t) {
		return true
	}
	if now.Before(t) {
		return true
	}

	return false
}

// AddDummyTime is UnitTest用のNowが返す時刻を追加する
// SetPermafrost と同時に利用した場合、Permafrostが優先される
func AddDummyTime(t ...time.Time) {
	dummyTime = append(dummyTime, t...)
}

// SetPermafrost is UnitTest用のNowが返す時刻を固定する
// AddDummyTime と同時に利用した場合、Permafrostが優先される
func SetPermafrost(t time.Time) {
	permafrost = t
}
