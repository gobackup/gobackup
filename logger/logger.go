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
	myLog   *log.Logger
}

var (
	_logFlag     = log.Ldate | log.Ltime
	TimeFormat   = "2006/01/02 15:04:05"
	_myLog       = log.New(&writer{os.Stdout, TimeFormat}, "", 0)
	sharedLogger Logger
	isTest       = os.Getenv("GO_ENV") == "test"
	isDebug      = os.Getenv("DEBUG") == "true"
)

type writer struct {
	io.Writer
	timeFormat string
}

func (w writer) Write(b []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(time.Now().Format(w.timeFormat)+" "), b...))
}

func init() {
	if isTest {
		if err := os.MkdirAll("../log", 0777); err != nil {
			log.Printf("create log dir failed: %v\n", err)
		}

		logfile, _ := os.OpenFile("../log/test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		_myLog = log.New(logfile, "", _logFlag)
	}
	sharedLogger = newLogger()
}

func SetLogger(logPath string) {
	logfile, _ := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	multi := io.MultiWriter(logfile, os.Stdout)
	_myLog = log.New(&writer{multi, TimeFormat}, "", 0)
	sharedLogger = newLogger()
}

func newLogger() Logger {
	return Logger{_logFlag, _myLog}
}

func Tag(tag string) Logger {
	return sharedLogger.Tag(color.CyanString(fmt.Sprintf("[%s] ", tag)))
}

func (logger Logger) Prefix() string {
	return logger.myLog.Prefix()
}

func (logger Logger) Writer() io.Writer {
	return logger.myLog.Writer()
}

func (logger Logger) Tag(tag string) Logger {
	logger.myLog.SetPrefix(tag)
	return logger
}

func (logger Logger) log(v ...interface{}) {
	logger.myLog.Print(v...)
}

func (logger Logger) logln(v ...interface{}) {
	logger.myLog.Println(v...)
}

// Print log
func (logger Logger) Print(v ...interface{}) {
	logger.log(v...)
}

// Println log
func (logger Logger) Println(v ...interface{}) {
	logger.logln(v...)
}

// Printf log
func (logger Logger) Printf(format string, v ...interface{}) {
	logger.Print(fmt.Sprintf(format, v...))
}

// Debug log
func (logger Logger) Debug(v ...interface{}) {
	if isDebug || isTest {
		logger.logln("[debug] ", fmt.Sprint(v...))
	}
}

// Debugf log
func (logger Logger) Debugf(format string, v ...interface{}) {
	logger.Debug(fmt.Sprintf(format, v...))
}

// Info log
func (logger Logger) Info(v ...interface{}) {
	logger.logln(v...)
}

// Infof log
func (logger Logger) Infof(format string, v ...interface{}) {
	logger.Info(fmt.Sprintf(format, v...))
}

// Warn log
func (logger Logger) Warn(v ...interface{}) {
	c := color.YellowString(fmt.Sprint(v...))
	logger.logln(c)
}

// Warnf log
func (logger Logger) Warnf(format string, v ...interface{}) {
	logger.Warn(fmt.Sprintf(format, v...))
}

// Error log
func (logger Logger) Error(v ...interface{}) {
	c := color.RedString(fmt.Sprint(v...))
	logger.logln(c)
}

// Errorf log
func (logger Logger) Errorf(format string, v ...interface{}) {
	logger.Error(fmt.Sprintf(format, v...))
}

// Fatal log
func (logger Logger) Fatal(v ...interface{}) {
	c := color.MagentaString(fmt.Sprint(v...))
	logger.logln(c)
	os.Exit(1)
}

// Fatalf log
func (logger Logger) Fatalf(format string, v ...interface{}) {
	logger.Fatal(fmt.Sprintf(format, v...))
}

// Print log
func Print(v ...interface{}) {
	sharedLogger.Print(v...)
}

// Printf log
func Printf(format string, v ...interface{}) {
	sharedLogger.Printf(format, v...)
}

// Println log
func Println(v ...interface{}) {
	sharedLogger.Println(v...)
}

// Debug log
func Debug(v ...interface{}) {
	sharedLogger.Debug(v...)
}

// Debugf log
func Debugf(format string, v ...interface{}) {
	sharedLogger.Debugf(format, v...)
}

// Info log
func Info(v ...interface{}) {
	sharedLogger.Info(v...)
}

// Infof log
func Infof(format string, v ...interface{}) {
	sharedLogger.Infof(format, v...)
}

// Warn log
func Warn(v ...interface{}) {
	sharedLogger.Warn(v...)
}

// Warnf log
func Warnf(format string, v ...interface{}) {
	sharedLogger.Warnf(format, v...)
}

// Error log
func Error(v ...interface{}) {
	sharedLogger.Error(v...)
}

// Errorf log
func Errorf(format string, v ...interface{}) {
	sharedLogger.Errorf(format, v...)
}

// Fatal log
func Fatal(v ...interface{}) {
	sharedLogger.Fatal(v...)
}

// Fatalf log
func Fatalf(format string, v ...interface{}) {
	sharedLogger.Fatalf(format, v...)
}
