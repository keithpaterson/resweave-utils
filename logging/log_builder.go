package logging

import (
	"fmt"

	"github.com/mortedecai/resweave"
)

type LogKey string
type LogValue interface{}

// tag::loggingkeys[]

// well-known logging keys
const (
	LogKeyStatus   = "status"
	LogKeyError    = "error"
	LogKeyResource = "resource"

	logKeyLogFault = "logFault"
)

// end::loggingkeys[]
// tag::loggingvalues[]

// well-known logging values (mainly status values)
const (
	LogStatusStarted   = "started"
	LogStatusCompleted = "completed"
	LogStatusError     = "error"
)

// end::loggingvalues[]

// LogFactory is intended to be included in your struct, and is used to instantiate a
// LogBuilder instance when you need to log an info, error or debug message.
//
//	type MyClass struct {
//	    LogFactory
//	}
//
// A LogFactory includes a [resweave.LogHolder] so you can use the usual resweave logging
// functions if that is convenient.
//
// Note that [resource.EasyResource] will always contain a LogFactory, so your resource objects
// will always have access to these functions.
type LogFactory struct {
	resweave.LogHolder
}

// Instantiates a new Info log-builder.
// This presets the json struct with
//
//	{"status": status}.
//
// Example:
//
//	r.NewInfo("doSomething", LogStatusStarted).Log() = {..., "status": "started"}
func (b LogFactory) NewInfo(funcName string, status LogValue) *LogBuilder {
	return newLogBuilder(b, logTypeInfo, funcName).
		WithStatus(status)
}

// Instantiates a new Error log-builder.
// This presets the logged json struct with
//
//	{"status": status, "error": err.Error()}.
//
// Example:
//
//	r.NewError("doSomething", errors.New("foo")).Log() = {..., "status": "error", "error": "foo"}
func (b LogFactory) NewError(funcName string, err error) *LogBuilder {
	return newLogBuilder(b, logTypeError, funcName).
		WithStatus(LogStatusError).
		WithError(err)
}

// Instantiates a new Error log-builder.
// This presets the logged json struct with
//
//	{"status": status, "error": fmt.Sprintf("%s: %w", msg, err)}.
//
// Example:
//
//	r.NewErrorMessage("doSomething", errors.New("foo"), "failed").Log() = {..., "status": "error", "error": "failed: foo"}
func (b LogFactory) NewErrorMessage(funcName string, err error, msg string) *LogBuilder {
	return newLogBuilder(b, logTypeError, funcName).
		WithStatus(LogStatusError).
		WithErrorMessage(err, msg)
}

// Instantiates a new Debug log-builer.
func (b LogFactory) NewDebug(funcName string) *LogBuilder {
	return newLogBuilder(b, logTypeDebug, funcName)
}

type logType int

const (
	logTypeInfo logType = iota
	logTypeError
	logTypeDebug
)

type LogBuilder struct {
	resweave.LogHolder

	logType  logType
	funcName string
	params   []interface{}
}

func newLogBuilder(logHolder resweave.LogHolder, logType logType, funcName string) *LogBuilder {
	return &LogBuilder{LogHolder: logHolder, logType: logType, funcName: funcName}
}

// Adds a "status" field with the provided value
func (b *LogBuilder) WithStatus(status LogValue) *LogBuilder {
	return b.With(LogKeyStatus, status)
}

// Adds a "resource" field with the provided resource name.
//
// Note that when logging from within an EasyResource implementation, you don't need this
// because the resource struct will have initialized its logger with the resource name already.
func (b *LogBuilder) WithResource(name resweave.ResourceName) *LogBuilder {
	return b.With(LogKeyResource, name)
}

// Adds an "error" field with the error provided
func (b *LogBuilder) WithError(err error) *LogBuilder {
	return b.With(LogKeyError, err)
}

// Adds an "error" field with the the value "msg: err"
func (b *LogBuilder) WithErrorMessage(err error, msg string) *LogBuilder {
	return b.WithError(fmt.Errorf("%s: %w", msg, err))
}

// Adds a custom field.
//
// You should ensure that the field name is a json-compatible key value.
func (b *LogBuilder) With(name LogKey, value LogValue) *LogBuilder {
	b.params = append(b.params, string(name), value)
	return b
}

// Composes the message from the field data provided and emits it to the logger.
func (b *LogBuilder) Log() {
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
