package logger

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
)

var (
	logFlag = log.Ldate | log.Ltime
	myLog   = log.New(os.Stdout, "", logFlag)
)

func init() {
	isTest := os.Getenv("GO_ENV") == "test"
	if isTest {
		os.MkdirAll("../log", 0777)
		logfile, _ := os.OpenFile("../log/test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		myLog = log.New(logfile, "", logFlag)
	}
}

// Print log
func Print(v ...interface{}) {
	myLog.Print(v...)
}

// Println log
func Println(v ...interface{}) {
	myLog.Println(v...)
}

// Debug log
func Debug(v ...interface{}) {
	myLog.Println("[debug]", fmt.Sprint(v...))
}

// Info log
func Info(v ...interface{}) {
	myLog.Println(v...)
}

// Warn log
func Warn(v ...interface{}) {
	c := color.YellowString(fmt.Sprint(v...))
	myLog.Println(c)
}

// Error log
func Error(v ...interface{}) {
	c := color.RedString(fmt.Sprint(v...))
	myLog.Println(c)
}
