package utils

import (
	"log"
	"os"
)

var defaultLogger = dfdLogger{
	li: log.New(os.Stderr, White("INFO:"),
		log.LstdFlags|log.Lshortfile),
	ld: log.New(os.Stderr, Yellow("DEBUG:"),
		log.LstdFlags|log.Lshortfile),
}
