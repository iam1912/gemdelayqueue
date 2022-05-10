package log

import (
	"testing"

	"github.com/iam1912/gemseries/gemdelayqueue/config"
)

func TestFileLog(t *testing.T) {
	c := config.MustGetConfig()
	InitServeFileLogger(c.ServeInfoLog, c.ServeErrorLog)
	InitClientFileLogger(c.ClientInfoLog, c.ClientErrorLog)
	clientFileInfoLog.Printf("HELLOWORLD\n")
}
