package lib

import (
	"io"
	"log"
	"os"
	"fmt"
)

type Loger struct {
	Log log.Logger
}

func (obj *Loger) Init() {
	logFile, err := os.OpenFile("../log/Messages.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	obj.Log.SetOutput(mw)
	obj.Log.SetFlags(log.Default().Flags())
}

func (obj *Loger) Println(args ...interface{} ) {
	logstr := ""
	for _, v := range args {
		logstr+=fmt.Sprint(v)+" "
	}
	obj.Log.Println(logstr)
}