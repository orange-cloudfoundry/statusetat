package common

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
)

type LogrusHclogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
	name   string
}

func (l *LogrusHclogger) GetLevel() hclog.Level {
	switch l.logger.GetLevel() {
	case logrus.InfoLevel:
		return hclog.Info
	case logrus.TraceLevel:
		return hclog.Trace
	case logrus.DebugLevel:
		return hclog.Debug
	case logrus.WarnLevel:
		return hclog.Warn
	case logrus.ErrorLevel:
		return hclog.Error
	}
	return hclog.NoLevel
}

func NewLogrusHclogger(logger *logrus.Logger) *LogrusHclogger {

	return &LogrusHclogger{logger: logger, entry: logrus.NewEntry(logger)}
}

//const (
//	// NoLevel is a special level used to indicate that no level has been
//	// set and allow for a default to be used.
//	NoLevel Level = 0
//
//	// Trace is the most verbose level. Intended to be used for the tracing
//	// of actions in code, such as function enters/exits, etc.
//	Trace Level = 1
//
//	// Debug information for programmer lowlevel analysis.
//	Debug Level = 2
//
//	// Info information about steady state operations.
//	Info Level = 3
//
//	// Warn information about rare but handled events.
//	Warn Level = 4
//
//	// Error information about unrecoverable events.
//	Error Level = 5
//)

func (l *LogrusHclogger) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.NoLevel:
		l.Info(msg, args...)
	case hclog.Trace:
		l.Trace(msg, args...)
	case hclog.Debug:
		l.Debug(msg, args...)
	case hclog.Info:
		l.Info(msg, args...)
	case hclog.Warn:
		l.Warn(msg, args...)
	case hclog.Error:
		l.Error(msg, args...)
	}
}

func (l *LogrusHclogger) Trace(msg string, args ...interface{}) {
	l.CreateEntry(args).Trace(msg)
}

func (l *LogrusHclogger) Debug(msg string, args ...interface{}) {
	l.CreateEntry(args).Debug(msg)
}

func (l *LogrusHclogger) Info(msg string, args ...interface{}) {
	l.CreateEntry(args).Info(msg)
}

func (l *LogrusHclogger) Warn(msg string, args ...interface{}) {
	l.CreateEntry(args).Warn(msg)
}

func (l *LogrusHclogger) Error(msg string, args ...interface{}) {
	l.CreateEntry(args).Error(msg)
}

func (l *LogrusHclogger) IsTrace() bool {
	return l.logger.GetLevel() == logrus.TraceLevel
}

func (l *LogrusHclogger) IsDebug() bool {
	return l.logger.GetLevel() >= logrus.DebugLevel
}

func (l *LogrusHclogger) IsInfo() bool {
	return l.logger.GetLevel() >= logrus.InfoLevel
}

func (l *LogrusHclogger) IsWarn() bool {
	return l.logger.GetLevel() >= logrus.WarnLevel
}

func (l *LogrusHclogger) IsError() bool {
	return l.logger.GetLevel() >= logrus.ErrorLevel
}

func (l *LogrusHclogger) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (l *LogrusHclogger) With(args ...interface{}) hclog.Logger {
	l.entry = l.CreateEntry(args)
	return l
}

func (l *LogrusHclogger) Name() string {
	return l.name
}

func (l *LogrusHclogger) Named(name string) hclog.Logger {
	l.entry = l.entry.WithField("name", name)
	l.name = name
	return l
}

func (l *LogrusHclogger) ResetNamed(name string) hclog.Logger {
	l.entry = logrus.NewEntry(l.logger)
	l.name = ""
	return l
}

func (l *LogrusHclogger) SetLevel(level hclog.Level) {
	switch level {
	case hclog.NoLevel:
		l.logger.SetLevel(logrus.InfoLevel)
	case hclog.Trace:
		l.logger.SetLevel(logrus.TraceLevel)
	case hclog.Debug:
		l.logger.SetLevel(logrus.DebugLevel)
	case hclog.Info:
		l.logger.SetLevel(logrus.InfoLevel)
	case hclog.Warn:
		l.logger.SetLevel(logrus.WarnLevel)
	case hclog.Error:
		l.logger.SetLevel(logrus.ErrorLevel)
	}
}

func (l *LogrusHclogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(logrus.StandardLogger().Out, "", log.LstdFlags)
}

func (l *LogrusHclogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return l.logger.Writer()
}
func (l *LogrusHclogger) CreateEntry(args []interface{}) *logrus.Entry {
	if len(args)%2 != 0 {
		args = append(args, "<unknown>")
	}

	fields := make(logrus.Fields)
	for i := 0; i < len(args); i += 2 {
		k, ok := args[i].(string)
		if !ok {
			continue
		}
		v := args[i+1]
		fields[k] = v
	}

	return l.entry.WithFields(fields)
}
