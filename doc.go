package site

//go:generate mockgen -destination=../mocks/http_mocks.go -package=mocks net/http ResponseWriter
//go:generate mockgen -destination=../mocks/io_reader_mocks.go -package=mocks io ReadCloser
