package main

import (
	"fmt"
	"os"
)

var (
	logFile *os.File
)

func openLog() {
	var err error
	logFile, err = os.Create("log.txt")
	must(err)
}

func closeLog() {
	if logFile == nil {
		return
	}
	logFile.Close()
	logFile = nil
}

// TODO: should take additional format and args for optional message
func logError(err error) {
	if err != nil {
		return
	}
	lg("%s", err.Error())
}

func lg(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if logFile != nil {
		fmt.Fprint(logFile, s)
	}
	fmt.Print(s)
}

// TODO: have just one
func logVerbose(format string, args ...interface{}) {
	verbose(format, args...)
}

func verbose(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if logFile != nil {
		fmt.Fprint(logFile, s)
	}
}

var (
	doTempLog = false
)

func logTemp(format string, args ...interface{}) {
	if !doTempLog {
		return
	}
	lg(format, args...)
}
