package logger

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func NewLogger(logFile string) (*Logger, error) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	multiInfo := io.MultiWriter(os.Stdout, file)
	multiError := io.MultiWriter(os.Stderr, file)

	return &Logger{
		infoLog:  log.New(multiInfo, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(multiError, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}

func (l *Logger) Info(v ...interface{}) {
	l.infoLog.Println(v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.errorLog.Println(v...)
}

func (l *Logger) InfoWithRequestID(requestID string, v ...interface{}) {
	l.infoLog.Println(append([]interface{}{"requestID:", requestID}, v...)...)
}

func (l *Logger) ErrorWithRequestID(requestID string, v ...interface{}) {
	l.errorLog.Println(append([]interface{}{"requestID:", requestID}, v...)...)
}
