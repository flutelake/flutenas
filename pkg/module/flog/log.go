package flog

import (
	"fmt"
	"log"
)

type flog struct {
	log.Logger

	// log level
	// 0 -> error
	// 1 - 9 -> info
	// 10 - 99 -> warning
	// 100 - x -> debug
	level int
}

// globel logger instance
var logger *flog

func NewLogger(level int) {
	logger = &flog{
		level: level,
	}
}

func (fl *flog) logf(level int, str string, v ...any) {
	if level >= fl.level {
		return
	}
	flag := "ERROR"
	if 0 < level && level <= 9 {
		flag = "INFO"
	} else if 10 <= level && level <= 99 {
		flag = "WARN"
	} else if level >= 100 {
		flag = "DEBUG"
	}
	log.Printf(fmt.Sprintf("[%s][%d] ", flag, level)+str, v...)
}

func Errorf(str string, v ...any) { logger.logf(0, str, v...) }

func Infof(str string, v ...any) { logger.logf(5, str, v...) }

func Warnf(str string, v ...any) { logger.logf(50, str, v...) }

func Debugf(str string, v ...any) { logger.logf(500, str, v...) }

func Fatal(v ...any) { logger.Fatal(v...) }

func Fatalf(str string, v ...any) { logger.Fatalf(str, v...) }
