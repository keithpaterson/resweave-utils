package logging

import (
	"go.uber.org/zap"
)

var (
	rootLogger *zap.SugaredLogger
)

// RootLogger instantiates a root-level logger; usually this is invoked in main() to establish the
// logging subsystem
func RootLogger() (*zap.SugaredLogger, error) {
	if rootLogger != nil {
		return rootLogger, nil
	}

	var logger *zap.Logger
	var err error
	if logger, err = zap.NewDevelopment(); err != nil {
		return nil, err
	}
	rootLogger = logger.Sugar()
	return rootLogger, nil
}

// NamedLogger instantiates a logger which will include 'name' in all the logs.
func NamedLogger(name string) (*zap.SugaredLogger, error) {
	root, err := RootLogger()
	if err != nil {
		return nil, err
	}
	return root.Named(name), nil
}
