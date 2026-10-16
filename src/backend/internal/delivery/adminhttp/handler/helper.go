package handler

import (
	"strconv"
	"time"
)

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func timeNowUTC() time.Time {
	return time.Now().UTC()
}
