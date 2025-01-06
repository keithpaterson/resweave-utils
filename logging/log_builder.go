package logging

import (
	"fmt"

	"github.com/mortedecai/resweave"
)

type LogKey string
type LogValue interface{}

// well-known logging keys
const (
	LogKeyStatus   = "status"
	LogKeyError    = "error"
	LogKeyResource = "resource"

	logKeyLogFault = "logFault"
)

// well-known logging values (mainly status values)
const (
	LogStatusStarted   = "started"
	LogStatusCompleted = "completed"
	LogStatusError     = "error"
)

type LogFactory struct {
	resweave.LogHolder
}

func (b LogFactory) NewInfo(funcName string, status LogValue) *logBuilder {
	return newLogBuilder(b, logTypeInfo, funcName).
		WithStatus(status)
}

func (b LogFactory) NewError(funcName string, err error) *logBuilder {
	return newLogBuilder(b, logTypeError, funcName).
		WithStatus(LogStatusError).
		WithError(err)
}

func (b LogFactory) NewErrorMessage(funcName string, err error, msg string) *logBuilder {
	return newLogBuilder(b, logTypeError, funcName).
		WithStatus(LogStatusError).
		WithErrorMessage(err, msg)
}

func (b LogFactory) NewDebug(funcName string) *logBuilder {
	return newLogBuilder(b, logTypeDebug, funcName)
}

type logType int

const (
	logTypeInfo logType = iota
	logTypeError
	logTypeDebug
)

type logBuilder struct {
	resweave.LogHolder

	logType  logType
	funcName string
	params   []interface{}
}

func newLogBuilder(logHolder resweave.LogHolder, logType logType, funcName string) *logBuilder {
	return &logBuilder{LogHolder: logHolder, logType: logType, funcName: funcName}
}

func (b *logBuilder) WithStatus(status LogValue) *logBuilder {
	return b.With(LogKeyStatus, status)
}

func (b *logBuilder) WithResource(name resweave.ResourceName) *logBuilder {
	return b.With(LogKeyResource, name)
}

func (b *logBuilder) WithError(err error) *logBuilder {
	return b.With(LogKeyError, err)
}

func (b *logBuilder) WithErrorMessage(err error, msg string) *logBuilder {
	return b.WithError(fmt.Errorf("%s: %w", msg, err))
}

func (b *logBuilder) With(name LogKey, value LogValue) *logBuilder {
	b.params = append(b.params, string(name), value)
	return b
}

func (b *logBuilder) Log() {
	switch b.logType {
	case logTypeInfo:
		b.Infow(b.funcName, b.params...)
	case logTypeError:
		b.Errorw(b.funcName, b.params...)
	case logTypeDebug:
		b.Debugw(b.funcName, b.params...)
	default:
		b.Errorw(b.funcName, append([]interface{}{logKeyLogFault, "unsupported log type"}, b.params...)...)
	}
}
