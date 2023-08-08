package utils

import (
	"fmt"
	"math/rand"
	"time"
)

/**
 * 生成随机id
 * @return string
 */
func CreateRandId() string {
	rand.Seed(time.Now().UnixNano())

	// 生成随机字符串
	randomStr := make([]byte, 8)
	for i := range randomStr {
		randomStr[i] = byte(rand.Intn(26) + 97) // 生成小写字母
	}

	// 生成时间戳字符串
	timestampStr := fmt.Sprintf("%d", time.Now().UnixNano())

	return string(randomStr) + timestampStr[9:]
}
