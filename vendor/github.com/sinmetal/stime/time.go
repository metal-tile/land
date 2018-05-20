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
