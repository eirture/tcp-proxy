package log

import (
	"log"
	"os"
)

var (
	errLogger = log.New(os.Stderr, "", log.LstdFlags)
	outLogger = log.New(os.Stdout, "", log.LstdFlags)

	level = INFO
)

func SetLevel(l Level) {
	level = l
}

func Error(v ...interface{}) {
	errLogger.Print(v...)
}

func Errorln(v ...interface{}) {
	errLogger.Println(v...)
}

func Errorf(format string, v ...interface{}) {
	errLogger.Printf(format, v...)
}

func Warning(v ...interface{}) {
	outLogger.Print(v...)
}

func Warningln(v ...interface{}) {
	outLogger.Println(v...)
}

func Warningf(format string, v ...interface{}) {
	outLogger.Printf(format, v...)
}

func Info(v ...interface{}) {
	if level >= INFO {
		outLogger.Print(v...)
	}
}

func Infoln(v ...interface{}) {
	if level >= INFO {
		outLogger.Println(v...)
	}
}

func Infof(format string, v ...interface{}) {
	if level >= INFO {
		outLogger.Printf(format, v...)
	}
}

func Debug(v ...interface{}) {
	if level >= DEBUG {
		outLogger.Print(v...)
	}
}

func Debugln(v ...interface{}) {
	if level >= DEBUG {
		outLogger.Println(v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if level >= DEBUG {
		outLogger.Printf(format, v...)
	}
}
