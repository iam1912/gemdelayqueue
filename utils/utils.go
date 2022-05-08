package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/iam1912/gemseries/gemdelayqueue/consts"
)

func StringToInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func StringToInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func GetJobKey(topic, id string) string {
	return fmt.Sprintf("%s:%s-%s", consts.JobPrefix, topic, id)
}

func GetBucket(s string, i int) string {
	return fmt.Sprintf("%s:%s-%d", consts.BucketPrefix, s, i)
}

func InferRootDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var infer func(d string) string
	infer = func(d string) string {
		if exists(d + "/config.yml") {
			return d
		}
		return infer(filepath.Dir(d))
	}

	return infer(cwd)
}
func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
