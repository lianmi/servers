package dateutil

import (
	"fmt"
	"time"
)

//返回当天年月日字符串 ： 2020-10-06
func GetDateString() string {
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	// hour := time.Now().Format("15")
	// min := time.Now().Format("04")
	// second := time.Now().Format("05")

	return fmt.Sprintf("%s-%s-%s", year, month, day)
}

//返回当天年月字符串 ： 2020-10
func GetYearMonthString() string {
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	// day := time.Now().Format("02")
	// hour := time.Now().Format("15")
	// min := time.Now().Format("04")
	// second := time.Now().Format("05")

	return fmt.Sprintf("%s-%s", year, month)
}
