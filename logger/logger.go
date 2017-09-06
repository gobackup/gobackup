package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	logFlag = log.Ldate | log.Ltime | log.LUTC
	myLog   = log.New(os.Stdout, "", logFlag)
)

func init() {
	log.Println("Init logger")
}

// Print log
func Print(v ...interface{}) {
	myLog.Print(v...)
}

// Println log
func Println(v ...interface{}) {
	myLog.Println(v...)
}

func Info(v ...interface{}) {
	myLog.Println(v...)
}

// Error log
func Error(v ...interface{}) {
	myLog.Println("[error]", fmt.Sprint(v...))
}
