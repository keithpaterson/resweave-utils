// Package resweave-utils provides utility functions and classes that are designed to
// be used in conjunction with the [mortedecai/resweave] package.
package resweave_utils

//go:generate mockgen -destination=./mocks/http_mocks.go -package=mocks net/http ResponseWriter
//go:generate mockgen -destination=./mocks/io_reader_mocks.go -package=mocks io ReadCloser
//go:generate mockgen -destination=./mocks/resweave_mocks.go -package=mocks github.com/mortedecai/resweave Server
