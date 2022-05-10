package log

import (
	"testing"

	"github.com/iam1912/gemseries/gemdelayqueue/config"
)

func TestFileLog(t *testing.T) {
	c := config.MustGetConfig()
	InitFileLogger(c.InfoLog, c.ErrorLog)
	FileInfo("HELLO WORLD")
}
