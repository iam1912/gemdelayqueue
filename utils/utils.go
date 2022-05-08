package utils

import (
	"fmt"
	"strconv"
)

func StringToInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func StringToInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func SpliceKey(topic, id string) string {
	return fmt.Sprintf("%s-%s", topic, id)
}
