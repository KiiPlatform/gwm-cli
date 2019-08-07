package main

import (
	"os"

	kii "github.com/KiiPlatform/kii_go"
	"github.com/sirupsen/logrus"
)

// Logger implementation of kii.KiiLogger
type Logger struct {
}

var (
	// force to verify whether Logger implement all the method of KiiLogger interface.
	_      kii.KiiLogger = (*Logger)(nil)
	stdLog *logrus.Logger
)

func init() {
	stdLog = logrus.New()
	stdLog.Out = os.Stdout
}

// Debug writes debug message to log.
func (l *Logger) Debug(message string) {
	stdLog.Debug(message)
}

// Debugf formats debug message according to a format specifier and write it to log.
func (l *Logger) Debugf(format string, v ...interface{}) {
	stdLog.Debug(format, v)
}

// Info writes info message to log.
func (l *Logger) Info(message string) {
	stdLog.Info(message)
}

// Infof formats info message according to a format specifier and write it to log.
func (l *Logger) Infof(format string, v ...interface{}) {
	stdLog.Info(format, v)
}

// Warn writes warn message to log.
func (l *Logger) Warn(message string) {
	stdLog.Warn(message)
}

// Warnf formats warn message according to a format specifier and write it to log.
func (l *Logger) Warnf(format string, v ...interface{}) {
	stdLog.Warnf(format, v)
}

// Error writes error message to log.
func (l *Logger) Error(message string) {
	stdLog.Error(message)
}

// Errorf formats error message according to a format specifier and write it to log.
func (l *Logger) Errorf(format string, v ...interface{}) {
	stdLog.Errorf(format, v)
}
