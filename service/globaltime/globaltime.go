package globaltime

import "time"

// FixedTime represent a fixed moment in time. Set this variable to anything different from the default value for
// time.Time and the value will be returned in Now() function in place of the current time
var FixedTime time.Time

// Now returns the current time (time.Now()) if no FixedTime has been set. Otherwise, it returns FixedTime.
// Use this in place of time.Now() to allow testing w/ custom time.
func Now() time.Time {
	if FixedTime.After(time.Time{}) {
		return FixedTime
	}
	return time.Now()
}

// Since returns the time passed since the parameter tm.
func Since(tm time.Time) time.Duration {
	return Now().Sub(tm)
}
