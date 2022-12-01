package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

type Logger struct {
	logFlag int
	myLog   log.Logger
}

var (
	_logFlag     = log.Ldate | log.Ltime
	_myLog       = log.New(&writer{os.Stdout, "2006/01/02 15:04:05"}, "", 0)
	sharedLogger Logger
)

type writer struct {
	io.Writer
	timeFormat string
}

func (w writer) Write(b []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(time.Now().Format(w.timeFormat)+" "), b...))
}

func init() {
	isTest := os.Getenv("GO_ENV") == "test"
	if isTest {
		os.MkdirAll("../log", 0777)
		logfile, _ := os.OpenFile("../log/test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		_myLog = log.New(logfile, "", _logFlag)
	}
	sharedLogger = newLogger()
}

func newLogger() Logger {
	return Logger{_logFlag, *_myLog}
}

func Tag(tag string) Logger {
	return sharedLogger.Tag(color.CyanString(fmt.Sprintf("[%s] ", tag)))
}

func (logger Logger) Tag(tag string) Logger {
	logger.myLog.SetPrefix(tag)
	return logger
}

// Print log
func (logger Logger) Print(v ...interface{}) {
	logger.myLog.Print(v...)
}

// Println log
func (logger Logger) Println(v ...interface{}) {
	logger.myLog.Println(v...)
}

// Debug log
func (logger Logger) Debug(v ...interface{}) {
	logger.myLog.Println("[debug]", fmt.Sprint(v...))
}

// Info log
func (logger Logger) Info(v ...interface{}) {
	logger.Println(v...)
}

// Warn log
func (logger Logger) Warn(v ...interface{}) {
	c := color.YellowString(fmt.Sprint(v...))
	logger.myLog.Println(c)
}

// Error log
func (logger Logger) Error(v ...interface{}) {
	c := color.RedString(fmt.Sprint(v...))
	logger.myLog.Println(c)
}

// Print log
func Print(v ...interface{}) {
	sharedLogger.Print(v...)
}

// Println log
func Println(v ...interface{}) {
	sharedLogger.Println(v...)
}

// Debug log
func Debug(v ...interface{}) {
	sharedLogger.Debug(v...)
}

// Info log
func Info(v ...interface{}) {
	sharedLogger.Info(v...)
}

// Warn log
func Warn(v ...interface{}) {
	sharedLogger.Warn(v...)
}

// Error log
func Error(v ...interface{}) {
	sharedLogger.Error(v...)
}
