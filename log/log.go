package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
)

type Logger struct {
	TraceLevel int
	trace      *log.Logger
	info       *log.Logger
	warn       *log.Logger
	error      *log.Logger
}

func New(out io.Writer) *Logger {
	logger := new(Logger)
	logger.trace = log.New(out, "[T] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logger.info = log.New(out, "[I] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logger.warn = log.New(out, "[W] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logger.error = log.New(out, "[E] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	return logger
}

var std = New(os.Stderr)

func Println(v ...interface{}) {
	std.outputln(std.info, 0, v...)
}

func Printf(format string, v ...interface{}) {
	std.outputf(std.info, 0, format, v...)
}

func Trace(v ...interface{}) {
	std.outputln(std.trace, 0, v...)
}

func Tracef(format string, v ...interface{}) {
	std.outputf(std.trace, 0, format, v...)
}

func Info(v ...interface{}) {
	std.outputln(std.info, 0, v...)
}

func Infof(format string, v ...interface{}) {
	std.outputf(std.info, 0, format, v...)
}

func Warn(v ...interface{}) {
	std.outputln(std.warn, std.TraceLevel, v...)
}

func Warnf(format string, v ...interface{}) {
	std.outputf(std.warn, std.TraceLevel, format, v...)
}

func Error(v ...interface{}) {
	std.outputln(std.error, std.TraceLevel, v...)
}

func Errorf(format string, v ...interface{}) {
	std.outputf(std.error, std.TraceLevel, format, v...)
}

func Fatal(v ...interface{}) {
	std.outputln(std.error, std.TraceLevel, v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	std.outputf(std.error, std.TraceLevel, format, v...)
	os.Exit(1)
}

func (l *Logger) outputln(logger *log.Logger, tracelevel int, v ...interface{}) {
	s := fmt.Sprintln(v...)
	if tracelevel > 0 {
		s += l.getTraceInfo(tracelevel)
	}
	logger.Output(3, s)
}

func (l *Logger) outputf(logger *log.Logger, tracelevel int, format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	if tracelevel > 0 {
		s += l.getTraceInfo(tracelevel)
	}
	logger.Output(3, s)
}

func (l *Logger) getTraceInfo(level int) string {
	t := ""
	for i := 0; i < level; i++ {
		_, file, line, ok := runtime.Caller(3 + i)
		if !ok {
			break
		}
		t += fmt.Sprintln("in", file, line)
	}
	return t
}
