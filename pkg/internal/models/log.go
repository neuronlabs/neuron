package models

import (
	"github.com/kucjac/uni-logger"
	stdLog "log"
	"os"
)

var log *unilogger.BasicLogger

func init() {
	log = unilogger.NewBasicLogger(os.Stdout, "models", stdLog.Ldate|stdLog.Ltime|stdLog.Lshortfile)
}
