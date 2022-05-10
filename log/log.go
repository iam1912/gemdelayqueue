package log

import (
	"log"
	"os"
	"path/filepath"
)

var (
	errorLog     = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog      = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	fileInfoLog  *log.Logger
	fileErrorlog *log.Logger
)

var (
	Error      = errorLog.Println
	Errorf     = errorLog.Printf
	Info       = infoLog.Println
	Infof      = infoLog.Printf
	FileInfo   = fileInfoLog.Println
	FileInfof  = fileInfoLog.Printf
	FileError  = fileErrorlog.Println
	FileErrorf = fileErrorlog.Printf
)

func InitFileLogger(infoLog, errorLog string) {
	file := filepath.Dir(infoLog)
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		os.MkdirAll(file, os.ModePerm)
	}
	fileInfo, err := os.OpenFile(infoLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic("info file log failed:" + err.Error())
	}
	fileInfoLog = log.New(fileInfo, "[info ] ", log.LstdFlags|log.Lshortfile)
	FileInfo = fileInfoLog.Println
	FileInfof = fileInfoLog.Printf

	fileError, err := os.OpenFile(errorLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic("error file log failed:" + err.Error())
	}
	fileErrorlog = log.New(fileError, "[error ] ", log.LstdFlags|log.Lshortfile)
	FileError = fileErrorlog.Println
	FileErrorf = fileErrorlog.Printf
}

func InfofOutStdoutFile(format string, v ...any) {
	Infof(format, v)
	FileInfof(format, v)
}

func ErrorfOutStdoutFile(format string, v ...any) {
	Errorf(format, v)
	FileErrorf(format, v)
}
