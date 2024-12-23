package logging

import (
	"fmt"

	"github.com/mortedecai/resweave"
)

type LogFactory struct {
	resweave.LogHolder
}

func (b LogFactory) NewInfo(funcName string, status string) *logBuilder {
	return newLogBuilder(b, logTypeInfo, funcName).
		WithStatus(status)
}

func (b LogFactory) NewError(funcName string, err error) *logBuilder {
	return newLogBuilder(b, logTypeError, funcName).
		WithStatus("error").
		WithError(err)
}

func (b LogFactory) NewErrorMessage(funcName string, err error, msg string) *logBuilder {
	return newLogBuilder(b, logTypeError, funcName).
		WithStatus("error").
		WithErrorMessage(err, msg)
}

func (b LogFactory) NewDebug(funcName string) *logBuilder {
	return newLogBuilder(b, logTypeDebug, funcName)
}

type LogBuilder interface {
	WithStatus(status string) LogBuilder
	WithResource(name resweave.ResourceName) LogBuilder
	WithError(err error) LogBuilder
	WithErrorMessage(err error, msg string) LogBuilder
	With(name string, value interface{}) LogBuilder
	Log()
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

func (b *logBuilder) WithStatus(status string) *logBuilder {
	return b.With("Status", status)
}

func (b *logBuilder) WithResource(name resweave.ResourceName) *logBuilder {
	return b.With("Resource", name)
}

func (b *logBuilder) WithError(err error) *logBuilder {
	return b.With("Error", err)
}

func (b *logBuilder) WithErrorMessage(err error, msg string) *logBuilder {
	return b.WithError(fmt.Errorf("%s: %w", msg, err))
}

func (b *logBuilder) With(name string, value interface{}) *logBuilder {
	b.params = append(b.params, name, value)
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
		b.Errorw(b.funcName, append([]interface{}{"LogFault", "unsupported log type"}, b.params...)...)
	}
}
