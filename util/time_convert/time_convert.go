package timeconvert

import (
	"log"
	"time"
)

// 将字符串转化为time.Time, 如果出错, panic
func MustStrToTime(s string) time.Time {
	res, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		log.Panicf("failed to time.Parse: %v\n", err)
	}
	return res
}

// time.Time转化成字符串
func TimeToStr(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
