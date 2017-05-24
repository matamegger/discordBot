package logging

import (
	loglib "github.com/op/go-logging"
	"io"
)

type Logger struct {
	log *loglib.Logger
}

func NewLogger(name string, out io.Writer, err io.Writer) *Logger {
	logger, _ := loglib.GetLogger(name)
	format, _ := loglib.NewStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:04d}%{color:reset} %{message}`)
	logger.ExtraCalldepth=1

	outBackend := loglib.NewLogBackend(out,"",0)
	errBackend := loglib.NewLogBackend(err,"",0)

	outBackendFormatter := loglib.NewBackendFormatter(outBackend,format)
	errBackendFormatter := loglib.NewBackendFormatter(errBackend,format)

	errBackendLeveled := loglib.AddModuleLevel(errBackendFormatter)
	errBackendLeveled.SetLevel(loglib.ERROR,"")

	loglib.SetBackend(outBackendFormatter,errBackendLeveled)
	return &Logger{log: logger}
}

func (log *Logger) Debug(message string) {
	log.log.Debug(message)
}

func (log *Logger) Info(message string) {
	log.log.Info(message)
}

func (log *Logger) Notice(message string) {
	log.log.Notice(message)
}

func (log *Logger) Warning(message string) {
	log.log.Warning(message)
}

func (log *Logger) Error(message string) {
	log.log.Error(message)
}

func (log *Logger) Critical(message string) {
	log.log.Critical(message)
}

func (log *Logger) Debugf(message string, args ...interface{}) {
	log.log.Debugf(message, args)
}

func (log *Logger) Infof(message string, args ...interface{}) {
	log.log.Infof(message, args)
}

func (log *Logger) Noticef(message string, args ...interface{}) {
	log.log.Noticef(message, args)
}

func (log *Logger) Warningf(message string, args ...interface{}) {
	log.log.Warningf(message, args)
}

func (log *Logger) Errorf(message string, args ...interface{}) {
	log.log.Errorf(message, args)
}

func (log *Logger) Criticalf(message string, args ...interface{}) {
	log.log.Criticalf(message, args)
}
