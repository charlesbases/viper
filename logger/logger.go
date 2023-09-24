package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
)

const logfile = "log.txt"

var log io.WriteCloser
var lock sync.Mutex

func init() {
	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	log = file
}

// Error .
func Error(v ...interface{}) {
	if len(v) != 0 && v[0] != nil {
		lock.Lock()
		log.Write([]byte("Err: " + fmt.Sprint(v...)))
		lock.Unlock()
	}
}

// Errorf .
func Errorf(format string, v ...interface{}) {
	if len(format) != 0 {
		lock.Lock()
		log.Write([]byte(fmt.Sprintf("Err: "+format, v...)))
		lock.Unlock()
	}
}

// Close .
func Close() {
	if log != nil {
		log.Close()
	}
}
