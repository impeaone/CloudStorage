package logger

import (
	consts "CloudStorageProject-FileServer/pkg/Constants"
	"CloudStorageProject-FileServer/pkg/Error"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Log struct {
	Level string // Level - уровень логгирования. NONE: нету логов, INFO, WARNING, ERROR
}

func NewLog(level string) *Log {
	choseLevel := level
	if level != "INFO" && level != "WARNING" && level != "ERROR" {
		choseLevel = "INFO"
	}
	return &Log{
		Level: choseLevel,
	}
}

func (logs *Log) Info(message string, place string) {
	if logs.Level != "INFO" {
		return
	}
	log.Println("\nLevel: Info" + "\nMessage: " + message + "\nPlace: " + place + "\n")
	go WriteLogsToFile(time.Now().Format("02.01.2006 15:04:05") +
		"\nLevel: Info" + "\nMessage: " + message + "\nPlace: " + place)
}

func (logs *Log) Warning(message string, place string) {
	if logs.Level == "ERROR" {
		return
	}
	log.Println("\nLevel: Warning" + "\nMessage: " + message + "\nPlace: " + place + "\n")
	go WriteLogsToFile(time.Now().Format("02.01.2006 15:04:05") +
		"\nLevel: Warning" + "\nMessage: " + message + "\nPlace: " + place)
}

func (logs *Log) Error(message string, place string) {
	log.Println("\nLevel: Error" + "\nMessage: " + message + "\nPlace: " + place + "\n")
	go WriteLogsToFile(time.Now().Format("02.01.2006 15:04:05") +
		"\nLevel: Error" + "\nMessage: " + message + "\nPlace: " + place)
}

func GetPlace() string {
	_, file, line, _ := runtime.Caller(1)
	split := strings.Split(file, "/")
	StartFile := split[len(split)-1]
	place := StartFile + ":" + strconv.Itoa(line)
	return place
}

var FileMTX sync.Mutex

func WriteLogsToFile(LogText string) {
	var logPath string
	if runtime.GOOS == "windows" {
		logPath = consts.LoggerPathWindows
	} else {
		logPath = consts.LoggerPathLinux
	}
	FileMTX.Lock()
	defer FileMTX.Unlock()
	file, err := os.OpenFile(
		logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("\nLevel: Error" + "\nMessage: " + Error.LogFileDoesNotOpen + ": " + err.Error() + "\nPlace: " +
			GetPlace() + "\n")
	}
	_, err = file.WriteString(LogText + "\n\n")
	if err != nil {
		log.Println("\nLevel: Error" + "\nMessage: " + Error.LogFileDoesNotWrite + ": " + err.Error() + "\nPlace: " +
			GetPlace() + "\n")
	}
	file.Close()
}
