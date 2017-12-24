package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	reset  = "\033[0m"
	red    = "\033[31;1m"
	green  = "\033[32;1m"
	yellow = "\033[33;1m"
)

type Level int

const ENV_CART_LOG_LEVEL = "CART_LOG_LEVEL"

const (
	Info Level = 1 + iota
	Warn
	Error
)

var logLevel = Info
var Logger = &logger{}

type logger struct {
	mu    sync.Mutex
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func init() {
	l := os.Getenv(ENV_CART_LOG_LEVEL)
	if len(l) == 0 {
		SetLevel(Info)
	} else {
		var level Level
		switch strings.ToLower(l) {
		case "info":
			level = Info
			break
		case "warn":
			level = Warn
			break
		case "error":
			level = Error
			break
		}
		SetLevel(level)
	}
	Logger.info = log.New(os.Stdout, "[INFO]  ", log.LstdFlags)
	Logger.warn = log.New(os.Stderr, "[WARN]  ", log.LstdFlags)
	Logger.error = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)
	log.SetOutput(os.Stdout)
}

func SetLevel(level Level) {
	logLevel = level
}

func SetFileOutput(dir string) {
	writer := DefaultFileWriter(dir)
	SetOutput(writer)
}

func SetOutput(writer io.Writer) {
	Logger.info.SetOutput(writer)
	Logger.warn.SetOutput(writer)
	Logger.error.SetOutput(writer)
	log.SetOutput(writer)
	if writer == os.Stdout || writer == os.Stderr {
		Logger.info.SetPrefix(green + "[INFO]  " + reset)
		Logger.warn.SetPrefix(yellow + "[WARN]  " + reset)
		Logger.error.SetPrefix(red + "[ERROR] " + reset)
		Logger.warn.SetFlags(log.LstdFlags)
		Logger.info.SetFlags(log.LstdFlags)
		Logger.error.SetFlags(log.LstdFlags)
		log.SetFlags(log.LstdFlags)
	} else {
		Logger.info.SetPrefix("[INFO]  ")
		Logger.warn.SetPrefix("[WARN]  ")
		Logger.error.SetPrefix("[ERROR] ")
		Logger.info.SetFlags(log.Ltime)
		Logger.warn.SetFlags(log.Ltime)
		Logger.error.SetFlags(log.Ltime)
	}
}

func (l *logger) Info(format string, v ...interface{}) {
	if logLevel <= Info {
		l.OutPut(l.info, fmt.Sprintf(format, v...))
	}
}
func (l *logger) Warn(format string, v ...interface{}) {
	if logLevel <= Warn {
		l.OutPut(l.warn, fmt.Sprintf(format, v...))
	}
}
func (this *logger) Error(format string, v ...interface{}) {
	if logLevel <= Error {
		this.OutPut(this.error, fmt.Sprintf(format, v...))
	}
}

func (this *logger) OutPut(l *log.Logger, s string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	l.Println(s)
}
