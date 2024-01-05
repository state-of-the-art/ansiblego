package log

import (
	"fmt"
	"log"
	"os"
)

const (
	ERROR uint8 = 4
	WARN        = 8
	INFO        = 32
	DEBUG       = 64
	TRACE       = 128
)

var (
	UseTimestamp bool  = true
	Verbosity    uint8 = INFO

	TraceLogger *log.Logger
	DebugLogger *log.Logger
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	// Init default loggers
	InitLoggers(os.Stderr)
}

func SetVerbosity(level string) error {
	switch level {
	case "trace":
		Verbosity = TRACE
	case "debug":
		Verbosity = DEBUG
	case "info":
		Verbosity = INFO
	case "warn":
		Verbosity = WARN
	case "error":
		Verbosity = ERROR
	default:
		return fmt.Errorf("Unable to parse verbosity level: %s", level)
	}

	return nil
}

func InitLoggers(output *os.File) error {
	flags := log.Lmsgprefix

	// Showing short file for debug verbosity
	if Verbosity >= DEBUG {
		flags |= log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	} else if UseTimestamp {
		flags |= log.Ldate | log.Ltime
	}

	TraceLogger = log.New(output, "TRACE:\t", flags)
	DebugLogger = log.New(output, "DEBUG:\t", flags)
	InfoLogger = log.New(output, "INFO :\t", flags)
	WarnLogger = log.New(output, "WARN :\t", flags)
	ErrorLogger = log.New(output, "ERROR:\t", flags)

	return nil
}

func GetInfoLogger() *log.Logger {
	return InfoLogger
}

func Trace(v ...any) {
	if Verbosity >= TRACE {
		TraceLogger.Output(2, fmt.Sprintln(v...))
	}
}

func Tracef(format string, v ...any) {
	if Verbosity >= TRACE {
		TraceLogger.Output(2, fmt.Sprintf(format+"\n", v...))
	}
}

func Debug(v ...any) {
	if Verbosity >= DEBUG {
		DebugLogger.Output(2, fmt.Sprintln(v...))
	}
}

func Debugf(format string, v ...any) {
	if Verbosity >= DEBUG {
		DebugLogger.Output(2, fmt.Sprintf(format+"\n", v...))
	}
}

func Info(v ...any) {
	if Verbosity >= INFO {
		InfoLogger.Output(2, fmt.Sprintln(v...))
	}
}

func Infof(format string, v ...any) {
	if Verbosity >= INFO {
		InfoLogger.Output(2, fmt.Sprintf(format+"\n", v...))
	}
}

func Warn(v ...any) error {
	msg := fmt.Sprintln(v...)
	if Verbosity >= WARN {
		WarnLogger.Output(2, msg)
	}
	return fmt.Errorf("%s", msg)
}

func Warnf(format string, v ...any) error {
	if Verbosity >= WARN {
		WarnLogger.Output(2, fmt.Sprintf(format+"\n", v...))
	}
	return fmt.Errorf(format, v...)
}

func Error(v ...any) error {
	msg := fmt.Sprintln(v...)
	if Verbosity >= ERROR {
		ErrorLogger.Output(2, msg)
	}
	return fmt.Errorf("%s", msg)
}

func Errorf(format string, v ...any) error {
	if Verbosity >= ERROR {
		ErrorLogger.Output(2, fmt.Sprintf(format+"\n", v...))
	}
	return fmt.Errorf(format, v...)
}
