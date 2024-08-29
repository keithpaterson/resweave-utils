package logging

import (
	"go.uber.org/zap"
)

var (
	rootLogger *zap.SugaredLogger
)

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

func NamedLogger(name string) (*zap.SugaredLogger, error) {
	root, err := RootLogger()
	if err != nil {
		return nil, err
	}
	return root.Named(name), nil
}
