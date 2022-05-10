package log

import (
	"log"
	"os"
	"path/filepath"

	"github.com/iam1912/gemseries/gemdelayqueue/utils"
)

var (
	errorLog           = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog            = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	serveFileInfoLog   *log.Logger
	serveFileErrorlog  *log.Logger
	clientFileInfoLog  *log.Logger
	clientFileErrorLog *log.Logger
)

var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

func InitServeFileLogger(infoLog, errorLog string) error {
	name := filepath.Dir(infoLog)
	ok := utils.Exists(name)
	if !ok {
		os.MkdirAll(name, os.ModePerm)
	}
	fileInfo, err := os.OpenFile(infoLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	serveFileInfoLog = log.New(fileInfo, "[info ] ", log.LstdFlags|log.Lshortfile)
	fileError, err := os.OpenFile(errorLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	serveFileErrorlog = log.New(fileError, "[error ] ", log.LstdFlags|log.Lshortfile)
	return nil
}

func InitClientFileLogger(infoLog, errorLog string) error {
	name := filepath.Dir(infoLog)
	ok := utils.Exists(name)
	if !ok {
		os.MkdirAll(name, os.ModePerm)
	}
	fileInfo, err := os.OpenFile(infoLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	clientFileInfoLog = log.New(fileInfo, "[info ] ", log.LstdFlags|log.Lshortfile)
	fileError, err := os.OpenFile(errorLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	clientFileErrorLog = log.New(fileError, "[error ] ", log.LstdFlags|log.Lshortfile)
	return nil
}

func ServeInfo(format string, v ...any) {
	Infof(format, v)
	serveFileInfoLog.Printf(format, v...)
}

func ServeError(format string, v ...any) {
	Errorf(format, v)
	serveFileErrorlog.Printf(format, v...)
}

func ClientInfo(format string, v ...any) {
	Infof(format, v)
	clientFileInfoLog.Printf(format, v...)
}

func ClientError(format string, v ...any) {
	Errorf(format, v)
	clientFileErrorLog.Printf(format, v...)
}
