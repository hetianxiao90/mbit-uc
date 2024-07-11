package main

import (
	"fmt"
	"time"
	"uc/internal/util"
)

func main() {
	// 记录开始时间
	startTime := time.Now()
	for i := 0; i < 100000; i++ {
		fmt.Println(util.RandInt64(20000000000, 99999999999))
	}
	// 记录结束时间
	endTime := time.Now()

	// 计算执行时间
	duration := endTime.Sub(startTime)

	// 输出执行时间
	fmt.Printf("代码执行时间： %v ", duration)

}
