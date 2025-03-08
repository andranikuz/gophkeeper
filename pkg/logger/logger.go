package logger

import (
	"log"
	"os"
)

// InfoLogger используется для вывода информационных сообщений.
var InfoLogger *log.Logger

// ErrorLogger используется для вывода сообщений об ошибках.
var ErrorLogger *log.Logger

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
