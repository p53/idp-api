package logging

import (
	"log"
	"os"
)

func GetLogger() *log.Logger {
	logger := log.New(os.Stderr, "logger: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	return logger
}
