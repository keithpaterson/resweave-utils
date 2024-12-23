package logging

import (
	"errors"

	"github.com/mortedecai/resweave"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

var _ = Describe("LogBuilder", func() {
	var (
		observed zapcore.Core
		logs     *observer.ObservedLogs
		logger   *zap.SugaredLogger
		holder   resweave.LogHolder
	)

	Context("Log Factory", func() {
		BeforeEach(func() {
			observed, logs = observer.New(zapcore.DebugLevel)
			logger = zap.New(observed).Sugar()
			holder = resweave.NewLogholder("test", nil)
		})

		It("should be able to log info messages", func() {
			// Arrange
			factory := LogFactory{LogHolder: holder}
			factory.SetLogger(logger, false)

			// Act
			factory.NewInfo("logfunc", "was-logged").Log()

			Expect(logs.All()).To(HaveLen(1))
			entry := logs.All()[0]
			Expect(entry.Level).To(Equal(zapcore.InfoLevel))
			Expect(entry.Message).To(Equal("logfunc"))
			Expect(entry.ContextMap()).To(Equal(map[string]interface{}{"Status": "was-logged"}))
		})
		It("should be able to log simple error messages", func() {
			// Arrange
			factory := LogFactory{LogHolder: holder}
			factory.SetLogger(logger, false)

			// Act
			factory.NewError("logfunc", errors.New("log error")).Log()

			Expect(logs.All()).To(HaveLen(1))
			entry := logs.All()[0]
			Expect(entry.Level).To(Equal(zapcore.ErrorLevel))
			Expect(entry.Message).To(Equal("logfunc"))
			Expect(entry.ContextMap()).To(Equal(map[string]interface{}{"Status": "error", "Error": "log error"}))
		})
		It("should be able to log complex error messages", func() {
			// Arrange
			factory := LogFactory{LogHolder: holder}
			factory.SetLogger(logger, false)

			// Act
			factory.NewErrorMessage("logfunc", errors.New("log error"), "failed to foo").Log()

			Expect(logs.All()).To(HaveLen(1))
			entry := logs.All()[0]
			Expect(entry.Level).To(Equal(zapcore.ErrorLevel))
			Expect(entry.Message).To(Equal("logfunc"))
			Expect(entry.ContextMap()).To(Equal(map[string]interface{}{"Status": "error", "Error": "failed to foo: log error"}))
		})
		It("should be able to log debug messages", func() {
			// Arrange
			factory := LogFactory{LogHolder: holder}
			factory.SetLogger(logger, false)

			// Act
			factory.NewDebug("logfunc").With("key", "value").Log()

			Expect(logs.All()).To(HaveLen(1))
			entry := logs.All()[0]
			// can't seem to test that the log level is 'debug'; not sure how to enable that
			// outside of build-time.
			//Expect(entry.Level).To(Equal(zapcore.ErrorLevel))
			Expect(entry.Message).To(Equal("logfunc"))
			Expect(entry.ContextMap()).To(Equal(map[string]interface{}{"key": "value"}))
		})
	})

	Context("Log Builder", func() {
		BeforeEach(func() {
			observed, logs = observer.New(zapcore.DebugLevel)
			logger = zap.New(observed).Sugar()
			holder = resweave.NewLogholder("test", nil)
		})

		type testFunc func(b *logBuilder)

		DescribeTable("builder functions",
			func(logType logType, testFn testFunc, expectedContext map[string]interface{}) {
				// Arrange
				builder := newLogBuilder(holder, logType, "testfn")
				builder.SetLogger(logger, false)

				// Act
				testFn(builder)
				builder.Log()

				// Assert
				Expect(logs.All()).To(HaveLen(1))
				entry := logs.All()[0]
				Expect(entry.Message).To(Equal("testfn"))
				Expect(entry.ContextMap()).To(Equal(expectedContext))
			},
			Entry("info with status", logTypeInfo,
				func(b *logBuilder) { b.WithStatus("logged") },
				map[string]interface{}{"Status": "logged"}),
			Entry("info with status and custom value", logTypeInfo,
				func(b *logBuilder) { b.WithStatus("logged").With("key", "value") },
				map[string]interface{}{"Status": "logged", "key": "value"}),
			Entry("info with custom value", logTypeInfo,
				func(b *logBuilder) { b.With("key", "value") },
				map[string]interface{}{"key": "value"}),

			Entry("info with multiple custom value", logTypeInfo,
				func(b *logBuilder) { b.With("key", "value").With("another", "value") },
				map[string]interface{}{"key": "value", "another": "value"}),
			Entry("error with error", logTypeInfo,
				func(b *logBuilder) { b.WithError(errors.New("failed")) },
				map[string]interface{}{"Error": "failed"}),
			Entry("error with error and resource", logTypeInfo,
				func(b *logBuilder) { b.WithError(errors.New("failed")).WithResource("foo-resource") },
				map[string]interface{}{"Error": "failed", "Resource": "foo-resource"}),
			Entry("error with error message", logTypeInfo,
				func(b *logBuilder) { b.WithErrorMessage(errors.New("failed"), "tried") },
				map[string]interface{}{"Error": "tried: failed"}),
			Entry("error with error and custom value", logTypeInfo,
				func(b *logBuilder) { b.WithError(errors.New("failed")).With("attempt", 1) },
				map[string]interface{}{"Error": "failed", "attempt": int64(1)}),
			Entry("incorrect type logs an error", logType(-1),
				func(b *logBuilder) { /* no extra logs */ },
				map[string]interface{}{"LogFault": "unsupported log type"}),
			Entry("incorrect type with custom values logs an error", logType(-1),
				func(b *logBuilder) { b.WithStatus("pooched").WithResource("foo-res").With("not", "unusual") },
				map[string]interface{}{"LogFault": "unsupported log type", "Status": "pooched", "Resource": "foo-res", "not": "unusual"}),
		)
	})

})
