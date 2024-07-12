package logger

import (
	"log"
	"os"
)

var (
	infoLogger    = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warningLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger   = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	fatalLogger   = log.New(os.Stderr, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func Info(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

// Warning logs a warning message using Printf for formatting
func Warning(format string, v ...interface{}) {
	warningLogger.Printf(format, v...)
}

// Error logs an error message using Printf for formatting
func Error(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}

// Fatal logs a fatal message using Printf for formatting and then exits the program
func Fatal(format string, v ...interface{}) {
	fatalLogger.Fatalf(format, v...)
}
