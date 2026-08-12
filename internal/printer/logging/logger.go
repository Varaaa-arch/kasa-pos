package logging

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func New() *Logger {
	return &Logger{
		Logger: log.New(
			os.Stdout,
			"[PRINT-AGENT] ",
			log.Ldate|log.Ltime|log.Lmicroseconds,
		),
	}
}
